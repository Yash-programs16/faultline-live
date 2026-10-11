package incidents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/telemetry"
)

type IncidentDetector struct {
	db            *database.DB
	tel           *telemetry.Telemetry
	logger        *slog.Logger
	mu            sync.RWMutex
	incidents     map[string]*models.Incident
	baselines     map[string]models.MetricsResponse
	thresholds    Thresholds
	checkInterval time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
	recovery      map[string]int
}

type Thresholds struct {
	LatencyP95Ms       float64
	ErrorRate          float64
	DBLatencyP95Ms     float64
	RequestRateRPS     float64
	MinRequests        int64
	MinTrafficRequests int64
}

func DefaultThresholds() Thresholds {
	return Thresholds{
		LatencyP95Ms:       500,
		ErrorRate:          0.05,
		DBLatencyP95Ms:     200,
		RequestRateRPS:     0.05,
		MinRequests:        10,
		MinTrafficRequests: 20,
	}
}

func NewIncidentDetector(db *database.DB, tel *telemetry.Telemetry, logger *slog.Logger) *IncidentDetector {
	return &IncidentDetector{
		db:            db,
		tel:           tel,
		logger:        logger,
		incidents:     make(map[string]*models.Incident),
		recovery:      make(map[string]int),
		baselines:     make(map[string]models.MetricsResponse),
		thresholds:    DefaultThresholds(),
		checkInterval: 10 * time.Second,
		stopCh:        make(chan struct{}),
	}
}

func (d *IncidentDetector) Start(ctx context.Context) {
	d.loadHistory(ctx)
	d.loadBaselines(ctx)
	go d.run(ctx)
}

func (d *IncidentDetector) Stop() {
	d.stopOnce.Do(func() { close(d.stopCh) })
}

func (d *IncidentDetector) run(ctx context.Context) {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.check(ctx)
		}
	}
}

func (d *IncidentDetector) check(ctx context.Context) {
	metrics := d.tel.GetMetrics()
	endpointMetrics := d.tel.GetEndpointMetrics()

	// Evaluate this sample against the previously learned baseline before updating it.
	// Updating first would let the current spike contaminate its own baseline.
	d.checkGlobalMetrics(ctx, metrics)

	for endpoint, em := range endpointMetrics {
		d.checkEndpoint(ctx, endpoint, em)
	}

	d.checkRecovery(ctx)

	if d.noActiveIncidents() && len(endpointMetrics) > 0 {
		d.updateBaselines(endpointMetrics)
	}
}

func (d *IncidentDetector) checkGlobalMetrics(ctx context.Context, m models.MetricsResponse) {
	if m.RequestCount < d.thresholds.MinRequests {
		return
	}

	// High latency
	if m.LatencyP95 > d.thresholds.LatencyP95Ms {
		key := "global:high_latency"
		if d.tel.ShouldFireIncident(key) {
			d.fireIncident(ctx, models.Incident{
				Type:             "high_latency",
				Severity:         severityFromLatency(m.LatencyP95, d.thresholds.LatencyP95Ms),
				Title:            "Elevated API Latency",
				Description:      fmt.Sprintf("Global P95 latency is %.0fms (threshold: %.0fms)", m.LatencyP95, d.thresholds.LatencyP95Ms),
				AffectedEndpoint: "global",
				Observed:         map[string]interface{}{"latency_p95_ms": m.LatencyP95, "request_count": m.RequestCount},
				Baseline:         d.getBaseline("global"),
				Threshold:        map[string]interface{}{"latency_p95_ms": d.thresholds.LatencyP95Ms},
				Evidence:         []string{fmt.Sprintf("P95 latency %.0fms exceeds threshold %.0fms", m.LatencyP95, d.thresholds.LatencyP95Ms)},
			})
		}
	}

	// High error rate
	if m.ErrorRate > d.thresholds.ErrorRate {
		key := "global:high_error_rate"
		if d.tel.ShouldFireIncident(key) {
			d.fireIncident(ctx, models.Incident{
				Type:             "high_error_rate",
				Severity:         severityFromErrorRate(m.ErrorRate),
				Title:            "Elevated Error Rate",
				Description:      fmt.Sprintf("Global error rate is %.2f%% (threshold: %.2f%%)", m.ErrorRate*100, d.thresholds.ErrorRate*100),
				AffectedEndpoint: "global",
				Observed:         map[string]interface{}{"error_rate": m.ErrorRate, "error_count": m.ErrorCount, "request_count": m.RequestCount},
				Baseline:         d.getBaseline("global"),
				Threshold:        map[string]interface{}{"error_rate": d.thresholds.ErrorRate},
				Evidence:         []string{fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%", m.ErrorRate*100, d.thresholds.ErrorRate*100)},
			})
		}
	}

	// High DB latency
	if m.DBLatencyP95 > d.thresholds.DBLatencyP95Ms {
		key := "global:high_db_latency"
		if d.tel.ShouldFireIncident(key) {
			d.fireIncident(ctx, models.Incident{
				Type:             "high_db_latency",
				Severity:         severityFromLatency(m.DBLatencyP95, d.thresholds.DBLatencyP95Ms),
				Title:            "Elevated Database Latency",
				Description:      fmt.Sprintf("Database P95 latency is %.0fms (threshold: %.0fms)", m.DBLatencyP95, d.thresholds.DBLatencyP95Ms),
				AffectedEndpoint: "database",
				Observed:         map[string]interface{}{"db_latency_p95_ms": m.DBLatencyP95, "request_count": m.RequestCount},
				Baseline:         d.getBaseline("global"),
				Threshold:        map[string]interface{}{"db_latency_p95_ms": d.thresholds.DBLatencyP95Ms},
				Evidence:         []string{fmt.Sprintf("DB P95 latency %.0fms exceeds threshold %.0fms", m.DBLatencyP95, d.thresholds.DBLatencyP95Ms)},
			})
		}
	}
}

func (d *IncidentDetector) checkEndpoint(ctx context.Context, endpoint string, em models.MetricsResponse) {
	if em.RequestCount < d.thresholds.MinRequests {
		return
	}

	key := "endpoint:" + endpoint

	// High latency on endpoint
	if em.LatencyP95 > d.thresholds.LatencyP95Ms {
		incidentKey := key + ":high_latency"
		if d.tel.ShouldFireIncident(incidentKey) {
			d.fireIncident(ctx, models.Incident{
				Type:             "high_latency",
				Severity:         severityFromLatency(em.LatencyP95, d.thresholds.LatencyP95Ms),
				Title:            "Elevated Endpoint Latency",
				Description:      fmt.Sprintf("Endpoint %s P95 latency is %.0fms", endpoint, em.LatencyP95),
				AffectedEndpoint: endpoint,
				Observed:         map[string]interface{}{"latency_p95_ms": em.LatencyP95, "request_count": em.RequestCount},
				Baseline:         d.getBaseline(key),
				Threshold:        map[string]interface{}{"latency_p95_ms": d.thresholds.LatencyP95Ms},
				Evidence:         []string{fmt.Sprintf("Endpoint %s P95 latency %.0fms exceeds threshold", endpoint, em.LatencyP95)},
			})
		}
	}

	// High error rate on endpoint
	if em.ErrorRate > d.thresholds.ErrorRate {
		incidentKey := key + ":high_error_rate"
		if d.tel.ShouldFireIncident(incidentKey) {
			d.fireIncident(ctx, models.Incident{
				Type:             "high_error_rate",
				Severity:         severityFromErrorRate(em.ErrorRate),
				Title:            "Elevated Endpoint Error Rate",
				Description:      fmt.Sprintf("Endpoint %s error rate is %.2f%%", endpoint, em.ErrorRate*100),
				AffectedEndpoint: endpoint,
				Observed:         map[string]interface{}{"error_rate": em.ErrorRate, "error_count": em.ErrorCount, "request_count": em.RequestCount},
				Baseline:         d.getBaseline(key),
				Threshold:        map[string]interface{}{"error_rate": d.thresholds.ErrorRate},
				Evidence:         []string{fmt.Sprintf("Endpoint %s error rate %.2f%% exceeds threshold", endpoint, em.ErrorRate*100)},
			})
		}
	}

	// A bounded local traffic burst can be detected as unusual request volume.
	// This is a simple threshold rule, not a trained ML/DDoS classifier.
	if em.RequestCount >= d.thresholds.MinTrafficRequests && em.RequestRate >= d.thresholds.RequestRateRPS {
		incidentKey := key + ":traffic_spike"
		if d.tel.ShouldFireIncident(incidentKey) {
			severity := "medium"
			if em.RequestRate >= d.thresholds.RequestRateRPS*4 {
				severity = "high"
			}
			baseline := d.getBaseline(key)
			d.fireIncident(ctx, models.Incident{
				Type:             "traffic_spike",
				Severity:         severity,
				Title:            "Unusual Request Volume",
				Description:      fmt.Sprintf("Endpoint %s received %d requests in the current telemetry window (%.3f req/s)", endpoint, em.RequestCount, em.RequestRate),
				AffectedEndpoint: endpoint,
				Observed: map[string]interface{}{
					"request_count":    em.RequestCount,
					"request_rate_rps": em.RequestRate,
				},
				Baseline:  baseline,
				Threshold: map[string]interface{}{"min_request_count": d.thresholds.MinTrafficRequests, "request_rate_rps": d.thresholds.RequestRateRPS},
				Evidence: []string{
					fmt.Sprintf("Endpoint request rate %.3f req/s exceeded demo threshold %.3f req/s", em.RequestRate, d.thresholds.RequestRateRPS),
					fmt.Sprintf("Observed %d requests in the telemetry window", em.RequestCount),
				},
			})
		}
	}
}

func (d *IncidentDetector) checkRecovery(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, inc := range d.incidents {
		if inc.Status == "resolved" {
			continue
		}

		// Check if the condition that triggered this incident has recovered
		recovered := false
		switch inc.Type {
		case "high_latency":
			metrics := d.tel.GetMetrics()
			if inc.AffectedEndpoint == "global" || inc.AffectedEndpoint == "database" {
				if metrics.RequestCount >= d.thresholds.MinRequests && metrics.LatencyP95 < d.thresholds.LatencyP95Ms*0.8 {
					recovered = true
				}
			} else {
				endpointMetrics := d.tel.GetEndpointMetrics()
				if em, ok := endpointMetrics[inc.AffectedEndpoint]; ok {
					if em.RequestCount >= d.thresholds.MinRequests && em.LatencyP95 < d.thresholds.LatencyP95Ms*0.8 {
						recovered = true
					}
				}
			}
		case "high_error_rate":
			metrics := d.tel.GetMetrics()
			if inc.AffectedEndpoint == "global" {
				if metrics.RequestCount >= d.thresholds.MinRequests && metrics.ErrorRate < d.thresholds.ErrorRate*0.5 {
					recovered = true
				}
			} else {
				endpointMetrics := d.tel.GetEndpointMetrics()
				if em, ok := endpointMetrics[inc.AffectedEndpoint]; ok {
					if em.RequestCount >= d.thresholds.MinRequests && em.ErrorRate < d.thresholds.ErrorRate*0.5 {
						recovered = true
					}
				}
			}
		case "high_db_latency":
			metrics := d.tel.GetMetrics()
			if metrics.DBObservationCount >= d.thresholds.MinRequests && metrics.DBLatencyP95 < d.thresholds.DBLatencyP95Ms*0.8 {
				recovered = true
			}
		case "traffic_spike":
			endpointMetrics := d.tel.GetEndpointMetrics()
			if em, ok := endpointMetrics[inc.AffectedEndpoint]; ok && em.RequestCount >= d.thresholds.MinRequests && em.RequestRate < d.thresholds.RequestRateRPS*0.5 {
				recovered = true
			}
		}

		if !recovered {
			d.recovery[key] = 0
			continue
		}
		d.recovery[key]++
		if d.recovery[key] >= 2 {
			now := time.Now()
			inc.Status = "resolved"
			inc.ResolvedAt = &now
			for i := range inc.Actions {
				if inc.Actions[i].Status == "succeeded" && inc.Actions[i].CreatedAt.Before(now) {
					inc.Actions[i].RecoveryVerified = true
				}
			}
			cooldownKey := "endpoint:" + inc.AffectedEndpoint + ":" + inc.Type
			if inc.AffectedEndpoint == "global" || inc.AffectedEndpoint == "database" {
				cooldownKey = "global:" + inc.Type
			}
			d.tel.MarkIncidentResolved(cooldownKey)
			d.persistIncident(ctx, inc)
			d.logger.Info("incident resolved", "id", inc.ID, "type", inc.Type)
		}
	}
}

func (d *IncidentDetector) fireIncident(ctx context.Context, inc models.Incident) {
	d.mu.Lock()
	for _, existing := range d.incidents {
		if existing.Type == inc.Type && existing.AffectedEndpoint == inc.AffectedEndpoint && existing.Status == "firing" {
			existing.Observed = inc.Observed
			existing.Evidence = inc.Evidence
			existing.LastObservedAt = time.Now()
			snapshot := cloneIncident(existing)
			d.mu.Unlock()
			d.persistIncident(ctx, snapshot)
			return
		}
	}
	inc.ID = models.NewIncidentID()
	inc.Status = "firing"
	inc.StartedAt = time.Now()
	inc.CreatedAt = inc.StartedAt
	inc.LastObservedAt = inc.StartedAt
	d.incidents[inc.ID] = &inc
	snapshot := cloneIncident(&inc)
	d.mu.Unlock()
	d.persistIncident(ctx, snapshot)
	d.logger.Warn("incident fired", "id", inc.ID, "type", inc.Type, "endpoint", inc.AffectedEndpoint)
}

func (d *IncidentDetector) persistIncident(ctx context.Context, inc *models.Incident) {
	if d.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	observedJSON, _ := json.Marshal(inc.Observed)
	baselineJSON, _ := json.Marshal(inc.Baseline)
	thresholdJSON, _ := json.Marshal(inc.Threshold)
	evidenceJSON, _ := json.Marshal(inc.Evidence)

	var resolvedAt interface{}
	if inc.ResolvedAt != nil {
		resolvedAt = *inc.ResolvedAt
	}

	_, err := d.db.Pool.Exec(ctx, `
		INSERT INTO incidents (id, type, severity, title, description, affected_endpoint, observed, baseline, threshold, evidence, status, started_at, resolved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			resolved_at = EXCLUDED.resolved_at,
			observed = EXCLUDED.observed,
			baseline = EXCLUDED.baseline,
			threshold = EXCLUDED.threshold,
			evidence = EXCLUDED.evidence
	`, inc.ID, inc.Type, inc.Severity, inc.Title, inc.Description, inc.AffectedEndpoint,
		observedJSON, baselineJSON, thresholdJSON, evidenceJSON, inc.Status, inc.StartedAt, resolvedAt, inc.CreatedAt)
	if err != nil {
		d.logger.Error("persist incident", "error", err)
	}
}

func (d *IncidentDetector) loadBaselines(ctx context.Context) {
	if d.db == nil {
		return
	}
	rows, err := d.db.Pool.Query(ctx, `SELECT endpoint, metrics FROM baselines`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var endpoint string
		var metricsJSON []byte
		if err := rows.Scan(&endpoint, &metricsJSON); err != nil {
			continue
		}
		var m models.MetricsResponse
		if json.Unmarshal(metricsJSON, &m) == nil {
			d.mu.Lock()
			d.baselines[endpoint] = m
			d.mu.Unlock()
		}
	}
}

func (d *IncidentDetector) updateBaselines(endpointMetrics map[string]models.MetricsResponse) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update global baseline
	global := d.tel.GetMetrics()
	d.baselines["global"] = global

	// Update endpoint baselines
	for endpoint, em := range endpointMetrics {
		d.baselines["endpoint:"+endpoint] = em
	}

	if d.db == nil {
		return
	}
	// Persist to database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for endpoint, m := range d.baselines {
		metricsJSON, _ := json.Marshal(m)
		_, _ = d.db.Pool.Exec(ctx, `
			INSERT INTO baselines (endpoint, metrics, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (endpoint) DO UPDATE SET metrics = EXCLUDED.metrics, updated_at = NOW()
		`, endpoint, metricsJSON)
	}
}

func (d *IncidentDetector) getBaseline(key string) map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if m, ok := d.baselines[key]; ok {
		return map[string]interface{}{
			"latency_p95_ms":   m.LatencyP95,
			"error_rate":       m.ErrorRate,
			"request_count":    m.RequestCount,
			"request_rate_rps": m.RequestRate,
			"db_latency_p95":   m.DBLatencyP95,
		}
	}
	return map[string]interface{}{}
}

func (d *IncidentDetector) noActiveIncidents() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, inc := range d.incidents {
		if inc.Status == "firing" {
			return false
		}
	}
	return true
}

func (d *IncidentDetector) GetIncidents() []models.Incident {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]models.Incident, 0, len(d.incidents))
	for _, inc := range d.incidents {
		result = append(result, *cloneIncident(inc))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.After(result[j].StartedAt) })
	return result
}

func (d *IncidentDetector) GetIncident(id string) *models.Incident {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, inc := range d.incidents {
		if inc.ID == id {
			return cloneIncident(inc)
		}
	}
	return nil
}

func severityFromLatency(current, threshold float64) string {
	ratio := current / threshold
	if ratio >= 3 {
		return "critical"
	}
	if ratio >= 2 {
		return "high"
	}
	if ratio >= 1.5 {
		return "medium"
	}
	return "low"
}

func severityFromErrorRate(rate float64) string {
	if rate >= 0.2 {
		return "critical"
	}
	if rate >= 0.1 {
		return "high"
	}
	if rate >= 0.05 {
		return "medium"
	}
	return "low"
}

func cloneIncident(inc *models.Incident) *models.Incident {
	data, _ := json.Marshal(inc)
	var copy models.Incident
	_ = json.Unmarshal(data, &copy)
	return &copy
}
func (d *IncidentDetector) SetInvestigation(id string, result *models.Investigation) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	inc, ok := d.incidents[id]
	if !ok {
		return false
	}
	inc.Investigation = result
	return true
}

func (d *IncidentDetector) AddAction(id string, action models.Remediation) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if inc, ok := d.incidents[id]; ok {
		inc.Actions = append(inc.Actions, action)
	}
}
