# Faultline Live

A small e-commerce-style web application with a live operational dashboard demonstrating real HTTP request telemetry, controlled fault injection, and observable metric changes.

## Architecture

```
faultline-live/
├── frontend/          # React + TypeScript + Vite
├── backend/           # Go + Gorilla Mux + pgx/v5
│   ├── cmd/api/       # Application entry point
│   ├── internal/
│   │   ├── config/    # Environment configuration
│   │   ├── store/     # PostgreSQL connection, migrations, queries
│   │   ├── api/       # HTTP handlers, routes, middleware
│   │   ├── telemetry/ # Request middleware, aggregation, buffering
│   │   └── faults/    # Payment fault simulator and controller
│   └── migrations/    # SQL migrations
└── tests/             # Integration tests
```

## Prerequisites

- **Node.js** 18+ and npm
- **Go** 1.22+
- **Neon PostgreSQL** account (free tier works)

## Quick Start

### 1. Set up Neon PostgreSQL

1. Create a free account at [Neon](https://console.neon.tech/)
2. Create a new project
3. Copy the connection string (looks like `postgresql://user:pass@ep-xxx.region.aws.neon.tech/dbname?sslmode=require`)

### 2. Configure Backend

```bash
cd backend
cp .env.example .env
# Edit .env and paste your DATABASE_URL
```

### 3. Configure Frontend

```bash
cd frontend
cp .env.example .env.local
# Edit .env.local if your backend runs on a different URL
```

### 4. Install Dependencies

```bash
# Backend
cd backend
go mod download

# Frontend
cd ../frontend
npm install
```

### 5. Run the Application

**Terminal 1 - Backend:**
```bash
cd backend
go run ./cmd/api
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

### 6. Access the Application

- **Storefront:** http://localhost:5173
- **Dashboard:** http://localhost:5173/dashboard
- **API Health:** http://localhost:8080/api/health

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check (app + DB) |
| GET | `/api/products` | List all products |
| GET | `/api/products/{id}` | Get product details |
| POST | `/api/cart/validate` | Validate cart items |
| POST | `/api/checkout` | Simulated checkout |
| GET | `/api/metrics/live` | Real-time telemetry windows |
| GET | `/api/incidents` | Incident list (empty in Phase 2) |
| POST | `/api/demo/fault` | Activate fault injection |
| DELETE | `/api/demo/fault` | Deactivate fault injection |

## Fault Injection

The dashboard includes controls for two bounded fault scenarios:

### Scenario A: Payment Latency
- Adds configurable delay (100ms - 10s) to payment processing
- Uses context-aware waiting
- Automatically expires after configured duration (10s - 5min)

### Scenario B: Payment Failures
- Causes configurable percentage of payments to fail (1% - 100%)
- Deterministic behavior using seeded random generator
- Automatically expires after configured duration (10s - 5min)

**Security:** Fault controls are restricted to localhost by default (`FAULT_LOCAL_ONLY=true`). For non-localhost deployments, explicit admin authorization must be configured.

## Dashboard Features

The operational dashboard (`/dashboard`) displays real-time metrics polled every 2 seconds:

- **Checkout Latency** (mean & p95)
- **Payment Service Latency**
- **Error Rate** (%)
- **Request Rate** (req/s)
- **Live Charts** for latency, error rate, and request rate
- **Fault Injection Controls** with safety limits
- **Current Fault Status** with activation/expiry times

## Testing the End-to-End Flow

1. Start both frontend and backend
2. Browse products on the storefront
3. Add items to cart and proceed to checkout
4. Open the dashboard - observe baseline metrics
5. Activate "Payment Latency" fault (e.g., 2000ms delay for 60s)
6. Place more orders - observe increased latency in charts
7. Deactivate fault or wait for expiry
8. Place more orders - observe recovery to baseline
9. Repeat with "Payment Failure" fault (e.g., 50% failure rate)
10. Observe error rate spike and recovery

## Project Structure Details

### Backend

- **cmd/api/main.go**: Startup, DI, server lifecycle, graceful shutdown
- **internal/config**: Environment validation, no secrets in logs
- **internal/store**: pgxpool with sensible limits, parameterized SQL, migrations, idempotent seeding
- **internal/api**: Gorilla Mux routes, consistent JSON responses, telemetry middleware
- **internal/telemetry**: 5-second aggregation windows, p95 from distribution, bounded in-memory buffer, DB persistence
- **internal/faults**: Thread-safe fault controller, auto-expiry, validated parameters, local-only by default

### Frontend

- **src/pages**: Home (product grid), ProductDetail, Cart, Checkout, Dashboard
- **src/lib/api.ts**: Centralized HTTP client with timeouts, error handling, TypeScript types
- **src/components**: Reusable UI components (cards, charts, forms, toasts)
- **src/types**: Shared TypeScript interfaces

## Environment Variables

### Backend (.env)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| DATABASE_URL | Yes | - | Neon PostgreSQL connection string |
| SERVER_HOST | No | localhost | Server bind address |
| SERVER_PORT | No | 8080 | Server port |
| TELEMETRY_WINDOW_SIZE | No | 5s | Aggregation window duration |
| TELEMETRY_BUFFER_SIZE | No | 120 | In-memory window buffer size |
| TELEMETRY_PERSIST_ENABLED | No | true | Persist windows to PostgreSQL |
| FAULT_MAX_LATENCY_MS | No | 10000 | Maximum allowed latency injection |
| FAULT_MAX_FAILURE_RATE | No | 1.0 | Maximum allowed failure rate |
| FAULT_MAX_DURATION_SECONDS | No | 300 | Maximum fault duration |
| FAULT_LOCAL_ONLY | No | true | Restrict fault controls to localhost |

### Frontend (.env.local)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| VITE_API_BASE_URL | No | http://localhost:8080 | Backend API base URL |

## Known Limitations

- No authentication/authorization (Phase 2 scope)
- Incidents endpoint returns empty list (Phase 3+)
- No Redis, Kafka, Kubernetes, Prometheus, Grafana
- No ML/Isolation Forest (Phase 3+)
- Single backend instance, no horizontal scaling
- SQLite not supported - requires Neon PostgreSQL

## Development

### Run Tests

```bash
# Backend tests
cd backend
go test -race ./...

# Frontend tests (when added)
cd frontend
npm test
```

### Lint

```bash
# Backend
cd backend
go vet ./...
gofmt -l .

# Frontend
cd frontend
npm run lint
```

## Phase Gates

### Phase 0 ✓
- Frontend starts (Vite + React + TypeScript)
- Go dependencies resolve
- Git initialized
- No secrets tracked

### Phase 1 ✓
- Products retrieved from Neon through Go
- Product details work
- Invalid product IDs return 404
- Cart validation handles valid/invalid quantities
- Simulated checkout works
- Database queries use parameters
- Frontend errors are understandable

### Phase 2 ✓
- Real telemetry collection via middleware
- 5-second aggregation windows with p95
- Bounded payment latency injection
- Bounded payment failure injection
- Safe fault controls (localhost only, validated params, auto-expiry)
- Dashboard polling with live charts
- Recovery observable after fault removal

## Next Steps (Phase 3+)

- ML-based anomaly detection (Isolation Forest)
- Root cause analysis
- Kubernetes deployment
- Prometheus/Grafana integration
- Authentication & authorization
- Multi-service architecture

## License

MIT