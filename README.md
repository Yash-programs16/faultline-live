# Faultline Live

E-commerce storefront with operational monitoring, incident detection, fault injection, and Gemma-powered incident analysis.

## Architecture

```
faultline-live/
├── frontend/          # React 19 + TypeScript + Vite + Tailwind 4 storefront
├── backend/           # Go + Gorilla Mux + PostgreSQL + Ollama/Gemma
└── README.md
```

## Prerequisites

- **Node.js** 20+ and npm
- **Go** 1.26.1 (matches `backend/go.mod`)
- **PostgreSQL** (Neon hosted - connection string provided)
- **Ollama** with Gemma model (`ollama pull gemma2:2b`)

## Quick Start (Windows PowerShell)

### 1. Backend Setup

```powershell
cd faultline-live\backend

# Copy environment template
Copy-Item .env.example .env

# Edit .env with your credentials (required: DATABASE_URL, JWT_SECRET)
notepad .env

# Install Go dependencies
go mod tidy

# Run database migrations
go run ./cmd/migrate

# Start backend server
go run .
```

### 2. Frontend Setup

```powershell
cd faultline-live\frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

### 3. Ollama/Gemma (for AI investigations)

```powershell
# Install Ollama from https://ollama.ai
# Pull Gemma model
ollama pull gemma2:2b

# Start Ollama service
ollama serve
```

### 4. Verify

- Faultline Dashboard: http://localhost:5173/
- Sample Storefront Workload: http://localhost:5173/store
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/api/health
- Operations Dashboard: http://localhost:5173/ops (local demo; no storefront login)

## Demo Sequence

1. **Open the dashboard** - Start at `http://localhost:5173/`; the sample workload is available at `/store` and `/produtos`
2. **Open the Faultline Operations Dashboard** - Navigate to `/ops` (no customer account needed in local demo mode).
3. **Inject a controlled fault** - Use the Simulation panel to add latency, HTTP 503 errors, or simulated database failures to `/api/products`.
4. **Generate bounded traffic** - Use the dashboard button (30 requests, max 3 concurrent).
5. **Observe telemetry** - Watch latency/error metrics, then allow the detector cycle to run.
6. **Investigate with Gemma** - Open an incident and click "Iniciar Investigação".
7. **Remove the fault** - Clear the active simulation and verify metrics recover.
8. **Optional workload** - Visit the storefront to generate normal product traffic; customer login/checkout is not required for the Faultline demo.

## Environment Variables

### Backend (`.env`)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Neon PostgreSQL connection string |
| `JWT_SECRET` | Yes | 32+ char random string for token signing |
| `OLLAMA_BASE_URL` | No | Default: `http://localhost:11434` |
| `OLLAMA_MODEL` | No | Default: `gemma2:2b` |
| `FAULTS_ENABLED` | No | Enable controlled local fault simulations (set `true` for the demo) |
| `OPS_AUTH_REQUIRED` | No | Require JWT on ops API; for local demo set `false` and bind `SERVER_HOST=127.0.0.1` |
| `TELEMETRY_WINDOW` | No | Use `1m` for fast demo detection feedback |

### Frontend (`.env.local`)

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_BASE_URL` | No | Backend URL (empty = use Vite proxy) |

## API Endpoints

### Health & Metrics
- `GET /api/health` - Liveness check
- `GET /api/ready` - Readiness check (DB, Ollama)
- `GET /api/metrics` - Current metrics
- `GET /api/metrics/history?window=1h` - Historical metrics

### Products
- `GET /api/products` - List with filters (q, category, brands, price, rating, sort, page)
- `GET /api/products/{slug}` - Product detail
- `GET /api/categories` - All categories
- `GET /api/brands` - All brands

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/me` - Current user

### Account
- `GET /api/account/addresses` - List addresses
- `POST /api/account/addresses` - Create address
- `PATCH /api/account/addresses/{id}` - Update address
- `DELETE /api/account/addresses/{id}` - Delete address
- `GET /api/account/wishlist` - Get wishlist product IDs
- `POST /api/account/wishlist/{productId}` - Add to wishlist
- `DELETE /api/account/wishlist/{productId}` - Remove from wishlist

### Orders
- `POST /api/orders` - Create order (validates stock, prices, coupons)
- `GET /api/orders` - List user orders
- `GET /api/orders/{id}` - Order detail

### Operations (JWT-protected by default; local-demo no-login mode requires `OPS_AUTH_REQUIRED=false` plus a loopback bind)
- `GET /api/ops/overview` - Dashboard overview
- `GET /api/ops/services` - Service health
- `GET /api/ops/incidents` - List incidents
- `GET /api/ops/incidents/{id}` - Incident detail
- `POST /api/ops/incidents/{id}/investigate` - Gemma investigation
- `GET /api/ops/faults` - List active faults
- `POST /api/ops/faults` - Create fault
- `DELETE /api/ops/faults/{id}` - Delete fault

## Detection scope

The current detector uses explicit latency/error/database thresholds plus a simple request-volume heuristic; it is not a trained Isolation Forest model or a production DDoS detector. The SQL injection button is a harmless probe against a parameterized search query; it does not pretend to detect arbitrary SQL injection attempts. Automated remediation is not yet claimed by this build.

## Fault Injection

Fault injection is disabled by default in secure deployments. For the local-only hackathon demo, set `SERVER_HOST=127.0.0.1`, `OPS_AUTH_REQUIRED=false`, and `FAULTS_ENABLED=true` in backend `.env`. Never expose unauthenticated operations/fault controls to a network.

### Fault Types

1. **Latency** - Add delay to route
   ```json
   {
     "type": "latency",
     "target_route": "/api/products",
     "parameters": { "latency_ms": 2000 }
   }
   ```

2. **Error** - Inject 5xx errors
   ```json
   {
     "type": "error",
     "target_route": "/api/products",
     "parameters": { "error_rate": 0.3, "error_code": 503 }
   }
   ```

3. **DB Failure** - Simulate DB errors
   ```json
   {
     "type": "db_failure",
     "target_route": "/api/products",
     "parameters": { "failure_rate": 0.5 }
   }
   ```

### Safety Controls
- Max duration: 10 minutes (configurable)
- Max latency: 10 seconds
- Max error rate: 50%
- Auto-expiry and cleanup
- Requires authentication

## Database Schema

Key tables:
- `users` - Authentication
- `categories`, `products`, `product_variants`, `product_specs`, `product_reviews` - Catalog
- `addresses`, `wishlist_items` - Account
- `coupons` - Discounts
- `orders`, `order_items` - Orders
- `incidents`, `investigations` - Observability
- `faults` - Fault injection state
- `metric_samples`, `baselines` - Telemetry
- `audit_logs` - Audit trail

## Project Structure

### Frontend (`frontend/`)
```
src/
├── api/           # Typed API client layer
├── components/    # Reusable UI components
├── pages/         # Route pages (including /ops dashboard)
├── stores/        # Zustand state (auth, cart, orders, wishlist, UI)
├── data/          # Types, mock data (fallback)
└── lib/           # Utilities
```

### Backend (`backend/`)
```
internal/
├── config/        # Configuration loading
├── database/      # pgxpool connection
├── middleware/    # HTTP middleware (auth, logging, metrics, CORS)
├── handlers/      # HTTP handlers (products, auth, orders, ops)
├── repositories/  # Database queries
├── models/        # Data structures
├── telemetry/     # Metrics collection & aggregation
├── incidents/     # Anomaly detection
├── gemma/         # Ollama/Gemma integration
├── faults/        # Fault injection
└── response/      # HTTP response helpers
migrations/        # SQL migrations
```

## Testing

### Backend
```powershell
cd faultline-live\backend
go test ./...
```

### Frontend
```powershell
cd faultline-live\frontend
npm run build
```

## Production Considerations

This is a **hackathon demo** - not production-ready. Known limitations:

- In-memory incident/metrics state (lost on restart)
- No distributed tracing
- Simple in-memory rate limiting
- JWT tokens not revoked on logout
- No HTTPS/TLS termination in Go (use reverse proxy)
- Single-instance fault injection state

## License

MIT