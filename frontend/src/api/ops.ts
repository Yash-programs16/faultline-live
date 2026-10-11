import { api, ApiError, normalizeApiResponse } from './client'

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
  dbObservationCount: number
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
  status: 'healthy' | 'degraded' | 'down' | 'unknown'
  latencyMs?: number
  errorRate?: number
  lastCheck: string
}

export interface OpsOverviewResponse {
 demoMode: boolean
 faultsEnabled: boolean
 metricsWindowSeconds: number
 model: string
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
  actions: Remediation[]
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
  try { return await api.get<ReadyResponse>('/api/ready') } catch (error) {
 if (error instanceof ApiError && error.status === 503 && error.data && typeof error.data === 'object' && 'checks' in error.data) return normalizeApiResponse(error.data) as ReadyResponse
 throw error
 }
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
  return api.post<FaultConfig>('/api/ops/faults', {
    type: fault.type,
    target_route: fault.targetRoute,
    parameters: fault.parameters,
  })
}

export async function deleteFault(id: string): Promise<void> {
  return api.delete<void>(`/api/ops/faults/${encodeURIComponent(id)}`)
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}
export interface TrafficRun {
 id: string; status: 'running' | 'completed' | 'stopped'; target: string; limit: number; concurrency: number
 success: number; failed: number; startedAt: string; finishedAt?: string; latenciesMs: number[]
}
export interface Remediation {
 id: string; incidentId: string; faultId: string; action: string; reason: string; status: string; createdAt: string; recoveryVerified: boolean
}
export function remediateIncident(id: string, faultId: string, reason: string) {
  return api.post<{ action: Remediation; recovery: string }>(`/api/ops/incidents/${encodeURIComponent(id)}/remediate`, { action: 'stop_fault', fault_id: faultId, reason })
}

export interface Defense {
  id: string
  type: string
  target: string
  requestsPerSecond: number
  burst: number
  enabled: boolean
  createdAt: string
  createdBy: string
  blockedRequests: number
}

export interface DefenseSuggestion {
  id: string
  label: string
  description: string
  action: 'enable_rate_limit' | 'stop_fault' | 'disable_defense'
  target: string
  requestsPerSecond?: number
  burst?: number
  active: boolean
  primary: boolean
}

export interface IncidentDefensesResponse {
  incidentId: string
  type: string
  target: string
  startedAt: string
  defenses: Defense[]
  suggestions: DefenseSuggestion[]
}

export async function fetchIncidentDefenses(id: string): Promise<IncidentDefensesResponse> {
  return api.get<IncidentDefensesResponse>(`/api/ops/incidents/${encodeURIComponent(id)}/defenses`)
}

export async function fetchDefenses(): Promise<Defense[]> {
  return api.get<Defense[]>('/api/ops/defenses')
}

export async function createDefense(input: { target: string; requestsPerSecond: number; burst?: number }): Promise<Defense> {
  return api.post<Defense>('/api/ops/defenses', {
    type: 'rate_limit',
    target: input.target,
    requests_per_second: input.requestsPerSecond,
    burst: input.burst ?? Math.round(input.requestsPerSecond * 2),
  })
}

export async function deleteDefense(idOrTarget: string): Promise<void> {
  return api.delete<void>(`/api/ops/defenses/${encodeURIComponent(idOrTarget)}`)
}

export function remediateWithDefense(
  id: string,
  body: { action: 'enable_rate_limit' | 'disable_defense'; target?: string; requestsPerSecond?: number; burst?: number; reason: string },
) {
  return api.post<{ action: Remediation; defense?: Defense; recovery: string }>(
    `/api/ops/incidents/${encodeURIComponent(id)}/remediate`,
    {
      action: body.action,
      target: body.target,
      requests_per_second: body.requestsPerSecond,
      burst: body.burst,
      reason: body.reason,
    },
  )
}

/** Last-resort hardcoded analysis so the UI never shows only "failed".
 * Mirrors the backend fallback: WHAT failed and WHEN. */
export function buildLocalFallback(incident: IncidentDetail): InvestigationResult {
  const when = incident.startedAt
  const what =
    incident.type === 'traffic_spike'
      ? 'Unusual HTTP request volume (possible flood)'
      : incident.type === 'high_error_rate'
        ? 'Elevated HTTP 5xx error rate'
        : incident.type === 'high_latency'
          ? 'Elevated API latency (P95)'
          : incident.type === 'high_db_latency'
            ? 'Elevated database latency (P95)'
            : incident.title || `Anomaly (${incident.type})`
  return {
    id: `LOCAL-${incident.id}`,
    incidentId: incident.id,
    summary: `${what} detected on ${incident.affectedEndpoint} at ${when} (severity: ${incident.severity}).`,
    rootCause: `${what} on ${incident.affectedEndpoint} starting ${when}. Check the observed-vs-baseline gap and any overlapping fault below.`,
    hypotheses: [
      {
        rank: 1,
        hypothesis: `${what} on ${incident.affectedEndpoint} starting ${when}.`,
        likelihood: 0.6,
        supportingEvidence: incident.evidence.length > 0 ? incident.evidence : [`Incident fired at ${when}.`],
        contradictingEvidence: [],
      },
    ],
    missingEvidence: ['Backend unreachable: per-client breakdown and traces unavailable.'],
    recommendedSteps: [
      `Confirm what failed: compare observed vs baseline at ${when}.`,
      'For HTTP floods, enable the rate-limiter defense on this page.',
    ],
    recoveryRecommendations: ['Apply the suggested defense below, then verify two healthy detector cycles.'],
    confidence: 0.5,
    confidenceExplanation: 'Local offline analysis: backend investigation endpoint unreachable.',
    createdAt: new Date().toISOString(),
    model: 'fallback-local (offline)',
  }
}

export function isOfflineInvestigation(model?: string) {
  return !!model && model.toLowerCase().startsWith('fallback')
}
