package faults

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/telemetry"
)

var (
	ErrFaultNotFound    = errors.New("fault not found")
	ErrFaultsDisabled   = errors.New("fault injection is disabled")
	ErrInvalidFaultType = errors.New("invalid fault type")
	ErrInvalidParameter = errors.New("invalid fault parameter")
)

type FaultError struct {
	Code int
	Msg  string
}

func (e *FaultError) Error() string {
	return e.Msg
}

func (e *FaultError) StatusCode() int {
	return e.Code
}

type FaultManager struct {
	db         *database.DB
	tel        *telemetry.Telemetry
	logger     *slog.Logger
	config     FaultConfig
	mu         sync.RWMutex
	faults     map[string]*models.Fault
	stopCh     chan struct{}
}

type FaultConfig struct {
	Enabled       bool
	MaxDuration   time.Duration
	MaxLatencyMs  int
	MaxErrorRate  float64
}

func NewFaultManager(db *database.DB, tel *telemetry.Telemetry, logger *slog.Logger, cfg FaultConfig) *FaultManager {
	return &FaultManager{
		db:     db,
		tel:    tel,
		logger: logger,
		config: cfg,
		faults: make(map[string]*models.Fault),
		stopCh: make(chan struct{}),
	}
}

func (m *FaultManager) Start(ctx context.Context) {
	m.loadFaults(ctx)
	go m.cleanupExpired(ctx)
}

func (m *FaultManager) Stop() {
	close(m.stopCh)
}

func (m *FaultManager) Create(ctx context.Context, faultType, targetRoute string, params map[string]interface{}, createdBy string) (*models.Fault, error) {
	if !m.config.Enabled {
		return nil, ErrFaultsDisabled
	}

	// Validate fault type
	validTypes := map[string]bool{"latency": true, "error": true, "db_failure": true}
	if !validTypes[faultType] {
		return nil, ErrInvalidFaultType
	}

	// Validate parameters
	if err := m.validateParams(faultType, params); err != nil {
		return nil, err
	}

	// Check if fault already exists for this route+type
	m.mu.RLock()
	for _, f := range m.faults {
		if f.Type == faultType && f.TargetRoute == targetRoute && f.Enabled {
			m.mu.RUnlock()
			return nil, fmt.Errorf("fault of type %s already active for route %s", faultType, targetRoute)
		}
	}
	m.mu.RUnlock()

	now := time.Now()
	expiresAt := now.Add(m.config.MaxDuration)

	fault := &models.Fault{
		ID:          models.NewFaultID(),
		Type:        faultType,
		TargetRoute: targetRoute,
		Parameters:  params,
		Enabled:     true,
		CreatedAt:   now,
		ExpiresAt:   &expiresAt,
		CreatedBy:   createdBy,
	}

	m.mu.Lock()
	m.faults[fault.ID] = fault
	m.mu.Unlock()

	m.persistFault(ctx, fault)
	m.logger.Info("fault created", "id", fault.ID, "type", fault.Type, "route", targetRoute, "by", createdBy)

	return fault, nil
}

func (m *FaultManager) validateParams(faultType string, params map[string]interface{}) error {
	switch faultType {
	case "latency":
		ms, ok := params["latency_ms"].(float64)
		if !ok {
			return fmt.Errorf("%w: latency_ms required", ErrInvalidParameter)
		}
		if ms < 0 || ms > float64(m.config.MaxLatencyMs) {
			return fmt.Errorf("%w: latency_ms must be between 0 and %d", ErrInvalidParameter, m.config.MaxLatencyMs)
		}
	case "error":
		rate, ok := params["error_rate"].(float64)
		if !ok {
			return fmt.Errorf("%w: error_rate required", ErrInvalidParameter)
		}
		if rate < 0 || rate > m.config.MaxErrorRate {
			return fmt.Errorf("%w: error_rate must be between 0 and %.2f", ErrInvalidParameter, m.config.MaxErrorRate)
		}
		// Optional: error_code
		if code, ok := params["error_code"].(float64); ok {
			if code < 500 || code > 599 {
				return fmt.Errorf("%w: error_code must be 5xx", ErrInvalidParameter)
			}
		}
	case "db_failure":
		rate, ok := params["failure_rate"].(float64)
		if !ok {
			return fmt.Errorf("%w: failure_rate required", ErrInvalidParameter)
		}
		if rate < 0 || rate > 1 {
			return fmt.Errorf("%w: failure_rate must be between 0 and 1", ErrInvalidParameter)
		}
	}
	return nil
}

func (m *FaultManager) GetAll() []models.Fault {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]models.Fault, 0, len(m.faults))
	for _, f := range m.faults {
		result = append(result, *f)
	}
	return result
}

func (m *FaultManager) Get(id string) (*models.Fault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, ok := m.faults[id]
	if !ok {
		return nil, ErrFaultNotFound
	}
	return f, nil
}

func (m *FaultManager) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	fault, ok := m.faults[id]
	if !ok {
		m.mu.Unlock()
		return ErrFaultNotFound
	}
	fault.Enabled = false
	delete(m.faults, id)
	m.mu.Unlock()

	m.persistFault(ctx, fault)
	m.logger.Info("fault deleted", "id", id)
	return nil
}

func (m *FaultManager) DeleteAll(ctx context.Context) {
	m.mu.Lock()
	for _, fault := range m.faults {
		fault.Enabled = false
		m.persistFault(ctx, fault)
	}
	m.faults = make(map[string]*models.Fault)
	m.mu.Unlock()
	m.logger.Info("all faults cleared")
}

func (m *FaultManager) Apply(route string, handler func() error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, fault := range m.faults {
		if !fault.Enabled {
			continue
		}
		if !m.matchRoute(fault.TargetRoute, route) {
			continue
		}

		switch fault.Type {
		case "latency":
			if ms, ok := fault.Parameters["latency_ms"].(float64); ok {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		case "error":
			rate := 0.0
			if r, ok := fault.Parameters["error_rate"].(float64); ok {
				rate = r
			}
			if rate > 0 && randomFloat() < rate {
				code := 500
				if c, ok := fault.Parameters["error_code"].(float64); ok {
					code = int(c)
				}
				return &FaultError{Code: code, Msg: "injected error"}
			}
		case "db_failure":
			rate := 0.0
			if r, ok := fault.Parameters["failure_rate"].(float64); ok {
				rate = r
			}
			if rate > 0 && randomFloat() < rate {
				return &FaultError{Code: 500, Msg: "injected database failure"}
			}
		}
	}
	return handler()
}

func (m *FaultManager) matchRoute(pattern, route string) bool {
	if pattern == "*" || pattern == route {
		return true
	}
	// Simple prefix match for patterns ending with *
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(route) >= len(prefix) && route[:len(prefix)] == prefix
	}
	return false
}

func (m *FaultManager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, f := range m.faults {
		if f.Enabled {
			count++
		}
	}
	return count
}

func (m *FaultManager) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := r.URL.Path
			// Use the route pattern from middleware if available
			if pattern := r.Context().Value("route_pattern"); pattern != nil {
				if p, ok := pattern.(string); ok {
					route = p
				}
			}

			err := m.Apply(route, func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			if err != nil {
				// Fault injected an error
				if faultErr, ok := err.(interface{ StatusCode() int }); ok {
					http.Error(w, err.Error(), faultErr.StatusCode())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		})
	}
}

func (m *FaultManager) loadFaults(ctx context.Context) {
	rows, err := m.db.Pool.Query(ctx, `
		SELECT id, type, target_route, parameters, enabled, created_at, expires_at, created_by
		FROM faults WHERE enabled = true AND (expires_at IS NULL OR expires_at > NOW())
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f models.Fault
		var paramsJSON []byte
		var expiresAt *time.Time
		if err := rows.Scan(&f.ID, &f.Type, &f.TargetRoute, &paramsJSON, &f.Enabled, &f.CreatedAt, &expiresAt, &f.CreatedBy); err != nil {
			continue
		}
		f.ExpiresAt = expiresAt
		json.Unmarshal(paramsJSON, &f.Parameters)
		m.mu.Lock()
		m.faults[f.ID] = &f
		m.mu.Unlock()
	}
}

func (m *FaultManager) persistFault(ctx context.Context, fault *models.Fault) {
	paramsJSON, _ := json.Marshal(fault.Parameters)
	_, err := m.db.Pool.Exec(ctx, `
		INSERT INTO faults (id, type, target_route, parameters, enabled, created_at, expires_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			expires_at = EXCLUDED.expires_at
	`, fault.ID, fault.Type, fault.TargetRoute, paramsJSON, fault.Enabled, fault.CreatedAt, fault.ExpiresAt, fault.CreatedBy)
	if err != nil {
		m.logger.Error("persist fault", "error", err)
	}
}

func (m *FaultManager) cleanupExpired(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for id, fault := range m.faults {
				if fault.ExpiresAt != nil && now.After(*fault.ExpiresAt) && fault.Enabled {
					fault.Enabled = false
					m.persistFault(ctx, fault)
					delete(m.faults, id)
					m.logger.Info("fault expired", "id", id)
				}
			}
			m.mu.Unlock()
		}
	}
}

func randomFloat() float64 {
	// Simple pseudo-random for demo
	return float64(time.Now().UnixNano()%1000) / 1000.0
}