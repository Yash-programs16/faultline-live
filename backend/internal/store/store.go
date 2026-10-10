package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Println("Database connection established")
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
	log.Println("Database connection closed")
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) Migrate(ctx context.Context) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
			stock BIGINT NOT NULL CHECK (stock >= 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_products_name ON products(name)`,
		`CREATE TABLE IF NOT EXISTS telemetry_windows (
			id BIGSERIAL PRIMARY KEY,
			window_start TIMESTAMPTZ NOT NULL,
			service VARCHAR(100) NOT NULL,
			request_count BIGINT NOT NULL DEFAULT 0,
			error_count BIGINT NOT NULL DEFAULT 0,
			error_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
			mean_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
			p95_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
			requests_per_second DOUBLE PRECISION NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(window_start, service)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_windows_window_start ON telemetry_windows(window_start DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_telemetry_windows_service ON telemetry_windows(service)`,
		`CREATE TABLE IF NOT EXISTS fault_events (
			id BIGSERIAL PRIMARY KEY,
			scenario VARCHAR(50) NOT NULL,
			activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deactivated_at TIMESTAMPTZ,
			params JSONB NOT NULL DEFAULT '{}'::jsonb,
			active BOOLEAN NOT NULL DEFAULT TRUE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fault_events_active ON fault_events(active) WHERE active`,
	}

	for i, migration := range migrations {
		if _, err := s.pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	applied, err := s.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	seedMigrations := []struct {
		version int64
		sql     string
	}{
		{
			version: 1,
			sql: `
				INSERT INTO products (name, description, price_cents, stock) VALUES
				('Wireless Headphones', 'Premium noise-cancelling wireless headphones with 30-hour battery life', 29999, 50),
				('Mechanical Keyboard', 'RGB backlit mechanical keyboard with Cherry MX Brown switches', 14999, 30),
				('USB-C Hub', '7-in-1 USB-C hub with HDMI, USB-A, SD card reader, and PD charging', 4999, 100),
				('Ergonomic Mouse', 'Vertical ergonomic mouse with adjustable DPI and wireless connectivity', 3999, 75),
				('Laptop Stand', 'Aluminum adjustable laptop stand with cable management', 2999, 60),
				('Webcam Cover', 'Magnetic sliding webcam cover for privacy (pack of 3)', 999, 200)
				ON CONFLICT DO NOTHING
			`,
		},
	}

	for _, m := range seedMigrations {
		if !applied[m.version] {
			if _, err := s.pool.Exec(ctx, m.sql); err != nil {
				return fmt.Errorf("seed migration %d failed: %w", m.version, err)
			}
			if _, err := s.pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, m.version); err != nil {
				return fmt.Errorf("record seed migration %d: %w", m.version, err)
			}
			log.Printf("Applied seed migration %d", m.version)
		}
	}

	return nil
}

func (s *Store) getAppliedMigrations(ctx context.Context) (map[int64]bool, error) {
	rows, err := s.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[int64]bool{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Stock       int64     `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Store) ListProducts(ctx context.Context) ([]Product, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, description, price_cents, stock, created_at, updated_at
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *Store) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, description, price_cents, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) ValidateCart(ctx context.Context, items []CartItem) (*CartValidationResult, error) {
	if len(items) == 0 {
		return &CartValidationResult{Valid: false, Errors: []string{"cart is empty"}}, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var result CartValidationResult
	result.Items = make([]CartItemDetail, 0, len(items))
	result.TotalCents = 0

	seen := make(map[int64]bool)
	for _, item := range items {
		if item.Quantity <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("product %d: quantity must be positive", item.ProductID))
			continue
		}
		if seen[item.ProductID] {
			result.Errors = append(result.Errors, fmt.Sprintf("product %d: duplicate product ID in cart", item.ProductID))
			continue
		}
		seen[item.ProductID] = true

		var product Product
		err := tx.QueryRow(ctx, `
			SELECT id, name, description, price_cents, stock, created_at, updated_at
			FROM products
			WHERE id = $1
		`, item.ProductID).Scan(&product.ID, &product.Name, &product.Description, &product.PriceCents, &product.Stock, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result.Errors = append(result.Errors, fmt.Sprintf("product %d: not found", item.ProductID))
				continue
			}
			return nil, err
		}

		if product.Stock < item.Quantity {
			result.Errors = append(result.Errors, fmt.Sprintf("product %d: insufficient stock (requested %d, available %d)", item.ProductID, item.Quantity, product.Stock))
			continue
		}

		subtotal := product.PriceCents * item.Quantity
		result.Items = append(result.Items, CartItemDetail{
			ProductID:   product.ID,
			Name:        product.Name,
			PriceCents:  product.PriceCents,
			Quantity:    item.Quantity,
			Stock:       product.Stock,
			SubtotalCents: subtotal,
		})
		result.TotalCents += subtotal
	}

	result.Valid = len(result.Errors) == 0
	return &result, nil
}

func (s *Store) Checkout(ctx context.Context, items []CartItem) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		if item.Quantity <= 0 {
			return "", fmt.Errorf("invalid quantity for product %d", item.ProductID)
		}

		result, err := tx.Exec(ctx, `
			UPDATE products
			SET stock = stock - $2, updated_at = NOW()
			WHERE id = $1 AND stock >= $2
		`, item.ProductID, item.Quantity)
		if err != nil {
			return "", err
		}
		if result.RowsAffected() == 0 {
			return "", fmt.Errorf("insufficient stock for product %d", item.ProductID)
		}
	}

	orderID := fmt.Sprintf("ORD-%d", time.Now().UnixMilli())
	if _, err := tx.Exec(ctx, `INSERT INTO orders (id, created_at) VALUES ($1, NOW())`, orderID); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return orderID, nil
}

var ErrProductNotFound = errors.New("product not found")

type CartItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type CartItemDetail struct {
	ProductID      int64 `json:"product_id"`
	Name           string `json:"name"`
	PriceCents     int64 `json:"price_cents"`
	Quantity       int64 `json:"quantity"`
	Stock          int64 `json:"stock"`
	SubtotalCents  int64 `json:"subtotal_cents"`
}

type CartValidationResult struct {
	Valid       bool             `json:"valid"`
	Items       []CartItemDetail `json:"items"`
	TotalCents  int64            `json:"total_cents"`
	Errors      []string         `json:"errors"`
}

type TelemetryWindow struct {
	WindowStart      time.Time `json:"window_start"`
	Service          string    `json:"service"`
	RequestCount     int64     `json:"request_count"`
	ErrorCount       int64     `json:"error_count"`
	ErrorRate        float64   `json:"error_rate"`
	MeanLatencyMs    float64   `json:"mean_latency_ms"`
	P95LatencyMs     float64   `json:"p95_latency_ms"`
	RequestsPerSecond float64  `json:"requests_per_second"`
}

func (s *Store) PersistTelemetryWindow(ctx context.Context, window TelemetryWindow) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO telemetry_windows (window_start, service, request_count, error_count, error_rate, mean_latency_ms, p95_latency_ms, requests_per_second)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (window_start, service) DO UPDATE SET
			request_count = EXCLUDED.request_count,
			error_count = EXCLUDED.error_count,
			error_rate = EXCLUDED.error_rate,
			mean_latency_ms = EXCLUDED.mean_latency_ms,
			p95_latency_ms = EXCLUDED.p95_latency_ms,
			requests_per_second = EXCLUDED.requests_per_second
	`, window.WindowStart, window.Service, window.RequestCount, window.ErrorCount, window.ErrorRate, window.MeanLatencyMs, window.P95LatencyMs, window.RequestsPerSecond)
	return err
}

func (s *Store) GetRecentTelemetryWindows(ctx context.Context, service string, limit int) ([]TelemetryWindow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT window_start, service, request_count, error_count, error_rate, mean_latency_ms, p95_latency_ms, requests_per_second
		FROM telemetry_windows
		WHERE service = $1
		ORDER BY window_start DESC
		LIMIT $2
	`, service, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []TelemetryWindow
	for rows.Next() {
		var w TelemetryWindow
		if err := rows.Scan(&w.WindowStart, &w.Service, &w.RequestCount, &w.ErrorCount, &w.ErrorRate, &w.MeanLatencyMs, &w.P95LatencyMs, &w.RequestsPerSecond); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

type FaultEvent struct {
	ID            int64           `json:"id"`
	Scenario      string          `json:"scenario"`
	ActivatedAt   time.Time       `json:"activated_at"`
	DeactivatedAt *time.Time      `json:"deactivated_at,omitempty"`
	Params        map[string]any  `json:"params"`
	Active        bool            `json:"active"`
}

func (s *Store) RecordFaultActivation(ctx context.Context, scenario string, params map[string]any) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO fault_events (scenario, params, active)
		VALUES ($1, $2, TRUE)
		RETURNING id
	`, scenario, params).Scan(&id)
	return id, err
}

func (s *Store) RecordFaultDeactivation(ctx context.Context, faultID int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE fault_events
		SET deactivated_at = NOW(), active = FALSE
		WHERE id = $1
	`, faultID)
	return err
}

func (s *Store) GetActiveFault(ctx context.Context) (*FaultEvent, error) {
	var f FaultEvent
	err := s.pool.QueryRow(ctx, `
		SELECT id, scenario, activated_at, deactivated_at, params, active
		FROM fault_events
		WHERE active = TRUE
		ORDER BY activated_at DESC
		LIMIT 1
	`).Scan(&f.ID, &f.Scenario, &f.ActivatedAt, &f.DeactivatedAt, &f.Params, &f.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}