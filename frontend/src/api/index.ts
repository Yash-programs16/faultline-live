// Main API exports - explicitly re-export to avoid conflicts

// Client & errors
export { api, ApiError, type ApiError as ApiErrorType } from './client'

// Products
export {
  fetchProducts,
  fetchProduct,
  fetchCategories,
  fetchBrands,
  type ProductsResponse,
  type ProductFilters,
} from './products'

// Auth
export {
  register,
  login,
  logout,
  fetchMe,
  fetchAddresses,
  createAddress,
  updateAddress,
  deleteAddress,
  fetchWishlist,
  addToWishlist,
  removeFromWishlist,
  setAuthToken,
  getAuthToken,
  type User,
  type Address,
  type RegisterRequest,
  type LoginRequest,
  type AuthResponse,
  type AddressRequest,
} from './auth'

// Orders
export {
  createOrder,
  fetchOrders,
  fetchOrder,
  type CartItem,
  type Address as OrderAddress,
  type CreateOrderRequest,
  type Order,
  type OrderItem,
  type OrdersResponse,
} from './orders'

// Account (re-exports from auth for convenience)
export {
  fetchAddresses as fetchAccountAddresses,
  createAddress as createAccountAddress,
  updateAddress as updateAccountAddress,
  deleteAddress as deleteAccountAddress,
  fetchWishlist as fetchAccountWishlist,
  addToWishlist as addToAccountWishlist,
  removeFromWishlist as removeFromAccountWishlist,
  type AccountOverview,
} from './account'

// Ops
export {
  fetchHealth,
  fetchReady,
  fetchMetrics,
  fetchMetricsHistory,
  fetchOpsOverview,
  fetchIncidents,
  fetchIncident,
  investigateIncident,
  fetchFaults,
  createFault,
  deleteFault,
  type HealthResponse,
  type ReadyResponse,
  type MetricsResponse,
  type MetricSample,
  type MetricsHistoryResponse,
  type ServiceStatus,
  type OpsOverviewResponse,
  type IncidentSummary,
  type IncidentDetail,
  type InvestigationResult,
  type Hypothesis,
  type FaultConfig,
} from './ops'

// Health
export {
  fetchHealth as fetchHealthCheck,
  fetchReady as fetchReadyCheck,
  type HealthResponse as HealthCheckResponse,
  type ReadyResponse as ReadyCheckResponse,
} from './health'