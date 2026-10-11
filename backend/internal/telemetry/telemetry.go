package telemetry

import (
	"context"
	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

type observation struct {
	at       time.Time
	endpoint string
	status   int
	us       float64
}
type dbObservation struct {
	at time.Time
	us float64
}
type Telemetry struct {
	db           *database.DB
	mu           sync.RWMutex
	window       time.Duration
	samples      []models.MetricSample
	requests     []observation
	queries      []dbObservation
	lastIncident map[string]time.Time
	cooldown     time.Duration
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func NewTelemetry(db *database.DB, window, cooldown time.Duration) *Telemetry {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Telemetry{db: db, window: window, cooldown: cooldown, lastIncident: map[string]time.Time{}, cancel: cancel, samples: []models.MetricSample{}}
	t.wg.Add(1)
	go t.run(ctx)
	return t
}
func (t *Telemetry) Stop() { t.cancel(); t.wg.Wait() }
func (t *Telemetry) RecordRequest(endpoint, method string, status int, duration time.Duration) {
	// Monitoring calls must not dilute the workload's error rate and latency.
	if strings.HasPrefix(endpoint, "/api/ops") || strings.HasPrefix(endpoint, "/api/metrics") || endpoint == "/api/health" || endpoint == "/api/ready" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests = append(t.requests, observation{time.Now(), method + " " + endpoint, status, float64(duration.Microseconds())})
	t.pruneLocked(time.Now())
	if len(t.requests) > 20000 {
		t.requests = t.requests[len(t.requests)-20000:]
	}
}
func (t *Telemetry) RecordDBLatency(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.queries = append(t.queries, dbObservation{time.Now(), float64(duration.Microseconds())})
	t.pruneLocked(time.Now())
	if len(t.queries) > 20000 {
		t.queries = t.queries[len(t.queries)-20000:]
	}
}
func (t *Telemetry) pruneLocked(now time.Time) {
	cutoff := now.Add(-t.window)
	i := 0
	for i < len(t.requests) && t.requests[i].at.Before(cutoff) {
		i++
	}
	t.requests = t.requests[i:]
	i = 0
	for i < len(t.queries) && t.queries[i].at.Before(cutoff) {
		i++
	}
	t.queries = t.queries[i:]
}
func (t *Telemetry) metricsLocked(endpoint string) models.MetricsResponse {
	now := time.Now()
	cutoff := now.Add(-t.window)
	m := models.MetricsResponse{Timestamp: now.Format(time.RFC3339)}
	var latencies, dbLatencies []float64
	for _, r := range t.requests {
		if r.at.Before(cutoff) || (endpoint != "" && r.endpoint != endpoint) {
			continue
		}
		m.RequestCount++
		if r.status >= 500 {
			m.ErrorCount++
		}
		latencies = append(latencies, r.us)
	}
	if endpoint == "" {
		for _, q := range t.queries {
			if !q.at.Before(cutoff) {
				dbLatencies = append(dbLatencies, q.us)
			}
		}
	}
	m.RequestRate = float64(m.RequestCount) / t.window.Seconds()
	if m.RequestCount > 0 {
		m.ErrorRate = float64(m.ErrorCount) / float64(m.RequestCount)
	}
	m.LatencyP50 = percentile(latencies, 50) / 1000
	m.LatencyP95 = percentile(latencies, 95) / 1000
	m.LatencyP99 = percentile(latencies, 99) / 1000
	m.DBLatencyP50 = percentile(dbLatencies, 50) / 1000
	m.DBLatencyP95 = percentile(dbLatencies, 95) / 1000
	m.DBObservationCount = int64(len(dbLatencies))
	return m
}
func (t *Telemetry) GetMetrics() models.MetricsResponse {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.metricsLocked("")
}
func (t *Telemetry) GetEndpointMetrics() map[string]models.MetricsResponse {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := map[string]models.MetricsResponse{}
	cutoff := time.Now().Add(-t.window)
	for _, r := range t.requests {
		if !r.at.Before(cutoff) {
			if _, ok := result[r.endpoint]; !ok {
				result[r.endpoint] = t.metricsLocked(r.endpoint)
			}
		}
	}
	return result
}
func (t *Telemetry) GetMetricsLocked() models.MetricSample {
	m := t.metricsLocked("")
	return models.MetricSample{Timestamp: time.Now(), RequestCount: m.RequestCount, ErrorCount: m.ErrorCount, LatencyP50: m.LatencyP50, LatencyP95: m.LatencyP95, LatencyP99: m.LatencyP99, DBLatencyP50: m.DBLatencyP50, DBLatencyP95: m.DBLatencyP95}
}
func (t *Telemetry) run(ctx context.Context) {
	defer t.wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	persist := time.NewTicker(5 * time.Minute)
	defer persist.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			t.pruneLocked(time.Now())
			sample := t.GetMetricsLocked()
			if sample.RequestCount > 0 {
				t.samples = append(t.samples, sample)
			}
			cutoff := time.Now().Add(-24 * time.Hour)
			i := 0
			for i < len(t.samples) && t.samples[i].Timestamp.Before(cutoff) {
				i++
			}
			t.samples = t.samples[i:]
			t.mu.Unlock()
		case <-persist.C:
			queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.persistSample(queryCtx)
			cancel()
		}
	}
}
func (t *Telemetry) GetHistory(window string) ([]models.MetricSample, error) {
	dur, err := time.ParseDuration(window)
	if err != nil || dur <= 0 || dur > 24*time.Hour {
		return nil, fmt.Errorf("window must be a positive duration up to 24h")
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := []models.MetricSample{}
	cutoff := time.Now().Add(-dur)
	for _, s := range t.samples {
		if s.Timestamp.After(cutoff) {
			result = append(result, s)
		}
	}
	return result, nil
}
func (t *Telemetry) persistSample(ctx context.Context) {
	if t.db == nil {
		return
	}
	t.mu.RLock()
	if len(t.samples) == 0 {
		t.mu.RUnlock()
		return
	}
	s := t.samples[len(t.samples)-1]
	t.mu.RUnlock()
	_, err := t.db.Pool.Exec(ctx, `INSERT INTO metric_samples (timestamp,request_count,error_count,latency_p50,latency_p95,latency_p99,db_latency_p50,db_latency_p95) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (timestamp) DO NOTHING`, s.Timestamp, s.RequestCount, s.ErrorCount, s.LatencyP50, s.LatencyP95, s.LatencyP99, s.DBLatencyP50, s.DBLatencyP95)
	if err != nil {
		slog.Warn("metric persistence failed")
	}
}
func (t *Telemetry) ShouldFireIncident(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	last, ok := t.lastIncident[key]
	if !ok || time.Since(last) > t.cooldown {
		t.lastIncident[key] = time.Now()
		return true
	}
	return false
}
func (t *Telemetry) MarkIncidentResolved(key string) {
	t.mu.Lock()
	delete(t.lastIncident, key)
	t.mu.Unlock()
}
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	vals := append([]float64(nil), values...)
	sort.Float64s(vals)
	return vals[int(float64(len(vals)-1)*p/100)]
}
