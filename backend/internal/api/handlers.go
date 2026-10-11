package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/faults"
	"faultline-live/backend/internal/payment"
	"faultline-live/backend/internal/store"
	"faultline-live/backend/internal/telemetry"
)

type Handler struct {
	store       *store.Store
	telemetry   *telemetry.Collector
	faultCtrl   *faults.FaultController
	paymentSim  *payment.Simulator
	config      *config.Config
}

func NewHandler(store *store.Store, telemetry *telemetry.Collector, faultCtrl *faults.FaultController, paymentSim *payment.Simulator, config *config.Config) *Handler {
	return &Handler{
		store:      store,
		telemetry:  telemetry,
		faultCtrl:  faultCtrl,
		paymentSim: paymentSim,
		config:     config,
	}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/health", h.Health).Methods("GET")
	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.HandleFunc("/api/products/{id}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/cart/validate", h.ValidateCart).Methods("POST")
	r.HandleFunc("/api/checkout", h.Checkout).Methods("POST")
	r.HandleFunc("/api/metrics/live", h.MetricsLive).Methods("GET")
	r.HandleFunc("/api/incidents", h.Incidents).Methods("GET")
	r.HandleFunc("/api/demo/fault", h.ActivateFault).Methods("POST")
	r.HandleFunc("/api/demo/fault", h.DeactivateFault).Methods("DELETE")
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbStatus := "connected"
	if err := h.store.Pool().Ping(ctx); err != nil {
		dbStatus = "disconnected"
	}

	status := "healthy"
	if dbStatus == "disconnected" {
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":    status,
		"database":  dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.ListProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"products": products})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	product, err := h.store.GetProduct(r.Context(), id)
	if err != nil {
		if err == store.ErrProductNotFound {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"product": product})
}

func (h *Handler) ValidateCart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []store.CartItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.store.ValidateCart(r.Context(), req.Items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate cart")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items          []store.CartItem `json:"items"`
		PaymentMethod  string           `json:"payment_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "cart is empty")
		return
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "simulated_card"
	}

	result, err := h.store.ValidateCart(r.Context(), req.Items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate cart")
		return
	}

	if !result.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "cart validation failed: " + result.Errors[0],
		})
		return
	}

	start := time.Now()

	var paymentResult payment.PaymentResult
	if h.paymentSim != nil {
		paymentResult = h.paymentSim.ProcessPayment(r.Context())
	} else {
		paymentResult = payment.PaymentResult{
			Success:          true,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}
	}

	processingTimeMs := time.Since(start).Milliseconds()
	if !paymentResult.Success {
		var status int
		if paymentResult.ErrorMessage == "payment processing timeout" {
			status = http.StatusGatewayTimeout
		} else {
			status = http.StatusBadGateway
		}
		writeJSON(w, status, map[string]any{
			"success":            false,
			"message":            paymentResult.ErrorMessage,
			"processing_time_ms": processingTimeMs,
		})
		return
	}

	orderID, err := h.store.Checkout(r.Context(), req.Items)
	if err != nil {
		errMsg := err.Error()
		var status int
		switch {
		case contains(errMsg, "not found"):
			status = http.StatusNotFound
		case contains(errMsg, "insufficient stock"):
			status = http.StatusConflict
		case contains(errMsg, "invalid quantity"):
			status = http.StatusBadRequest
		default:
			log.Printf("checkout error: %v", err)
			status = http.StatusInternalServerError
			errMsg = "internal server error"
		}
		writeJSON(w, status, map[string]any{
			"success":            false,
			"message":            errMsg,
			"processing_time_ms": processingTimeMs,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":            true,
		"order_id":           orderID,
		"message":            "order placed successfully",
		"processing_time_ms": processingTimeMs,
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (h *Handler) MetricsLive(w http.ResponseWriter, r *http.Request) {
	services := []string{"checkout", "payment", "cart", "catalog", "api"}
	var allWindows []telemetry.AggregatedWindow

	for _, svc := range services {
		windows := h.telemetry.GetRecentWindows(svc, 30)
		allWindows = append(allWindows, windows...)
	}

	response := map[string]any{
		"windows":       allWindows,
		"current_fault": h.faultCtrl.GetState(),
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Incidents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"incidents": []any{},
	})
}

func (h *Handler) ActivateFault(w http.ResponseWriter, r *http.Request) {
	if h.config.Faults.LocalOnly && !isLocalRequest(r) {
		writeError(w, http.StatusForbidden, "fault injection only allowed from localhost")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	scenario, params, durationSeconds, err := h.faultCtrl.ValidateRequest(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	state, err := h.faultCtrl.Activate(scenario, params, durationSeconds)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	jsonData, _ := h.faultCtrl.StateToJSON(state)
	var response map[string]any
	json.Unmarshal(jsonData, &response)
	response["success"] = true
	response["message"] = "fault activated"

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) DeactivateFault(w http.ResponseWriter, r *http.Request) {
	if h.config.Faults.LocalOnly && !isLocalRequest(r) {
		writeError(w, http.StatusForbidden, "fault injection only allowed from localhost")
		return
	}

	state, err := h.faultCtrl.Deactivate()
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	jsonData, _ := h.faultCtrl.StateToJSON(state)
	var response map[string]any
	json.Unmarshal(jsonData, &response)
	response["success"] = true
	response["message"] = "fault deactivated"

	writeJSON(w, http.StatusOK, response)
}

func isLocalRequest(r *http.Request) bool {
	host := r.RemoteAddr
	if host == "::1" || host == "127.0.0.1" || host == "[::1]" {
		return true
	}
	if len(host) > 0 && host[0] == '[' {
		return false
	}
	for i := 0; i < len(host); i++ {
		if host[i] == ':' {
			host = host[:i]
			break
		}
	}
	return host == "127.0.0.1" || host == "localhost"
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error":   http.StatusText(status),
		"message": message,
		"status":  status,
	})
}