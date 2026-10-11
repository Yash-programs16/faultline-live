package telemetry

import (
	"context"
	"sync"
	"time"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"
)

type Telemetry struct {
	db           *database.DB
	mu           sync.RWMutex
	window       time.Duration
	samples      []models.MetricSample
	requestCounts map[string]*endpointMetrics
	incidentMu   sync.Mutex
	lastIncident map[string]time.Time
	cooldown     time.Duration
}

type endpointMetrics struct {
	count      int64
	errors     int64
	latencies  []float64
	dbLatencies []float64
	lastReset  time.Time
}

func NewTelemetry(db *database.DB, window, cooldown time.Duration) *Telemetry {
	t := &Telemetry{
		db:            db,
		window:        window,
		samples:       make([]models.MetricSample, 0, 1000),
		requestCounts: make(map[string]*endpointMetrics),
		lastIncident:  make(map[string]time.Time),
		cooldown:      cooldown,
	}
	go t.aggregator()
	go t.persister()
	return t
}

func (t *Telemetry) RecordRequest(endpoint, method string, status int, duration time.Duration) {
	key := method + " " + endpoint
	now := time.Now()

	t.mu.Lock()
	em, ok := t.requestCounts[key]
	if !ok {
		em = &endpointMetrics{lastReset: now}
		t.requestCounts[key] = em
	}
	em.count++
	if status >= 500 {
		em.errors++
	}
	em.latencies = append(em.latencies, float64(duration.Microseconds()))
	if len(em.latencies) > 1000 {
		em.latencies = em.latencies[len(em.latencies)-1000:]
	}
	t.mu.Unlock()
}

func (t *Telemetry) RecordDBLatency(duration time.Duration) {
	t.mu.Lock()
	for _, em := range t.requestCounts {
		em.dbLatencies = append(em.dbLatencies, float64(duration.Microseconds()))
		if len(em.dbLatencies) > 1000 {
			em.dbLatencies = em.dbLatencies[len(em.dbLatencies)-1000:]
		}
	}
	t.mu.Unlock()
}

func (t *Telemetry) GetMetrics() models.MetricsResponse {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalRequests, totalErrors int64
	allLatencies := make([]float64, 0)
	allDBLatencies := make([]float64, 0)

	for _, em := range t.requestCounts {
		totalRequests += em.count
		totalErrors += em.errors
		allLatencies = append(allLatencies, em.latencies...)
		allDBLatencies = append(allDBLatencies, em.dbLatencies...)
	}

	var p50, p95, p99, dbP50, dbP95 float64
	if len(allLatencies) > 0 {
		p50 = percentile(allLatencies, 50)
		p95 = percentile(allLatencies, 95)
		p99 = percentile(allLatencies, 99)
	}
	if len(allDBLatencies) > 0 {
		dbP50 = percentile(allDBLatencies, 50)
		dbP95 = percentile(allDBLatencies, 95)
	}

	windowSec := t.window.Seconds()
	requestRate := float64(totalRequests) / windowSec
	errorRate := float64(0)
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests)
	}

	return models.MetricsResponse{
		RequestCount:    totalRequests,
		RequestRate:     requestRate,
		ErrorCount:      totalErrors,
		ErrorRate:       errorRate,
		LatencyP50:      p50 / 1000, // convert to ms
		LatencyP95:      p95 / 1000,
		LatencyP99:      p99 / 1000,
		DBLatencyP50:    dbP50 / 1000,
		DBLatencyP95:    dbP95 / 1000,
		ActiveFaults:    0, // will be set by faults package
		Timestamp:       time.Now().Format(time.RFC3339),
	}
}

func (t *Telemetry) GetEndpointMetrics() map[string]models.MetricsResponse {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]models.MetricsResponse)
	for key, em := range t.requestCounts {
		var p50, p95, p99, dbP50, dbP95 float64
		if len(em.latencies) > 0 {
			p50 = percentile(em.latencies, 50)
			p95 = percentile(em.latencies, 95)
			p99 = percentile(em.latencies, 99)
		}
		if len(em.dbLatencies) > 0 {
			dbP50 = percentile(em.dbLatencies, 50)
			dbP95 = percentile(em.dbLatencies, 95)
		}

		errorRate := float64(0)
		if em.count > 0 {
			errorRate = float64(em.errors) / float64(em.count)
		}

		result[key] = models.MetricsResponse{
			RequestCount:    em.count,
			ErrorCount:      em.errors,
			ErrorRate:       errorRate,
			LatencyP50:      p50 / 1000,
			LatencyP95:      p95 / 1000,
			LatencyP99:      p99 / 1000,
			DBLatencyP50:    dbP50 / 1000,
			DBLatencyP95:    dbP95 / 1000,
			Timestamp:       time.Now().Format(time.RFC3339),
		}
	}
	return result
}

func (t *Telemetry) aggregator() {
	ticker := time.NewTicker(t.window)
	defer ticker.Stop()

	for range ticker.C {
		t.mu.Lock()
		metrics := t.GetMetricsLocked()
		t.samples = append(t.samples, metrics)
		// Keep last 24 hours of samples (at 5min intervals = 288 samples)
		maxSamples := int(24 * time.Hour / t.window)
		if len(t.samples) > maxSamples {
			t.samples = t.samples[len(t.samples)-maxSamples:]
		}
		// Reset endpoint metrics
		t.requestCounts = make(map[string]*endpointMetrics)
		t.mu.Unlock()
	}
}

func (t *Telemetry) GetMetricsLocked() models.MetricSample {
	var totalRequests, totalErrors int64
	allLatencies := make([]float64, 0)
	allDBLatencies := make([]float64, 0)

	for _, em := range t.requestCounts {
		totalRequests += em.count
		totalErrors += em.errors
		allLatencies = append(allLatencies, em.latencies...)
		allDBLatencies = append(allDBLatencies, em.dbLatencies...)
	}

	var p50, p95, p99, dbP50, dbP95 float64
	if len(allLatencies) > 0 {
		p50 = percentile(allLatencies, 50)
		p95 = percentile(allLatencies, 95)
		p99 = percentile(allLatencies, 99)
	}
	if len(allDBLatencies) > 0 {
		dbP50 = percentile(allDBLatencies, 50)
		dbP95 = percentile(allDBLatencies, 95)
	}

	return models.MetricSample{
		Timestamp:       time.Now(),
		RequestCount:    totalRequests,
		ErrorCount:      totalErrors,
		LatencyP50:      p50 / 1000,
		LatencyP95:      p95 / 1000,
		LatencyP99:      p99 / 1000,
		DBLatencyP50:    dbP50 / 1000,
		DBLatencyP95:    dbP95 / 1000,
	}
}

func (t *Telemetry) GetHistory(window string) ([]models.MetricSample, error) {
	dur, err := time.ParseDuration(window)
	if err != nil {
		dur = time.Hour
	}
	cutoff := time.Now().Add(-dur)

	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]models.MetricSample, 0)
	for _, s := range t.samples {
		if s.Timestamp.After(cutoff) {
			result = append(result, s)
		}
	}
	return result, nil
}

func (t *Telemetry) persister() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		t.persistSample(ctx)
		cancel()
	}
}

func (t *Telemetry) persistSample(ctx context.Context) {
	t.mu.RLock()
	if len(t.samples) == 0 {
		t.mu.RUnlock()
		return
	}
	latest := t.samples[len(t.samples)-1]
	t.mu.RUnlock()

	_, err := t.db.Pool.Exec(ctx, `
		INSERT INTO metric_samples (timestamp, request_count, error_count, latency_p50, latency_p95, latency_p99, db_latency_p50, db_latency_p95)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (timestamp) DO UPDATE SET
			request_count = EXCLUDED.request_count,
			error_count = EXCLUDED.error_count,
			latency_p50 = EXCLUDED.latency_p50,
			latency_p95 = EXCLUDED.latency_p95,
			latency_p99 = EXCLUDED.latency_p99,
			db_latency_p50 = EXCLUDED.db_latency_p50,
			db_latency_p95 = EXCLUDED.db_latency_p95
	`, latest.Timestamp, latest.RequestCount, latest.ErrorCount, latest.LatencyP50, latest.LatencyP95, latest.LatencyP99, latest.DBLatencyP50, latest.DBLatencyP95)
	if err != nil {
		// Log error but don't crash
	}
}

func (t *Telemetry) ShouldFireIncident(key string) bool {
	t.incidentMu.Lock()
	defer t.incidentMu.Unlock()

	last, ok := t.lastIncident[key]
	if !ok || time.Since(last) > t.cooldown {
		t.lastIncident[key] = time.Now()
		return true
	}
	return false
}

func (t *Telemetry) MarkIncidentResolved(key string) {
	t.incidentMu.Lock()
	defer t.incidentMu.Unlock()
	delete(t.lastIncident, key)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	// Make a copy and sort
	vals := make([]float64, len(sorted))
	copy(vals, sorted)
	// Simple sort for small slices
	for i := 0; i < len(vals); i++ {
		for j := i + 1; j < len(vals); j++ {
			if vals[i] > vals[j] {
				vals[i], vals[j] = vals[j], vals[i]
			}
		}
	}
	idx := int(float64(len(vals)-1) * p / 100)
	if idx >= len(vals) {
		idx = len(vals) - 1
	}
	return vals[idx]
}