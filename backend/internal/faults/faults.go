package faults

import (
	"context"
	"encoding/json"
	"errors"
	"faultline-live/backend/internal/response"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
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
	db       *database.DB
	tel      *telemetry.Telemetry
	logger   *slog.Logger
	config   FaultConfig
	mu       sync.RWMutex
	faults   map[string]*models.Fault
	stopCh   chan struct{}
	stopOnce sync.Once
}

type FaultConfig struct {
	Enabled      bool
	MaxDuration  time.Duration
	MaxLatencyMs int
	MaxErrorRate float64
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
	m.stopOnce.Do(func() { close(m.stopCh) })
}

func (m *FaultManager) Create(ctx context.Context, faultType, targetRoute string, params map[string]interface{}, createdBy string) (*models.Fault, error) {
	if !m.config.Enabled {
		return nil, ErrFaultsDisabled
	}

	if targetRoute != "/api/products" {
		return nil, fmt.Errorf("%w: only /api/products is an allowed demo target", ErrInvalidParameter)
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
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.faults {
		if f.Type == faultType && f.TargetRoute == targetRoute && isActive(f) {
			return nil, fmt.Errorf("fault of type %s already active for route %s", faultType, targetRoute)
		}
	}

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

	if err := m.persistFault(ctx, fault); err != nil {
		return nil, err
	}
	m.faults[fault.ID] = fault
	m.logger.Info("fault created", "id", fault.ID, "type", fault.Type, "route", targetRoute, "by", createdBy)

	return cloneFault(fault), nil
}

func (m *FaultManager) validateParams(faultType string, params map[string]interface{}) error {
	switch faultType {
	case "latency":
		ms, ok := params["latency_ms"].(float64)
		if !ok {
			return fmt.Errorf("%w: latency_ms required", ErrInvalidParameter)
		}
		if math.IsNaN(ms) || math.IsInf(ms, 0) || ms < 0 || ms > float64(m.config.MaxLatencyMs) {
			return fmt.Errorf("%w: latency_ms must be between 0 and %d", ErrInvalidParameter, m.config.MaxLatencyMs)
		}
	case "error":
		rate, ok := params["error_rate"].(float64)
		if !ok {
			return fmt.Errorf("%w: error_rate required", ErrInvalidParameter)
		}
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > m.config.MaxErrorRate {
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
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > m.config.MaxErrorRate {
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
		if isActive(f) {
			result = append(result, *cloneFault(f))
		}
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
	return cloneFault(f), nil
}

func (m *FaultManager) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	fault, ok := m.faults[id]
	if !ok {
		m.mu.Unlock()
		return ErrFaultNotFound
	}
	disabled := cloneFault(fault)
	disabled.Enabled = false
	if err := m.persistFault(ctx, disabled); err != nil {
		m.mu.Unlock()
		return err
	}
	delete(m.faults, id)
	m.mu.Unlock()
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
	return m.applyContext(context.Background(), route, handler)
}

func (m *FaultManager) applyContext(ctx context.Context, route string, handler func() error) error {
	if !m.config.Enabled {
		return handler()
	}
	// Snapshot matching active faults while locked, then release the lock before
	// applying delays or running the downstream handler. Holding an RLock across
	// handler() deadlocks POST /api/ops/faults when it tries to acquire the write lock.
	m.mu.RLock()
	active := make([]models.Fault, 0, len(m.faults))
	for _, fault := range m.faults {
		if isActive(fault) && m.matchRoute(fault.TargetRoute, route) {
			copyFault := *fault
			copyFault.Parameters = make(map[string]interface{}, len(fault.Parameters))
			for key, value := range fault.Parameters {
				copyFault.Parameters[key] = value
			}
			active = append(active, copyFault)
		}
	}
	m.mu.RUnlock()

	for _, fault := range active {
		switch fault.Type {
		case "latency":
			if ms, ok := fault.Parameters["latency_ms"].(float64); ok && ms > 0 {
				timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			}
		case "error":
			rate, _ := fault.Parameters["error_rate"].(float64)
			if rate > 0 && randomFloat() < rate {
				code := 500
				if c, ok := fault.Parameters["error_code"].(float64); ok && c >= 500 && c <= 599 {
					code = int(c)
				}
				return &FaultError{Code: code, Msg: "injected error"}
			}
		case "db_failure":
			rate, _ := fault.Parameters["failure_rate"].(float64)
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
		if isActive(f) {
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

			err := m.applyContext(r.Context(), route, func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			if err != nil {
				// Fault injected an error
				if faultErr, ok := err.(interface{ StatusCode() int }); ok {
					response.JSON(w, faultErr.StatusCode(), map[string]string{"error": err.Error()})
				} else {
					response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				}
				return
			}
		})
	}
}

func (m *FaultManager) loadFaults(ctx context.Context) {
	if !m.config.Enabled || m.db == nil {
		return
	}
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
		if json.Unmarshal(paramsJSON, &f.Parameters) != nil || f.TargetRoute != "/api/products" || m.validateParams(f.Type, f.Parameters) != nil || f.ExpiresAt == nil {
			continue
		}
		m.mu.Lock()
		m.faults[f.ID] = &f
		m.mu.Unlock()
	}
}

func (m *FaultManager) persistFault(ctx context.Context, fault *models.Fault) error {
	if m.db == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	paramsJSON, _ := json.Marshal(fault.Parameters)
	_, err := m.db.Pool.Exec(ctx, `
		INSERT INTO faults (id, type, target_route, parameters, enabled, created_at, expires_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			expires_at = EXCLUDED.expires_at
	`, fault.ID, fault.Type, fault.TargetRoute, paramsJSON, fault.Enabled, fault.CreatedAt, fault.ExpiresAt, fault.CreatedBy)
	if err != nil {
		m.logger.Error("persist fault failed")
		return fmt.Errorf("fault persistence failed")
	}
	return nil
}

func (m *FaultManager) cleanupExpired(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
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

func randomFloat() float64 { return rand.Float64() }
func isActive(f *models.Fault) bool {
	return f.Enabled && f.ExpiresAt != nil && time.Now().Before(*f.ExpiresAt)
}
func cloneFault(f *models.Fault) *models.Fault {
	c := *f
	c.Parameters = map[string]interface{}{}
	for k, v := range f.Parameters {
		c.Parameters[k] = v
	}
	return &c
}
