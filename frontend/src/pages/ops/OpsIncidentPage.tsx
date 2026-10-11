import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { fetchIncident, investigateIncident, type IncidentDetail, type InvestigationResult, type Hypothesis } from '@/api/ops'
import { formatMs } from '@/lib/format'

function SeverityBadge({ severity }: { severity: IncidentDetail['severity'] }) {
  const configs = {
    low: 'bg-blue-100 text-blue-700',
    medium: 'bg-yellow-100 text-yellow-700',
    high: 'bg-orange-100 text-orange-700',
    critical: 'bg-red-100 text-red-700',
  }
  return <span className={`px-3 py-1 rounded-full text-sm font-medium ${configs[severity]}`}>{severity.toUpperCase()}</span>
}

function StatusBadge({ status }: { status: IncidentDetail['status'] }) {
  const configs = {
    firing: 'bg-red-100 text-red-700 animate-pulse',
    resolved: 'bg-green-100 text-green-700',
  }
  return <span className={`px-3 py-1 rounded-full text-sm font-medium ${configs[status]}`}>{status === 'firing' ? 'Ativo' : 'Resolvido'}</span>
}

function EvidenceSection({ title, items }: { title: string; items: Record<string, unknown>[] | string[] }) {
  return (
    <div className="bg-gray-50 rounded-lg p-4">
      <h4 className="font-medium text-gray-900 mb-3">{title}</h4>
      <ul className="space-y-2 text-sm text-gray-600">
        {items.map((item, i) => (
          <li key={i} className="flex items-start gap-2">
            <span className="text-brand-500">•</span>
            <span>{typeof item === 'string' ? item : JSON.stringify(item)}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function HypothesisCard({ hypothesis }: { hypothesis: Hypothesis }) {
  const likelihoodColor = hypothesis.likelihood > 0.7 ? 'text-green-600' : hypothesis.likelihood > 0.4 ? 'text-yellow-600' : 'text-red-600'
  return (
    <div className="border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <span className="px-2 py-0.5 bg-brand-100 text-brand-700 rounded-full text-xs font-medium">#{hypothesis.rank}</span>
            <span className="font-medium text-gray-900">{hypothesis.hypothesis}</span>
          </div>
          <div className="flex items-center gap-3 text-sm text-gray-600 mb-3">
            <span className={`font-medium ${likelihoodColor}`}>Probabilidade: {(hypothesis.likelihood * 100).toFixed(0)}%</span>
          </div>
          {hypothesis.supportingEvidence.length > 0 && (
            <div className="space-y-1 text-sm">
              <p className="font-medium text-gray-700">Evidências de apoio:</p>
              <ul className="list-disc list-inside text-gray-600 space-y-1">
                {hypothesis.supportingEvidence.map((e, i) => <li key={i}>{e}</li>)}
              </ul>
            </div>
          )}
          {hypothesis.contradictingEvidence.length > 0 && (
            <div className="mt-2 space-y-1 text-sm">
              <p className="font-medium text-gray-700">Evidências contraditórias:</p>
              <ul className="list-disc list-inside text-gray-600 space-y-1">
                {hypothesis.contradictingEvidence.map((e, i) => <li key={i}>{e}</li>)}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default function OpsIncidentPage() {
  const { id } = useParams<{ id: string }>()
  const incidentId = id ?? ''
  const [incident, setIncident] = useState<IncidentDetail | null>(null)
  const [investigation, setInvestigation] = useState<InvestigationResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [investigating, setInvestigating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!incidentId) return
    async function load() {
      try {
        setError(null)
        setLoading(true)
        const data = await fetchIncident(incidentId)
        setIncident(data)
        if (data.investigation) setInvestigation(data.investigation)
      } catch (err) {
        setError('Falha ao carregar incidente')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [incidentId])

  const handleInvestigate = async () => {
    if (!incidentId) return
    setInvestigating(true)
    setError(null)
    try {
      const result = await investigateIncident(incidentId)
      setInvestigation(result)
    } catch (err) {
      setError('Falha na investigação com Gemma. Verifique se o Ollama está rodando.')
      console.error(err)
    } finally {
      setInvestigating(false)
    }
  }

  if (loading) {
    return (
      <div className="flex h-96 w-full items-center justify-center">
        <span className="h-8 w-8 animate-spin rounded-full border-2 border-brand-500 border-t-transparent" />
      </div>
    )
  }

  if (!incident) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600">{error || 'Incidente não encontrado'}</p>
        <Link to="/ops" className="mt-4 inline-block px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600">
          Voltar ao painel
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link to="/ops" className="text-sm text-gray-500 hover:text-gray-700 mb-2 inline-block">← Voltar ao painel</Link>
          <h1 className="text-2xl font-bold text-gray-900">{incident.title}</h1>
          <p className="text-gray-500 mt-1">{incident.description}</p>
        </div>
        <div className="flex items-center gap-3">
          <SeverityBadge severity={incident.severity} />
          <StatusBadge status={incident.status} />
        </div>
      </div>

      {/* Incident Details */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Métricas Observadas</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              {Object.entries(incident.observed).map(([key, value]) => (
                <div key={key} className="bg-gray-50 rounded-lg p-3">
                  <p className="text-gray-500">{key}</p>
                  <p className="font-mono font-medium text-gray-900">{typeof value === 'number' ? (key.includes('latency') ? formatMs(value) : value.toLocaleString()) : String(value)}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Baseline (Comportamento Normal)</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              {Object.entries(incident.baseline).map(([key, value]) => (
                <div key={key} className="bg-gray-50 rounded-lg p-3">
                  <p className="text-gray-500">{key}</p>
                  <p className="font-mono font-medium text-gray-900">{typeof value === 'number' ? (key.includes('latency') ? formatMs(value) : value.toLocaleString()) : String(value)}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Thresholds / Limites</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              {Object.entries(incident.threshold).map(([key, value]) => (
                <div key={key} className="bg-gray-50 rounded-lg p-3">
                  <p className="text-gray-500">{key}</p>
                  <p className="font-mono font-medium text-gray-900">{typeof value === 'number' ? (key.includes('latency') ? formatMs(value) : value.toLocaleString()) : String(value)}</p>
                </div>
              ))}
            </div>
          </div>

          {incident.evidence.length > 0 && (
            <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Evidências Coletadas</h2>
              <ul className="space-y-2 text-sm text-gray-600">
                {incident.evidence.map((e, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <span className="text-brand-500">•</span>
                    <span>{e}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Investigation Panel */}
        <div className="space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm sticky top-24">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Investigação com Gemma</h2>
            {!investigation ? (
              <button
                className="w-full px-4 py-3 bg-brand-500 text-white rounded-lg hover:bg-brand-600 disabled:opacity-50"
                onClick={handleInvestigate}
                disabled={investigating}
              >
                {investigating ? 'Investigando...' : 'Iniciar Investigação'}
              </button>
            ) : (
              <div className="space-y-4">
                <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                  <p className="text-sm text-green-800">Investigação concluída com <strong>{investigation.model}</strong></p>
                  <p className="text-sm text-green-700 mt-1">Confiança: {(investigation.confidence * 100).toFixed(0)}%</p>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Resumo</p>
                  <p className="text-sm text-gray-600">{investigation.summary}</p>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Causa Raiz Provável</p>
                  <p className="text-sm text-gray-600 bg-gray-50 p-3 rounded">{investigation.rootCause}</p>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Hipóteses</p>
                  <div className="space-y-2">
                    {investigation.hypotheses.map((h) => (
                      <HypothesisCard key={h.rank} hypothesis={h} />
                    ))}
                  </div>
                </div>
                {investigation.missingEvidence.length > 0 && (
                  <div className="space-y-2">
                    <p className="font-medium text-gray-900">Evidências Faltantes</p>
                    <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                      {investigation.missingEvidence.map((e, i) => <li key={i}>{e}</li>)}
                    </ul>
                  </div>
                )}
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Passos Recomendados</p>
                  <ol className="list-decimal list-inside text-sm text-gray-600 space-y-1">
                    {investigation.recommendedSteps.map((s, i) => <li key={i}>{s}</li>)}
                  </ol>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Recuperação</p>
                  <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                    {investigation.recoveryRecommendations.map((r, i) => <li key={i}>{r}</li>)}
                  </ul>
                </div>
                <div className="pt-4 border-t border-gray-200 text-sm text-gray-500">
                  <p>{investigation.confidenceExplanation}</p>
                </div>
              </div>
            )}
            {error && <p className="mt-4 text-sm text-red-600">{error}</p>}
          </div>
        </div>
      </div>

      {/* Timeline */}
      <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Linha do Tempo</h2>
        <div className="space-y-4">
          <div className="flex items-start gap-3">
            <div className="w-2 h-2 bg-red-500 rounded-full mt-2 flex-shrink-0" />
            <div>
              <p className="font-medium text-gray-900">Incidente detectado</p>
              <p className="text-sm text-gray-500">{new Date(incident.startedAt).toLocaleString('pt-BR')}</p>
            </div>
          </div>
          {incident.investigation && incident.investigation.createdAt && (() => {
            const createdAt = incident.investigation.createdAt!
            return (
              <div className="flex items-start gap-3">
                <div className="w-2 h-2 bg-brand-500 rounded-full mt-2 flex-shrink-0" />
                <div>
                  <p className="font-medium text-gray-900">Investigação Gemma concluída</p>
                  <p className="text-sm text-gray-500">{new Date(createdAt).toLocaleString('pt-BR')}</p>
                </div>
              </div>
            )
          })()}
          {incident.resolvedAt && (
            (() => {
              const resolvedAt = incident.resolvedAt!
              return (
                <div className="flex items-start gap-3">
                  <div className="w-2 h-2 bg-green-500 rounded-full mt-2 flex-shrink-0" />
                  <div>
                    <p className="font-medium text-gray-900">Incidente resolvido</p>
                    <p className="text-sm text-gray-500">{new Date(resolvedAt).toLocaleString('pt-BR')}</p>
                  </div>
                </div>
              )
            })()
          )}
        </div>
      </div>
    </div>
  )
}