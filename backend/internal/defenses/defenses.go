package defenses

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/response"

	"github.com/gorilla/mux"
)

// Defense is a protective control placed in front of an endpoint.
// For this demo the only enforced type is "rate_limit"; the other
// suggestion types are advisory and map to existing remediations.
type Defense struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	Target           string    `json:"target"`
	RequestsPerSec   float64   `json:"requests_per_second"`
	Burst            int       `json:"burst"`
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by"`
	BlockedRequests  int64     `json:"blocked_requests"`
}

// Suggestion is a per-attack-type recommended defense shown on the
// incident page. Action is one of: enable_rate_limit, stop_fault,
// disable_defense.
type Suggestion struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Action      string  `json:"action"`
	Target      string  `json:"target"`
	RPS         float64 `json:"requests_per_second,omitempty"`
	Burst       int     `json:"burst,omitempty"`
	Active      bool    `json:"active"`
	Primary     bool    `json:"primary"`
}

type Manager struct {
	mu       sync.Mutex
	defenses map[string]*Defense
	hits     map[string][]time.Time // rate-limit sliding windows per ip+target
}

func NewManager() *Manager {
	return &Manager{
		defenses: make(map[string]*Defense),
		hits:     make(map[string][]time.Time),
	}
}

func (m *Manager) List() []Defense {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Defense, 0, len(m.defenses))
	for _, d := range m.defenses {
		c := *d
		out = append(out, c)
	}
	return out
}

func (m *Manager) GetByTarget(target string) *Defense {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := normalizeTarget(target)
	for _, d := range m.defenses {
		if d.Enabled && normalizeTarget(d.Target) == n {
			c := *d
			return &c
		}
	}
	return nil
}

// EnableRateLimit creates or updates a rate-limit defense for a target.
func (m *Manager) EnableRateLimit(target string, rps float64, burst int, createdBy string) (*Defense, error) {
	if normalizeTarget(target) == "" {
		return nil, fmt.Errorf("target is required")
	}
	if rps <= 0 || rps > 1000 {
		return nil, fmt.Errorf("requests_per_second must be between 1 and 1000")
	}
	if burst <= 0 {
		burst = int(rps)
	}
	if burst > 1000 {
		burst = 1000
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n := normalizeTarget(target)
	for _, d := range m.defenses {
		if normalizeTarget(d.Target) == n && d.Type == "rate_limit" {
			d.RequestsPerSec = rps
			d.Burst = burst
			d.Enabled = true
			d.CreatedBy = createdBy
			c := *d
			return &c, nil
		}
	}
	d := &Defense{
		ID:             models.NewFaultID(),
		Type:           "rate_limit",
		Target:         n,
		RequestsPerSec: rps,
		Burst:          burst,
		Enabled:        true,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}
	m.defenses[d.ID] = d
	c := *d
	return &c, nil
}

func (m *Manager) Disable(idOrTarget string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.defenses[idOrTarget]; ok {
		d.Enabled = false
		delete(m.defenses, idOrTarget)
		return nil
	}
	n := normalizeTarget(idOrTarget)
	for id, d := range m.defenses {
		if normalizeTarget(d.Target) == n {
			delete(m.defenses, id)
			return nil
		}
	}
	return fmt.Errorf("defense not found")
}

// Allow reports whether a request from ip to endpoint may proceed.
// It returns false when an active rate-limit defense is exhausted.
func (m *Manager) Allow(ip, endpoint string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	def := m.matchLocked(endpoint)
	if def == nil {
		return true
	}
	limit := def.Burst
	if limit <= 0 {
		limit = int(def.RequestsPerSec)
	}
	window := time.Second
	// Low RPS values (e.g. 2/s) still use a 1s window; burst absorbs spikes.
	key := ip + "|" + normalizeTarget(def.Target)
	now := time.Now()
	cutoff := now.Add(-window)
	kept := m.hits[key][:0]
	for _, t := range m.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= limit {
		m.hits[key] = kept
		def.BlockedRequests++
		return false
	}
	m.hits[key] = append(kept, now)
	return true
}

func (m *Manager) matchLocked(endpoint string) *Defense {
	ep := normalizeTarget(endpoint)
	for _, d := range m.defenses {
		if !d.Enabled {
			continue
		}
		t := normalizeTarget(d.Target)
		if t == ep || t == "global" || ep == t {
			return d
		}
		// "GET /api/products" should protect "/api/products" and vice versa.
		if strings.HasSuffix(ep, t) || strings.HasSuffix(t, ep) {
			return d
		}
	}
	return nil
}

// Middleware enforces active rate-limit defenses. It runs before fault
// injection so abusive floods are shed with 429 instead of consuming
// backend work.
func (m *Manager) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := r.URL.Path
			if route := mux.CurrentRoute(r); route != nil {
				if pattern, err := route.GetPathTemplate(); err == nil && pattern != "" {
					endpoint = pattern
				}
			}
			ip := r.RemoteAddr
			if i := strings.LastIndex(ip, ":"); i >= 0 {
				ip = ip[:i]
			}
			if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
				parts := strings.Split(xf, ",")
				if strings.TrimSpace(parts[0]) != "" {
					ip = strings.TrimSpace(parts[0])
				}
			}
			if !m.Allow(ip, endpoint) {
				w.Header().Set("Retry-After", "1")
				response.TooManyRequests(w, "rate limit defense is shedding excess requests for "+endpoint)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SuggestForIncident returns ordered defense options for an incident type.
// Every attack type gets at least a rate-limiter option plus the
// fault-removal option when the incident maps to an injectable fault.
func SuggestForIncident(incidentType, affectedEndpoint string, activeDefenses []Defense, hasOverlappingFault bool) []Suggestion {
	target := affectedEndpoint
	if strings.HasPrefix(target, "GET ") {
		target = strings.TrimPrefix(target, "GET ")
	}
	if target == "global" || target == "database" || target == "" {
		target = "/api/products"
	}
	active := false
	var rps float64
	var burst int
	for _, d := range activeDefenses {
		if d.Enabled && (normalizeTarget(d.Target) == normalizeTarget(target) || d.Target == "global") {
			active = true
			rps = d.RequestsPerSec
			burst = d.Burst
		}
	}
	rateDesc := "Cap requests per client with 429 shedding. Primary defense for HTTP floods (traffic_spike) and a safety net for error/latency storms."
	sugs := []Suggestion{
		{
			ID: "rate_limit", Label: "Rate limiter",
			Description: rateDesc,
			Action:      "enable_rate_limit", Target: target,
			RPS: defaultRPS(incidentType, rps), Burst: defaultBurst(incidentType, burst),
			Active: active, Primary: incidentType == "traffic_spike",
		},
	}
	if hasOverlappingFault || incidentType == "high_error_rate" || incidentType == "high_latency" || incidentType == "high_db_latency" {
		sugs = append(sugs, Suggestion{
			ID: "stop_fault", Label: "Stop injected fault",
			Description: "Removes the overlapping latency/error/db_failure simulation for this endpoint.",
			Action:      "stop_fault", Target: target, Primary: incidentType != "traffic_spike",
		})
	}
	if active {
		sugs = append(sugs, Suggestion{
			ID: "disable_defense", Label: "Disable rate limiter",
			Description: "Removes throttling once telemetry is healthy for two detector cycles.",
			Action:      "disable_defense", Target: target,
		})
	}
	// Order primary first.
	if len(sugs) == 3 && sugs[0].Primary == false && sugs[1].Primary {
		sugs[0], sugs[1] = sugs[1], sugs[0]
	}
	return sugs
}

func defaultRPS(incidentType string, current float64) float64 {
	if current > 0 {
		return current
	}
	switch incidentType {
	case "traffic_spike":
		return 10
	case "high_error_rate":
		return 20
	case "high_latency", "high_db_latency":
		return 15
	default:
		return 20
	}
}

func defaultBurst(incidentType string, current int) int {
	if current > 0 {
		return current
	}
	switch incidentType {
	case "traffic_spike":
		return 20
	default:
		return 40
	}
}

func normalizeTarget(t string) string {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "GET ")
	t = strings.TrimSpace(t)
	return t
}
