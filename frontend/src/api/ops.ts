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

export interface MetricsResponse {
  requestCount: number
  requestRate: number
  errorCount: number
  errorRate: number
  latencyP50: number
  latencyP95: number
  latencyP99: number
  dbLatencyP50: number
  dbLatencyP95: number
  activeFaults: number
  timestamp: string
}

export interface MetricSample {
  timestamp: string
  requestCount: number
  errorCount: number
  latencyP50: number
  latencyP95: number
  latencyP99: number
  dbLatencyP50: number
  dbLatencyP95: number
  errorRate?: number
}

export interface MetricsHistoryResponse {
  samples: MetricSample[]
  window: string
}

export interface ServiceStatus {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  latencyMs?: number
  errorRate?: number
  lastCheck: string
}

export interface OpsOverviewResponse {
  services: ServiceStatus[]
  metrics: MetricsResponse
  incidents: IncidentSummary[]
}

export interface IncidentSummary {
  id: string
  type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  affectedEndpoint: string
  startedAt: string
  resolvedAt?: string
  status: 'firing' | 'resolved'
}

export interface IncidentDetail extends IncidentSummary {
  description: string
  observed: Record<string, unknown>
  baseline: Record<string, unknown>
  threshold: Record<string, unknown>
  evidence: string[]
  investigation?: InvestigationResult
}

export interface InvestigationResult {
  id: string
  incidentId: string
  summary: string
  rootCause: string
  hypotheses: Hypothesis[]
  missingEvidence: string[]
  recommendedSteps: string[]
  recoveryRecommendations: string[]
  confidence: number
  confidenceExplanation: string
  createdAt: string
  model: string
}

export interface Hypothesis {
  rank: number
  hypothesis: string
  likelihood: number
  supportingEvidence: string[]
  contradictingEvidence: string[]
}

export interface FaultConfig {
  id: string
  type: 'latency' | 'error' | 'db_failure'
  targetRoute: string
  parameters: Record<string, unknown>
  enabled: boolean
  createdAt: string
  expiresAt?: string
  createdBy: string
}

export async function fetchHealth(): Promise<HealthResponse> {
  return api.get<HealthResponse>('/api/health')
}

export async function fetchReady(): Promise<ReadyResponse> {
  return api.get<ReadyResponse>('/api/ready')
}

export async function fetchMetrics(): Promise<MetricsResponse> {
  return api.get<MetricsResponse>('/api/metrics')
}

export async function fetchMetricsHistory(window = '1h'): Promise<MetricsHistoryResponse> {
  return api.get<MetricsHistoryResponse>(`/api/metrics/history?window=${encodeURIComponent(window)}`)
}

export async function fetchOpsOverview(): Promise<OpsOverviewResponse> {
  return api.get<OpsOverviewResponse>('/api/ops/overview')
}

export async function fetchIncidents(): Promise<IncidentSummary[]> {
  return api.get<IncidentSummary[]>('/api/ops/incidents')
}

export async function fetchIncident(id: string): Promise<IncidentDetail> {
  return api.get<IncidentDetail>(`/api/ops/incidents/${encodeURIComponent(id)}`)
}

export async function investigateIncident(id: string): Promise<InvestigationResult> {
  return api.post<InvestigationResult>(`/api/ops/incidents/${encodeURIComponent(id)}/investigate`, {})
}

export async function fetchFaults(): Promise<FaultConfig[]> {
  return api.get<FaultConfig[]>('/api/ops/faults')
}

export async function createFault(fault: Omit<FaultConfig, 'id' | 'createdAt' | 'createdBy'>): Promise<FaultConfig> {
  return api.post<FaultConfig>('/api/ops/faults', fault)
}

export async function deleteFault(id: string): Promise<void> {
  return api.delete<void>(`/api/ops/faults/${encodeURIComponent(id)}`)
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}