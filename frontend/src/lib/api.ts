/// <reference types="vite/client" />

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

const DEFAULT_TIMEOUT = 10000;
const CHECKOUT_TIMEOUT = 15000;

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit & { timeout?: number } = {}
): Promise<T> {
  const controller = new AbortController();
  const timeout = options.timeout ?? DEFAULT_TIMEOUT;
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      signal: controller.signal,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      let errorData: { error?: string; message?: string } = {};
      try {
        errorData = await response.json();
      } catch {
        // Ignore parse errors
      }
      throw new ApiError(
        errorData.message || errorData.error || `HTTP ${response.status}`,
        response.status,
        errorData
      );
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  } catch (error) {
    clearTimeout(timeoutId);
    if (error instanceof ApiError) {
      throw error;
    }
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError('Request timeout', 408);
    }
    throw new ApiError('Network error', 0, error);
  }
}

export const api = {
  get: <T>(endpoint: string, timeout?: number) => request<T>(endpoint, { method: 'GET', timeout }),
  post: <T>(endpoint: string, body: unknown, timeout?: number) =>
    request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
      timeout,
    }),
  delete: <T>(endpoint: string, timeout?: number) => request<T>(endpoint, { method: 'DELETE', timeout }),
};

import type { Product, CartItem, CartValidateResponse, CheckoutResponse, HealthResponse, MetricsResponse, FaultResponse, IncidentsResponse } from '../types';

export const productsApi = {
  list: () => api.get<{ products: Product[] }>('/api/products'),
  get: (id: number) => api.get<{ product: Product }>(`/api/products/${id}`),
};

export const cartApi = {
  validate: (items: CartItem[]) =>
    api.post<CartValidateResponse>('/api/cart/validate', { items }),
};

export const checkoutApi = {
  checkout: (items: CartItem[], paymentMethod: string) =>
    api.post<CheckoutResponse>('/api/checkout', { items, payment_method: paymentMethod }, CHECKOUT_TIMEOUT),
};

export const healthApi = {
  check: () => api.get<HealthResponse>('/api/health'),
};

export const metricsApi = {
  live: () => api.get<MetricsResponse>('/api/metrics/live'),
};

export const faultsApi = {
  create: (scenario: 'latency' | 'failure', params: Record<string, unknown>) =>
    api.post<FaultResponse>('/api/demo/fault', { scenario, params }),
  remove: () => api.delete<FaultResponse>('/api/demo/fault'),
};

export const incidentsApi = {
  list: () => api.get<IncidentsResponse>('/api/incidents'),
};