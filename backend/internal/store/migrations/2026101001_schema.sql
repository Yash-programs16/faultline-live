
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
    stock BIGINT NOT NULL CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_name
    ON products(name);

CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    order_id TEXT NOT NULL REFERENCES orders(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    PRIMARY KEY (order_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id
    ON order_items(order_id);

CREATE TABLE IF NOT EXISTS telemetry_windows (
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
    UNIQUE (window_start, service)
);

CREATE INDEX IF NOT EXISTS idx_telemetry_windows_window_start
    ON telemetry_windows(window_start DESC);

CREATE INDEX IF NOT EXISTS idx_telemetry_windows_service
    ON telemetry_windows(service);

CREATE TABLE IF NOT EXISTS fault_events (
    id BIGSERIAL PRIMARY KEY,
    scenario VARCHAR(50) NOT NULL,
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at TIMESTAMPTZ,
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_fault_events_active
    ON fault_events(active) WHERE active;
