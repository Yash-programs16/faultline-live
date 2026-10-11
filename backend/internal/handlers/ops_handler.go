package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/defenses"
	"faultline-live/backend/internal/faults"
	"faultline-live/backend/internal/gemma"
	"faultline-live/backend/internal/incidents"
	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/response"
	"faultline-live/backend/internal/telemetry"

	"github.com/gorilla/mux"
)

type OpsHandler struct {
	db              *database.DB
	tel             *telemetry.Telemetry
	detector        *incidents.IncidentDetector
	faultManager    *faults.FaultManager
	defenseManager  *defenses.Manager
	gemma           *gemma.GemmaClient
	gemini          *gemma.GeminiClient
	cfg             *config.Config
	startTime       time.Time
	trafficMu       sync.Mutex
	traffic         *trafficRun
	trafficCancel   context.CancelFunc
	investigationMu sync.Mutex
	investigating   map[string]bool
}

func NewOpsHandler(db *database.DB, tel *telemetry.Telemetry, detector *incidents.IncidentDetector, faultManager *faults.FaultManager, gemma *gemma.GemmaClient, cfg *config.Config) *OpsHandler {
	return &OpsHandler{
		db:             db,
		tel:            tel,
		detector:       detector,
		faultManager:   faultManager,
		defenseManager: defenses.NewManager(),
		gemma:          gemma,
		cfg:            cfg,
		startTime:      time.Now(),
		investigating:  map[string]bool{},
	}
}

// NewOpsHandlerWithDefenses allows wiring a shared defense manager (used by main).
func NewOpsHandlerWithDefenses(db *database.DB, tel *telemetry.Telemetry, detector *incidents.IncidentDetector, faultManager *faults.FaultManager, defenseManager *defenses.Manager, gemma *gemma.GemmaClient, cfg *config.Config) *OpsHandler {
	h := NewOpsHandler(db, tel, detector, faultManager, gemma, cfg)
	if defenseManager != nil {
		h.defenseManager = defenseManager
	}
	return h
}

// SetGeminiClient wires the optional Gemini cloud investigator. It is a
// setter (rather than a constructor argument) so existing call sites and
// tests keep compiling unchanged.
func (h *OpsHandler) SetGeminiClient(c *gemma.GeminiClient) {
	h.gemini = c
}

// activeModel reports the investigator shown in the UI: Gemini when a key
// is configured, otherwise the Ollama model.
func (h *OpsHandler) activeModel() string {
	if h.gemini != nil && h.gemini.Configured() {
		return h.gemini.ModelName()
	}
	return h.cfg.Ollama.Model
}

func (h *OpsHandler) RegisterRoutes(r *mux.Router) {
	// Public health checks
	r.HandleFunc("/api/health", h.Health).Methods("GET")
	r.HandleFunc("/api/ready", h.Ready).Methods("GET")
	r.HandleFunc("/api/metrics", h.Metrics).Methods("GET")
	r.HandleFunc("/api/metrics/history", h.MetricsHistory).Methods("GET")

	// Protected ops endpoints
	ops := r.PathPrefix("/api/ops").Subrouter()
	if h.cfg.OpsAuthRequired {
		ops.Use(middleware.Auth(&h.cfg.Auth))
		ops.Use(middleware.Operators(h.cfg.OperatorIDs))
	} else if h.cfg.AllowPublicDemo {
		// Hosted demo (Render + Vercel): requests arrive from the public
		// internet, so the loopback peer check cannot apply. The sim
		// controls stay bounded (expiry, error-rate caps) by design.
	} else {
		ops.Use(middleware.LocalDemo)
	}
	ops.HandleFunc("/traffic", h.GetTraffic).Methods("GET")
	ops.HandleFunc("/traffic", h.StartTraffic).Methods("POST")
	ops.HandleFunc("/traffic", h.StopTraffic).Methods("DELETE")
	ops.HandleFunc("/incidents/{id}/remediate", h.RemediateIncident).Methods("POST")
	ops.HandleFunc("/overview", h.Overview).Methods("GET")
	ops.HandleFunc("/services", h.Services).Methods("GET")
	ops.HandleFunc("/metrics", h.Metrics).Methods("GET")
	ops.HandleFunc("/incidents", h.ListIncidents).Methods("GET")
	ops.HandleFunc("/incidents/{id}", h.GetIncident).Methods("GET")
	ops.HandleFunc("/incidents/{id}/investigate", h.InvestigateIncident).Methods("POST")
	ops.HandleFunc("/incidents/{id}/defenses", h.IncidentDefenses).Methods("GET")
	ops.HandleFunc("/defenses", h.ListDefenses).Methods("GET")
	ops.HandleFunc("/defenses", h.CreateDefense).Methods("POST")
	ops.HandleFunc("/defenses/{id}", h.DeleteDefense).Methods("DELETE")
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
		}{Status: "fail", Message: "database connectivity check failed"}
	} else {
		checks["database"] = struct {
			Status  string `json:"status"`
			Message string `json:"message,omitempty"`
		}{Status: "ok"}
	}

	// Verify that Ollama is reachable and the configured model is installed.
	if h.cfg.Ollama.Enabled {
		if err := h.gemma.CheckReady(ctx); err != nil {
			checks["ollama"] = struct {
				Status  string `json:"status"`
				Message string `json:"message,omitempty"`
			}{Status: "fail", Message: err.Error()}
		} else {
			checks["ollama"] = struct {
				Status  string `json:"status"`
				Message string `json:"message,omitempty"`
			}{Status: "ok", Message: "Ollama reachable; configured model installed"}
		}
	}

	// Verify the Gemini cloud investigator when an API key is configured.
	if h.gemini != nil && h.gemini.Configured() {
		if err := h.gemini.CheckReady(ctx); err != nil {
			checks["gemini"] = struct {
				Status  string `json:"status"`
				Message string `json:"message,omitempty"`
			}{Status: "fail", Message: err.Error()}
		} else {
			checks["gemini"] = struct {
				Status  string `json:"status"`
				Message string `json:"message,omitempty"`
			}{Status: "ok", Message: "Gemini API reachable; key accepted for " + h.gemini.ModelName()}
		}
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
		Ready:     ready,
		Checks:    checks,
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
		response.BadRequest(w, err.Error())
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
		DemoMode:             h.cfg.DemoMode,
		FaultsEnabled:        h.cfg.Faults.Enabled,
		MetricsWindowSeconds: h.cfg.Telemetry.MetricsWindow.Seconds(),
		Model:                h.activeModel(),
		Services:             services,
		Metrics:              metrics,
		Incidents:            incidentSummaries,
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
		Description:   inc.Description,
		Observed:      inc.Observed,
		Baseline:      inc.Baseline,
		Threshold:     inc.Threshold,
		Evidence:      inc.Evidence,
		Investigation: inc.Investigation,
		Actions:       inc.Actions,
	}

	response.Success(w, detail)
}

func (h *OpsHandler) InvestigateIncident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	inc := h.detector.GetIncident(id)
	if inc == nil {
		response.NotFound(w, "Incident not found")
		return
	}
	// Return a cached investigation immediately so repeated clicks never
	// surface "failed" when the model is slow or offline.
	if inc.Investigation != nil {
		response.Success(w, inc.Investigation)
		return
	}

	h.investigationMu.Lock()
	if h.investigating[id] {
		h.investigationMu.Unlock()
		response.Error(w, http.StatusConflict, "investigation already running", "CONFLICT")
		return
	}
	h.investigating[id] = true
	h.investigationMu.Unlock()
	defer func() { h.investigationMu.Lock(); delete(h.investigating, id); h.investigationMu.Unlock() }()

	activeFaults := h.faultManager.GetAll()

	// deliverFallback persists best-effort and always returns usable
	// hardcoded data (what failed + when) instead of an Ollama error.
	deliverFallback := func(ollamaErr string) {
		fb := gemma.BuildFallbackInvestigation(*inc, activeFaults, ollamaErr)
		_ = h.persistInvestigation(r.Context(), fb)
		h.detector.SetInvestigation(id, fb)
		response.Created(w, fb)
	}

	req := gemma.InvestigationRequest{
		Incident:        *inc,
		EndpointMetrics: h.tel.GetEndpointMetrics(),
		ActiveFaults:    activeFaults,
		ServiceHealth: []gemma.ServiceHealth{
			{Name: "api", Status: serviceStatusFromMetrics(h.tel.GetMetrics()), LatencyMs: h.tel.GetMetrics().LatencyP95, ErrorRate: h.tel.GetMetrics().ErrorRate},
			{Name: "database", Status: dbStatusFromMetrics(h.tel.GetMetrics()), LatencyMs: h.tel.GetMetrics().DBLatencyP95},
		},
	}

	deliverInvestigation := func(investigation *models.Investigation) {
		// Persist best-effort: even if the DB write fails, still return
		// the model findings so the UI never shows a misleading "failed".
		_ = h.persistInvestigation(r.Context(), investigation)
		h.detector.SetInvestigation(id, investigation)
		response.Created(w, investigation)
	}

	// Provider chain: Gemini (cloud) first, then Ollama (local), then the
	// deterministic offline analysis. The first success wins.
	var providerErrs []string
	if h.gemini != nil && h.gemini.Configured() {
		if investigation, err := h.gemini.Investigate(r.Context(), req); err == nil {
			deliverInvestigation(investigation)
			return
		} else {
			providerErrs = append(providerErrs, "gemini: "+err.Error())
		}
	}

	if h.gemma != nil {
		if err := h.gemma.CheckReady(r.Context()); err != nil {
			providerErrs = append(providerErrs, "ollama: "+err.Error())
		} else if investigation, err := h.gemma.Investigate(r.Context(), req); err == nil {
			deliverInvestigation(investigation)
			return
		} else {
			providerErrs = append(providerErrs, "ollama: "+err.Error())
		}
	} else {
		providerErrs = append(providerErrs, "ollama: integration not available")
	}

	combined := ""
	for i, e := range providerErrs {
		if i > 0 {
			combined += " | "
		}
		combined += e
	}
	deliverFallback(combined)
}

func (h *OpsHandler) ListFaults(w http.ResponseWriter, r *http.Request) {
	faults := h.faultManager.GetAll()
	response.Success(w, faults)
}

type CreateFaultRequest struct {
	Type        string                 `json:"type"`
	TargetRoute string                 `json:"target_route"`
	Parameters  map[string]interface{} `json:"parameters"`
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

	createdBy := "local-demo"
	if userID, ok := middleware.GetUserID(r); ok {
		createdBy = "user:" + strconv.FormatInt(userID, 10)
	}

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

func (h *OpsHandler) ListDefenses(w http.ResponseWriter, r *http.Request) {
	response.Success(w, h.defenseManager.List())
}

type CreateDefenseRequest struct {
	Type           string  `json:"type"`
	Target         string  `json:"target"`
	RequestsPerSec float64 `json:"requests_per_second"`
	Burst          int     `json:"burst"`
}

func (h *OpsHandler) CreateDefense(w http.ResponseWriter, r *http.Request) {
	var req CreateDefenseRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if req.Type != "" && req.Type != "rate_limit" {
		response.BadRequest(w, "Only rate_limit defenses are enforced in this build")
		return
	}
	if req.Target == "" {
		response.BadRequest(w, "Target endpoint is required (e.g. /api/products)")
		return
	}
	createdBy := "local-demo"
	if userID, ok := middleware.GetUserID(r); ok {
		createdBy = "user:" + strconv.FormatInt(userID, 10)
	}
	d, err := h.defenseManager.EnableRateLimit(req.Target, req.RequestsPerSec, req.Burst, createdBy)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, d)
}

func (h *OpsHandler) DeleteDefense(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.defenseManager.Disable(id); err != nil {
		response.NotFound(w, "Defense not found")
		return
	}
	response.NoContent(w)
}

func (h *OpsHandler) IncidentDefenses(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	inc := h.detector.GetIncident(id)
	if inc == nil {
		response.NotFound(w, "Incident not found")
		return
	}
	active := h.defenseManager.List()
	hasFault := len(h.faultManager.GetAll()) > 0
	response.Success(w, map[string]interface{}{
		"incident_id": inc.ID,
		"type":        inc.Type,
		"target":      inc.AffectedEndpoint,
		"started_at":  inc.StartedAt.Format(time.RFC3339),
		"defenses":    h.defenseManager.List(),
		"suggestions": defenses.SuggestForIncident(inc.Type, inc.AffectedEndpoint, active, hasFault),
	})
}

func serviceStatusFromMetrics(m models.MetricsResponse) string {
	if m.RequestCount == 0 {
		return "unknown"
	}
	if m.ErrorRate > 0.1 || m.LatencyP95 > 1000 {
		return "down"
	}
	if m.ErrorRate > 0.05 || m.LatencyP95 > 500 {
		return "degraded"
	}
	return "healthy"
}

func dbStatusFromMetrics(m models.MetricsResponse) string {
	if m.DBObservationCount == 0 {
		return "unknown"
	}
	if m.DBLatencyP95 > 500 {
		return "down"
	}
	if m.DBLatencyP95 > 200 {
		return "degraded"
	}
	return "healthy"
}
