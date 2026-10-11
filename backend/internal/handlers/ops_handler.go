package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/gemma"
	"faultline-live/backend/internal/incidents"
	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/response"
	"faultline-live/backend/internal/telemetry"
	"faultline-live/backend/internal/faults"

	"github.com/gorilla/mux"
)

type OpsHandler struct {
	db           *database.DB
	tel          *telemetry.Telemetry
	detector     *incidents.IncidentDetector
	faultManager *faults.FaultManager
	gemma        *gemma.GemmaClient
	cfg          *config.Config
	startTime    time.Time
}

func NewOpsHandler(db *database.DB, tel *telemetry.Telemetry, detector *incidents.IncidentDetector, faultManager *faults.FaultManager, gemma *gemma.GemmaClient, cfg *config.Config) *OpsHandler {
	return &OpsHandler{
		db:           db,
		tel:          tel,
		detector:     detector,
		faultManager: faultManager,
		gemma:        gemma,
		cfg:          cfg,
		startTime:    time.Now(),
	}
}

func (h *OpsHandler) RegisterRoutes(r *mux.Router) {
	// Public health checks
	r.HandleFunc("/api/health", h.Health).Methods("GET")
	r.HandleFunc("/api/ready", h.Ready).Methods("GET")
	r.HandleFunc("/api/metrics", h.Metrics).Methods("GET")
	r.HandleFunc("/api/metrics/history", h.MetricsHistory).Methods("GET")

	// Protected ops endpoints
	ops := r.PathPrefix("/api/ops").Subrouter()
	ops.Use(middleware.Auth(&h.cfg.Auth))
	ops.HandleFunc("/overview", h.Overview).Methods("GET")
	ops.HandleFunc("/services", h.Services).Methods("GET")
	ops.HandleFunc("/metrics", h.Metrics).Methods("GET")
	ops.HandleFunc("/incidents", h.ListIncidents).Methods("GET")
	ops.HandleFunc("/incidents/{id}", h.GetIncident).Methods("GET")
	ops.HandleFunc("/incidents/{id}/investigate", h.InvestigateIncident).Methods("POST")
	ops.HandleFunc("/faults", h.ListFaults).Methods("GET")
	ops.HandleFunc("/faults", h.CreateFault).Methods("POST")
	ops.HandleFunc("/faults/{id}", h.DeleteFault).Methods("DELETE")
}

func (h *OpsHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.Success(w, models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime).String(),
	})
}

func (h *OpsHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	})

	// Database check
	if err := h.db.HealthCheck(ctx); err != nil {
		checks["database"] = struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}{Status: "fail", Message: err.Error()}
	} else {
		checks["database"] = struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}{Status: "ok"}
	}

	// Ollama check (if enabled)
	if h.cfg.Ollama.Enabled {
		checks["ollama"] = struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}{Status: "ok", Message: "configured"}
	}

	ready := true
	for _, check := range checks {
		if check.Status == "fail" {
			ready = false
			break
		}
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	response.JSON(w, status, models.ReadyResponse{
		Ready:   ready,
		Checks:  checks,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (h *OpsHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.tel.GetMetrics()
	metrics.ActiveFaults = h.faultManager.ActiveCount()
	response.Success(w, metrics)
}

func (h *OpsHandler) MetricsHistory(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "1h"
	}

	samples, err := h.tel.GetHistory(window)
	if err != nil {
		response.InternalError(w, "Failed to fetch metrics history")
		return
	}

	response.Success(w, models.MetricsHistoryResponse{
		Samples: samples,
		Window:  window,
	})
}

func (h *OpsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	metrics := h.tel.GetMetrics()
	metrics.ActiveFaults = h.faultManager.ActiveCount()

	endpointMetrics := h.tel.GetEndpointMetrics()
	services := make([]models.ServiceStatus, 0, len(endpointMetrics)+2)

	services = append(services, models.ServiceStatus{
		Name:      "api",
		Status:    serviceStatusFromMetrics(metrics),
		LatencyMs: metrics.LatencyP95,
		ErrorRate: metrics.ErrorRate,
		LastCheck: time.Now().Format(time.RFC3339),
	})

	services = append(services, models.ServiceStatus{
		Name:      "database",
		Status:    dbStatusFromMetrics(metrics),
		LatencyMs: metrics.DBLatencyP95,
		LastCheck: time.Now().Format(time.RFC3339),
	})

	for endpoint, em := range endpointMetrics {
		services = append(services, models.ServiceStatus{
			Name:      endpoint,
			Status:    serviceStatusFromMetrics(em),
			LatencyMs: em.LatencyP95,
			ErrorRate: em.ErrorRate,
			LastCheck: time.Now().Format(time.RFC3339),
		})
	}

	incidents := h.detector.GetIncidents()
	incidentSummaries := make([]models.IncidentSummary, 0, len(incidents))
	for _, inc := range incidents {
		incidentSummaries = append(incidentSummaries, models.IncidentSummary{
			ID:               inc.ID,
			Type:             inc.Type,
			Severity:         inc.Severity,
			Title:            inc.Title,
			AffectedEndpoint: inc.AffectedEndpoint,
			StartedAt:        inc.StartedAt.Format(time.RFC3339),
			ResolvedAt:       inc.ResolvedAtFormat(),
			Status:           inc.Status,
		})
	}

	response.Success(w, models.OpsOverviewResponse{
		Services: services,
		Metrics:  metrics,
		Incidents: incidentSummaries,
	})
}

func (h *OpsHandler) Services(w http.ResponseWriter, r *http.Request) {
	metrics := h.tel.GetMetrics()
	endpointMetrics := h.tel.GetEndpointMetrics()

	services := make([]models.ServiceStatus, 0, len(endpointMetrics)+2)

	services = append(services, models.ServiceStatus{
		Name:      "api",
		Status:    serviceStatusFromMetrics(metrics),
		LatencyMs: metrics.LatencyP95,
		ErrorRate: metrics.ErrorRate,
		LastCheck: time.Now().Format(time.RFC3339),
	})

	services = append(services, models.ServiceStatus{
		Name:      "database",
		Status:    dbStatusFromMetrics(metrics),
		LatencyMs: metrics.DBLatencyP95,
		LastCheck: time.Now().Format(time.RFC3339),
	})

	for endpoint, em := range endpointMetrics {
		services = append(services, models.ServiceStatus{
			Name:      endpoint,
			Status:    serviceStatusFromMetrics(em),
			LatencyMs: em.LatencyP95,
			ErrorRate: em.ErrorRate,
			LastCheck: time.Now().Format(time.RFC3339),
		})
	}

	response.Success(w, services)
}

func (h *OpsHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	incidents := h.detector.GetIncidents()
	summaries := make([]models.IncidentSummary, 0, len(incidents))
	for _, inc := range incidents {
		summaries = append(summaries, models.IncidentSummary{
			ID:               inc.ID,
			Type:             inc.Type,
			Severity:         inc.Severity,
			Title:            inc.Title,
			AffectedEndpoint: inc.AffectedEndpoint,
			StartedAt:        inc.StartedAt.Format(time.RFC3339),
			ResolvedAt:       inc.ResolvedAtFormat(),
			Status:           inc.Status,
		})
	}
	response.Success(w, summaries)
}

func (h *OpsHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	inc := h.detector.GetIncident(id)
	if inc == nil {
		response.NotFound(w, "Incident not found")
		return
	}

	detail := models.IncidentDetail{
		IncidentSummary: models.IncidentSummary{
			ID:               inc.ID,
			Type:             inc.Type,
			Severity:         inc.Severity,
			Title:            inc.Title,
			AffectedEndpoint: inc.AffectedEndpoint,
			StartedAt:        inc.StartedAt.Format(time.RFC3339),
			ResolvedAt:       inc.ResolvedAtFormat(),
			Status:           inc.Status,
		},
		Description: inc.Description,
		Observed:    inc.Observed,
		Baseline:    inc.Baseline,
		Threshold:   inc.Threshold,
		Evidence:    inc.Evidence,
	}

	response.Success(w, detail)
}

func (h *OpsHandler) InvestigateIncident(w http.ResponseWriter, r *http.Request) {
	if h.gemma == nil {
		response.ServiceUnavailable(w, "Gemma integration not available")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	inc := h.detector.GetIncident(id)
	if inc == nil {
		response.NotFound(w, "Incident not found")
		return
	}

	req := gemma.InvestigationRequest{
		Incident:        *inc,
		EndpointMetrics: h.tel.GetEndpointMetrics(),
		ActiveFaults:    h.faultManager.GetAll(),
		ServiceHealth: []gemma.ServiceHealth{
			{Name: "api", Status: "healthy", LatencyMs: h.tel.GetMetrics().LatencyP95, ErrorRate: h.tel.GetMetrics().ErrorRate},
			{Name: "database", Status: "healthy", LatencyMs: h.tel.GetMetrics().DBLatencyP95},
		},
	}

	investigation, err := h.gemma.Investigate(r.Context(), req)
	if err != nil {
		response.InternalError(w, "Investigation failed: "+err.Error())
		return
	}

	inc.Investigation = investigation

	response.Created(w, investigation)
}

func (h *OpsHandler) ListFaults(w http.ResponseWriter, r *http.Request) {
	faults := h.faultManager.GetAll()
	response.Success(w, faults)
}

type CreateFaultRequest struct {
	Type         string                 `json:"type"`
	TargetRoute  string                 `json:"target_route"`
	Parameters   map[string]interface{} `json:"parameters"`
}

func (h *OpsHandler) CreateFault(w http.ResponseWriter, r *http.Request) {
	var req CreateFaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Type == "" || req.TargetRoute == "" {
		response.BadRequest(w, "Type and target_route are required")
		return
	}

	userID, _ := middleware.GetUserID(r)
	createdBy := "user:" + strconv.FormatInt(userID, 10)

	fault, err := h.faultManager.Create(r.Context(), req.Type, req.TargetRoute, req.Parameters, createdBy)
	if err != nil {
		if err == faults.ErrFaultsDisabled {
			response.BadRequest(w, "Fault injection is disabled. Enable with FAULTS_ENABLED=true")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, fault)
}

func (h *OpsHandler) DeleteFault(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.faultManager.Delete(r.Context(), id); err != nil {
		if err == faults.ErrFaultNotFound {
			response.NotFound(w, "Fault not found")
			return
		}
		response.InternalError(w, "Failed to delete fault")
		return
	}

	response.NoContent(w)
}

func serviceStatusFromMetrics(m models.MetricsResponse) string {
	if m.ErrorRate > 0.1 || m.LatencyP95 > 1000 {
		return "down"
	}
	if m.ErrorRate > 0.05 || m.LatencyP95 > 500 {
		return "degraded"
	}
	return "healthy"
}

func dbStatusFromMetrics(m models.MetricsResponse) string {
	if m.DBLatencyP95 > 500 {
		return "down"
	}
	if m.DBLatencyP95 > 200 {
		return "degraded"
	}
	return "healthy"
}