package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	PasswordHash string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`        // minor units (cents)
	PriceOld    *int64    `json:"price_old,omitempty"`
	CategoryID  int64     `json:"category_id"`
	CategorySlug string   `json:"category_slug"`
	Brand       string    `json:"brand"`
	Rating      float64   `json:"rating"`
	ReviewsCount int      `json:"reviews_count"`
	Stock       int       `json:"stock"`
	Featured    bool      `json:"featured"`
	Bestseller  bool      `json:"bestseller"`
	IsNew       bool      `json:"is_new"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductVariant struct {
	ProductID int64    `json:"product_id"`
	Colors    []string `json:"colors"`
	Sizes     []string `json:"sizes"`
}

type ProductSpec struct {
	ProductID int64  `json:"product_id"`
	Label     string `json:"label"`
	Value     string `json:"value"`
}

type ProductReview struct {
	ProductID int64     `json:"product_id"`
	Author    string    `json:"author"`
	Rating    int       `json:"rating"`
	Date      time.Time `json:"date"`
	Comment   string    `json:"comment"`
}

type Address struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Label       string    `json:"label"`
	Recipient   string    `json:"recipient"`
	Street      string    `json:"street"`
	Number      string    `json:"number"`
	Complement  *string   `json:"complement,omitempty"`
	District    string    `json:"district"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Zip         string    `json:"zip"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

type WishlistItem struct {
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Coupon struct {
	Code     string    `json:"code"`
	Percent  float64   `json:"percent"` // 0.1 = 10%
	ValidFrom time.Time `json:"valid_from"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
	Active   bool      `json:"active"`
}

type Order struct {
	ID              string    `json:"id"`
	UserID          int64     `json:"user_id"`
	Subtotal        int64     `json:"subtotal"`
	Discount        int64     `json:"discount"`
	Shipping        int64     `json:"shipping"`
	Total           int64     `json:"total"`
	PaymentMethod   string    `json:"payment_method"`
	Status          string    `json:"status"` // placed, paid, shipped, delivered
	ShippingAddressID int64   `json:"shipping_address_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        int64     `json:"id"`
	OrderID   string    `json:"order_id"`
	ProductID int64     `json:"product_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	Image     string    `json:"image"`
	Color     *string   `json:"color,omitempty"`
	Size      *string   `json:"size,omitempty"`
	Qty       int       `json:"qty"`
}

type Incident struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Severity        string    `json:"severity"` // low, medium, high, critical
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	AffectedEndpoint string   `json:"affected_endpoint"`
	Observed        map[string]interface{} `json:"observed"`
	Baseline        map[string]interface{} `json:"baseline"`
	Threshold       map[string]interface{} `json:"threshold"`
	Evidence        []string  `json:"evidence"`
	Status          string    `json:"status"` // firing, resolved
	StartedAt       time.Time `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	Investigation   *Investigation `json:"investigation,omitempty"`
}

func (i *Incident) ResolvedAtFormat() string {
	if i.ResolvedAt != nil {
		return i.ResolvedAt.Format(time.RFC3339)
	}
	return ""
}

type Investigation struct {
	ID                  string    `json:"id"`
	IncidentID          string    `json:"incident_id"`
	Summary             string    `json:"summary"`
	RootCause           string    `json:"root_cause"`
	Hypotheses          []Hypothesis `json:"hypotheses"`
	MissingEvidence     []string  `json:"missing_evidence"`
	RecommendedSteps    []string  `json:"recommended_steps"`
	RecoveryRecommendations []string `json:"recovery_recommendations"`
	Confidence          float64   `json:"confidence"`
	ConfidenceExplanation string   `json:"confidence_explanation"`
	Model               string    `json:"model"`
	CreatedAt           time.Time `json:"created_at"`
}

type Hypothesis struct {
	Rank                 int      `json:"rank"`
	Hypothesis           string   `json:"hypothesis"`
	Likelihood           float64  `json:"likelihood"`
	SupportingEvidence   []string `json:"supporting_evidence"`
	ContradictingEvidence []string `json:"contradicting_evidence"`
}

type Fault struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"` // latency, error, db_failure
	TargetRoute  string                 `json:"target_route"`
	Parameters   map[string]interface{} `json:"parameters"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	CreatedBy    string                 `json:"created_by"`
}

type MetricSample struct {
	Timestamp       time.Time `json:"timestamp"`
	RequestCount    int64     `json:"request_count"`
	ErrorCount      int64     `json:"error_count"`
	LatencyP50      float64   `json:"latency_p50"`
	LatencyP95      float64   `json:"latency_p95"`
	LatencyP99      float64   `json:"latency_p99"`
	DBLatencyP50    float64   `json:"db_latency_p50"`
	DBLatencyP95    float64   `json:"db_latency_p95"`
}

// API Response types

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
}

type ReadyResponse struct {
	Ready   bool   `json:"ready"`
	Checks  map[string]struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	} `json:"checks"`
	Timestamp string `json:"timestamp"`
}

type MetricsResponse struct {
	RequestCount   int64   `json:"request_count"`
	RequestRate    float64 `json:"request_rate"`
	ErrorCount     int64   `json:"error_count"`
	ErrorRate      float64 `json:"error_rate"`
	LatencyP50     float64 `json:"latency_p50"`
	LatencyP95     float64 `json:"latency_p95"`
	LatencyP99     float64 `json:"latency_p99"`
	DBLatencyP50   float64 `json:"db_latency_p50"`
	DBLatencyP95   float64 `json:"db_latency_p95"`
	ActiveFaults   int     `json:"active_faults"`
	Timestamp      string  `json:"timestamp"`
}

type MetricsHistoryResponse struct {
	Samples []MetricSample `json:"samples"`
	Window  string         `json:"window"`
}

type ServiceStatus struct {
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	ErrorRate float64 `json:"error_rate,omitempty"`
	LastCheck string  `json:"last_check"`
}

type OpsOverviewResponse struct {
	Services   []ServiceStatus     `json:"services"`
	Metrics    MetricsResponse     `json:"metrics"`
	Incidents  []IncidentSummary   `json:"incidents"`
}

type IncidentSummary struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Severity         string `json:"severity"`
	Title            string `json:"title"`
	AffectedEndpoint string `json:"affected_endpoint"`
	StartedAt        string `json:"started_at"`
	ResolvedAt       string `json:"resolved_at,omitempty"`
	Status           string `json:"status"`
}

type IncidentDetail struct {
	IncidentSummary
	Description     string                 `json:"description"`
	Observed        map[string]interface{} `json:"observed"`
	Baseline        map[string]interface{} `json:"baseline"`
	Threshold       map[string]interface{} `json:"threshold"`
	Evidence        []string               `json:"evidence"`
	Investigation   *Investigation         `json:"investigation,omitempty"`
}

type InvestigationRequest struct {
	Incident        Incident
	EndpointMetrics map[string]MetricsResponse
	ActiveFaults    []Fault
	ServiceHealth   []ServiceHealth
}

type ServiceHealth struct {
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	ErrorRate float64 `json:"error_rate,omitempty"`
}

func NewIncidentID() string {
	return "INC-" + uuid.New().String()[:8]
}

func NewFaultID() string {
	return "FLT-" + uuid.New().String()[:8]
}

func NewOrderID() string {
	return "LJ-" + uuid.New().String()[:6]
}

func NewInvestigationID() string {
	return "INV-" + uuid.New().String()[:8]
}