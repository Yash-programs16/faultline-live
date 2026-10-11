import { api, ApiError } from './client'

export interface HealthResponse {
  status: 'ok' | 'degraded' | 'down'
  timestamp: string
  version: string
  uptime: string
}

export interface ReadyResponse {
  ready: boolean
  checks: Record<string, { status: 'ok' | 'fail'; message?: string }>
  timestamp: string
}

export async function fetchHealth(): Promise<HealthResponse> {
  return api.get<HealthResponse>('/api/health')
}

export async function fetchReady(): Promise<ReadyResponse> {
  return api.get<ReadyResponse>('/api/ready')
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}