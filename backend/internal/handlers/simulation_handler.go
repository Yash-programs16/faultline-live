package handlers

import (
	"context"
	"encoding/json"
	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/response"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type trafficRun struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Target      string     `json:"target"`
	Limit       int        `json:"limit"`
	Concurrency int        `json:"concurrency"`
	Success     int        `json:"success"`
	Failed      int        `json:"failed"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	Latencies   []float64  `json:"latencies_ms"`
}

func (h *OpsHandler) GetTraffic(w http.ResponseWriter, r *http.Request) {
	h.trafficMu.Lock()
	defer h.trafficMu.Unlock()
	response.Success(w, h.traffic)
}
func (h *OpsHandler) StartTraffic(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.DemoMode {
		response.Forbidden(w, "traffic generation requires local demo mode")
		return
	}
	h.trafficMu.Lock()
	if h.traffic != nil && h.traffic.Status == "running" {
		h.trafficMu.Unlock()
		response.Error(w, 409, "traffic already running", "CONFLICT")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	h.trafficCancel = cancel
	run := &trafficRun{ID: models.NewFaultID(), Status: "running", Target: "/api/products", Limit: 30, Concurrency: 3, StartedAt: time.Now(), Latencies: []float64{}}
	h.traffic = run
	// Serialize a copy while locked; workers cannot mutate the response.
	response.Created(w, *run)
	h.trafficMu.Unlock()
	go h.generateTraffic(ctx, cancel, run)
}
func (h *OpsHandler) generateTraffic(ctx context.Context, cancel context.CancelFunc, run *trafficRun) {
	defer cancel()
	client := &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	target := "http://" + net.JoinHostPort(h.cfg.Server.Host, h.cfg.Server.Port) + "/api/products?page=1&pageSize=5"
	var workers sync.WaitGroup
	for j := 0; j < run.Concurrency; j++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < run.Limit/run.Concurrency; i++ {
				if ctx.Err() != nil {
					return
				}
				start := time.Now()
				req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
				if err != nil {
					return
				}
				res, err := client.Do(req)
				ok := err == nil && res.StatusCode >= 200 && res.StatusCode < 300
				if res != nil {
					_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<20))
					res.Body.Close()
				}
				h.trafficMu.Lock()
				if ok {
					run.Success++
				} else {
					run.Failed++
				}
				run.Latencies = append(run.Latencies, float64(time.Since(start).Microseconds())/1000)
				h.trafficMu.Unlock()
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
			}
		}()
	}
	workers.Wait()
	h.trafficMu.Lock()
	defer h.trafficMu.Unlock()
	run.Status = "completed"
	if ctx.Err() != nil {
		run.Status = "stopped"
	}
	now := time.Now()
	run.FinishedAt = &now
}
func (h *OpsHandler) StopTraffic(w http.ResponseWriter, r *http.Request) {
	h.trafficMu.Lock()
	defer h.trafficMu.Unlock()
	if h.trafficCancel != nil {
		h.trafficCancel()
	}
	response.Success(w, map[string]string{"status": "stop_requested"})
}
func (h *OpsHandler) Stop() {
	h.trafficMu.Lock()
	if h.trafficCancel != nil {
		h.trafficCancel()
	}
	h.trafficMu.Unlock()
}

func (h *OpsHandler) RemediateIncident(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	inc := h.detector.GetIncident(id)
	if inc == nil {
		response.NotFound(w, "incident not found")
		return
	}
	var req struct {
		Action          string  `json:"action"`
		FaultID         string  `json:"fault_id"`
		Target          string  `json:"target"`
		RequestsPerSec  float64 `json:"requests_per_second"`
		Burst           int     `json:"burst"`
		Reason          string  `json:"reason"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req) != nil || len(strings.TrimSpace(req.Reason)) < 5 || len(req.Reason) > 1000 {
		response.BadRequest(w, "action and a reason (5–1000 characters) are required")
		return
	}
	switch req.Action {
	case "stop_fault":
		h.remediateStopFault(w, r, id, inc, req.FaultID, req.Reason)
	case "enable_rate_limit":
		h.remediateEnableRateLimit(w, r, id, inc, req.Target, req.RequestsPerSec, req.Burst, req.Reason)
	case "disable_defense":
		h.remediateDisableDefense(w, r, id, inc, req.Target, req.Reason)
	default:
		response.BadRequest(w, "unsupported action: use stop_fault, enable_rate_limit or disable_defense")
	}
}

func (h *OpsHandler) remediateStopFault(w http.ResponseWriter, r *http.Request, id string, inc *models.Incident, faultID, reason string) {
	if faultID == "" {
		response.BadRequest(w, "stop_fault requires fault_id")
		return
	}
	fault, err := h.faultManager.Get(faultID)
	if err != nil {
		response.NotFound(w, "fault not found")
		return
	}
	if inc.AffectedEndpoint != "global" && inc.AffectedEndpoint != "database" && inc.AffectedEndpoint != "GET "+fault.TargetRoute {
		response.Forbidden(w, "fault is outside incident scope")
		return
	}
	action := models.Remediation{ID: models.NewFaultID(), IncidentID: id, FaultID: fault.ID, Action: "stop_fault", Reason: reason, Status: "requested", CreatedAt: time.Now()}
	if h.auditAction(r.Context(), action) != nil {
		response.InternalError(w, "audit unavailable; remediation was not executed")
		return
	}
	action.Status = "succeeded"
	if err = h.faultManager.Delete(r.Context(), fault.ID); err != nil {
		action.Status = "failed"
	}
	if h.auditAction(r.Context(), action) != nil {
		action.Status += "_audit_failed"
	}
	h.detector.AddAction(id, action)
	response.Success(w, map[string]interface{}{"action": action, "recovery": "pending: requires fresh healthy workload samples over two detector checks", "metrics": h.tel.GetMetrics()})
}

func (h *OpsHandler) remediateEnableRateLimit(w http.ResponseWriter, r *http.Request, id string, inc *models.Incident, target string, rps float64, burst int, reason string) {
	if target == "" {
		target = inc.AffectedEndpoint
		if strings.HasPrefix(target, "GET ") {
			target = strings.TrimPrefix(target, "GET ")
		}
		if target == "global" || target == "database" || target == "" {
			target = "/api/products"
		}
	}
	if rps <= 0 {
		rps = 10
	}
	createdBy := "local-demo"
	if userID, ok := middleware.GetUserID(r); ok {
		createdBy = "user:" + strconv.FormatInt(userID, 10)
	}
	d, err := h.defenseManager.EnableRateLimit(target, rps, burst, createdBy)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	action := models.Remediation{ID: models.NewFaultID(), IncidentID: id, FaultID: d.ID, Action: "enable_rate_limit", Reason: reason, Status: "succeeded", CreatedAt: time.Now()}
	_ = h.auditAction(r.Context(), action)
	h.detector.AddAction(id, action)
	response.Success(w, map[string]interface{}{"action": action, "defense": d, "recovery": "rate limiter active: excess requests now receive HTTP 429; verify request rate and error rate recover over two detector checks", "metrics": h.tel.GetMetrics()})
}

func (h *OpsHandler) remediateDisableDefense(w http.ResponseWriter, r *http.Request, id string, inc *models.Incident, target, reason string) {
	if target == "" {
		target = inc.AffectedEndpoint
		if strings.HasPrefix(target, "GET ") {
			target = strings.TrimPrefix(target, "GET ")
		}
	}
	if err := h.defenseManager.Disable(target); err != nil {
		response.NotFound(w, "defense not found for target")
		return
	}
	action := models.Remediation{ID: models.NewFaultID(), IncidentID: id, Action: "disable_defense", Reason: reason, Status: "succeeded", CreatedAt: time.Now()}
	_ = h.auditAction(r.Context(), action)
	h.detector.AddAction(id, action)
	response.Success(w, map[string]interface{}{"action": action, "recovery": "defense removed; monitor telemetry for recurrence", "metrics": h.tel.GetMetrics()})
}
func (h *OpsHandler) auditAction(ctx context.Context, action models.Remediation) error {
	if h.db == nil {
		return fmt.Errorf("audit database unavailable")
	}
	data, err := json.Marshal(action)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err = h.db.Pool.Exec(ctx, `INSERT INTO audit_logs (action,resource_type,resource_id,metadata) VALUES ($1,'incident',$2,$3)`, "remediation."+action.Status, action.IncidentID, data)
	return err
}
func (h *OpsHandler) persistInvestigation(ctx context.Context, i *models.Investigation) error {
	if h.db == nil {
		return fmt.Errorf("database unavailable")
	}
	hypotheses, err := json.Marshal(i.Hypotheses)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err = h.db.Pool.Exec(ctx, `INSERT INTO investigations (id,incident_id,summary,root_cause,hypotheses,missing_evidence,recommended_steps,recovery_recommendations,confidence,confidence_explanation,model,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, i.ID, i.IncidentID, i.Summary, i.RootCause, hypotheses, i.MissingEvidence, i.RecommendedSteps, i.RecoveryRecommendations, i.Confidence, i.ConfidenceExplanation, i.Model, i.CreatedAt)
	return err
}
