import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'
import { fetchOpsOverview, fetchMetricsHistory, type OpsOverviewResponse, type MetricSample } from '@/api/ops'
import { formatNumber, formatMs } from '@/lib/format'

function StatCard({ label, value, delta, trend }: { label: string; value: string; delta?: string; trend?: 'up' | 'down' | 'neutral' }) {
  const deltaColor = trend === 'up' ? 'text-red-600' : trend === 'down' ? 'text-green-600' : 'text-muted-foreground'
  const deltaIcon = trend === 'up' ? '▲' : trend === 'down' ? '▼' : '●'
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
      <p className="text-sm font-medium text-gray-500">{label}</p>
      <p className="mt-1 text-2xl font-bold text-gray-900">{value}</p>
      {delta && (
        <p className={`mt-1 text-sm font-medium ${deltaColor} flex items-center gap-1`}>
          <span>{deltaIcon}</span> {delta}
        </p>
      )}
    </div>
  )
}

function ServiceStatusBadge({ status }: { status: 'healthy' | 'degraded' | 'down' }) {
  const configs = {
    healthy: 'bg-green-100 text-green-700',
    degraded: 'bg-yellow-100 text-yellow-700',
    down: 'bg-red-100 text-red-700',
  }
  return <span className={`px-2 py-1 rounded-full text-xs font-medium ${configs[status]}`}>{status}</span>
}

function IncidentRow({ incident, onClick }: { incident: OpsOverviewResponse['incidents'][0]; onClick: () => void }) {
  const severityColors = {
    low: 'bg-blue-100 text-blue-700',
    medium: 'bg-yellow-100 text-yellow-700',
    high: 'bg-orange-100 text-orange-700',
    critical: 'bg-red-100 text-red-700',
  }
  const statusColors = {
    firing: 'bg-red-100 text-red-700 animate-pulse',
    resolved: 'bg-green-100 text-green-700',
  }
  return (
    <tr className="hover:bg-gray-50 cursor-pointer" onClick={onClick}>
      <td className="px-4 py-3 text-sm font-medium text-gray-900">{incident.id}</td>
      <td className="px-4 py-3 text-sm text-gray-600">{incident.title}</td>
      <td className="px-4 py-3 text-sm">
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${severityColors[incident.severity]}`}>
          {incident.severity.toUpperCase()}
        </span>
      </td>
      <td className="px-4 py-3 text-sm text-gray-500">{incident.affectedEndpoint}</td>
      <td className="px-4 py-3 text-sm">
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${statusColors[incident.status]}`}>
          {incident.status === 'firing' ? 'Ativo' : 'Resolvido'}
        </span>
      </td>
      <td className="px-4 py-3 text-sm text-gray-500">{new Date(incident.startedAt).toLocaleString('pt-BR')}</td>
    </tr>
  )
}

function MetricsChart({ samples, metric, color }: { samples: MetricSample[]; metric: keyof MetricSample; color: string }) {
  if (samples.length < 2) return <div className="h-48 flex items-center justify-center text-gray-400">Sem dados</div>
  const values = samples.map((s) => s[metric] as number)
  const max = Math.max(...values)
  const min = Math.min(...values)
  const range = max - min || 1
  return (
    <div className="h-48 relative">
      <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
        <polyline
          fill="none"
          stroke={color}
          strokeWidth="1.5"
          points={samples
            .map((_, i) => `${(i / (samples.length - 1)) * 100},${100 - ((values[i] - min) / range) * 90}`)
            .join(' ')}
        />
      </svg>
    </div>
  )
}

export default function OpsDashboardPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [overview, setOverview] = useState<OpsOverviewResponse | null>(null)
  const [history, setHistory] = useState<MetricSample[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [selectedIncident, setSelectedIncident] = useState<OpsOverviewResponse['incidents'][0] | null>(null)

  async function loadData() {
    try {
      setError(null)
      const [overviewData, historyData] = await Promise.all([
        fetchOpsOverview(),
        fetchMetricsHistory('1h'),
      ])
      setOverview(overviewData)
      setHistory(historyData.samples)
    } catch (err) {
      setError('Falha ao carregar dados operacionais')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
    const interval = setInterval(() => {
      if (autoRefresh) loadData()
    }, 10000)
    return () => clearInterval(interval)
  }, [autoRefresh])

  if (loading && !overview) {
    return (
      <div className="flex h-96 w-full items-center justify-center">
        <span className="h-8 w-8 animate-spin rounded-full border-2 border-brand-500 border-t-transparent" />
      </div>
    )
  }

  if (error && !overview) {
    return (
      <div className="flex h-96 w-full items-center justify-center">
        <div className="text-center">
          <p className="text-red-600">{error}</p>
          <button className="mt-4 px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600" onClick={loadData}>
            Tentar novamente
          </button>
        </div>
      </div>
    )
  }

  const metrics = overview?.metrics
  const services = overview?.services || []
  const incidents = overview?.incidents || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Painel de Operações</h1>
          <p className="text-gray-500 mt-1">Monitoramento em tempo real do Faultline Live</p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-600">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded border-gray-300 text-brand-500 focus:ring-brand-500"
            />
            Auto-refresh (10s)
          </label>
          <button className="px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600" onClick={loadData}>
            Atualizar
          </button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Taxa de Requisições"
          value={`${formatNumber(metrics?.requestRate || 0)}/s`}
          delta={metrics?.errorRate ? `${(metrics.errorRate * 100).toFixed(2)}% erros` : undefined}
          trend={metrics && metrics.errorRate > 0.05 ? 'up' : 'down'}
        />
        <StatCard
          label="Latência P95"
          value={formatMs(metrics?.latencyP95 || 0)}
          trend={metrics && metrics.latencyP95 > 500 ? 'up' : 'down'}
        />
        <StatCard
          label="Latência DB P95"
          value={formatMs(metrics?.dbLatencyP95 || 0)}
          trend={metrics && metrics.dbLatencyP95 > 200 ? 'up' : 'down'}
        />
        <StatCard
          label="Falhas Ativas"
          value={String(metrics?.activeFaults || 0)}
          trend={metrics && metrics.activeFaults > 0 ? 'up' : 'down'}
        />
      </div>

      {/* Services & Latency Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Serviços</h2>
          <div className="space-y-3">
            {services.map((svc) => (
              <div key={svc.name} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <ServiceStatusBadge status={svc.status} />
                  <span className="font-medium text-gray-900">{svc.name}</span>
                </div>
                <div className="flex items-center gap-4 text-sm text-gray-500">
                  {svc.latencyMs != null && <span>Latência: {formatMs(svc.latencyMs)}</span>}
                  {svc.errorRate != null && <span>Erro: {(svc.errorRate * 100).toFixed(2)}%</span>}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Latência P95 (últimas 1h)</h2>
          <MetricsChart samples={history} metric="latencyP95" color="#465fff" />
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Taxa de Erro (últimas 1h)</h2>
          <MetricsChart
            samples={history.map((s) => ({ ...s, errorRate: s.requestCount > 0 ? s.errorCount / s.requestCount : 0 }))}
            metric="errorRate"
            color="#f04438"
          />
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Latência DB P95 (últimas 1h)</h2>
          <MetricsChart samples={history} metric="dbLatencyP95" color="#12b76a" />
        </div>
      </div>

      {/* Incidents */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="p-5 border-b border-gray-200 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Incidentes</h2>
          {incidents.length > 0 && (
            <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded-full text-sm font-medium">{incidents.length} total</span>
          )}
        </div>
        {incidents.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            Nenhum incidente detectado. O sistema está estável.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr className="text-left text-sm font-medium text-gray-500">
                  <th className="px-4 py-3">ID</th>
                  <th className="px-4 py-3">Título</th>
                  <th className="px-4 py-3">Severidade</th>
                  <th className="px-4 py-3">Endpoint</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Início</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {incidents.map((inc) => (
                  <IncidentRow key={inc.id} incident={inc} onClick={() => navigate(`/ops/incidents/${inc.id}`)} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}