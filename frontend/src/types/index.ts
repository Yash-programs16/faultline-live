export interface Product {
  id: number;
  name: string;
  description: string;
  price_cents: number;
  stock: number;
  created_at: string;
  updated_at: string;
}

export interface ProductsResponse {
  products: Product[];
}

export interface ProductResponse {
  product: Product;
}

export interface CartItem {
  product_id: number;
  quantity: number;
}

export interface CartValidateRequest {
  items: CartItem[];
}

export interface CartValidateResponse {
  valid: boolean;
  items: CartItemDetail[];
  total_cents: number;
  errors: string[];
}

export interface CartItemDetail {
  product_id: number;
  name: string;
  price_cents: number;
  quantity: number;
  stock: number;
  subtotal_cents: number;
}

export interface CheckoutRequest {
  items: CartItem[];
  payment_method: string;
}

export interface CheckoutResponse {
  success: boolean;
  order_id?: string;
  message: string;
  processing_time_ms?: number;
}

export interface HealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  database: 'connected' | 'disconnected';
  timestamp: string;
}

export interface MetricWindow {
  window_start: string;
  service: string;
  request_count: number;
  error_count: number;
  error_rate: number;
  mean_latency_ms: number;
  p95_latency_ms: number;
  requests_per_second: number;
}

export interface MetricsResponse {
  windows: MetricWindow[];
  current_fault?: FaultState;
}

export interface FaultState {
  scenario: 'latency' | 'failure';
  active: boolean;
  params: Record<string, unknown>;
  activated_at: string;
  expires_at?: string;
}

export interface FaultRequest {
  scenario: 'latency' | 'failure';
  params: {
    delay_ms?: number;
    failure_rate?: number;
    duration_seconds?: number;
  };
}

export interface FaultResponse {
  success: boolean;
  fault?: FaultState;
  message: string;
}

export interface Incident {
  id: number;
  title: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: 'open' | 'investigating' | 'resolved';
  created_at: string;
  resolved_at?: string;
  description: string;
}

export interface IncidentsResponse {
  incidents: Incident[];
}

export interface ApiError {
  error: string;
  message: string;
  status: number;
}