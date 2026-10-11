import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { fetchIncident, investigateIncident, fetchIncidentDefenses, remediateWithDefense, fetchFaults, buildLocalFallback, isOfflineInvestigation, type IncidentDetail, type InvestigationResult, type Hypothesis, type DefenseSuggestion, type FaultConfig } from '@/api/ops'
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
  return <span className={`px-3 py-1 rounded-full text-sm font-medium ${configs[status]}`}>{status === 'firing' ? 'Active' : 'Resolved'}</span>
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
            <span className={`font-medium ${likelihoodColor}`}>Probability: {(hypothesis.likelihood * 100).toFixed(0)}%</span>
          </div>
          {hypothesis.supportingEvidence.length > 0 && (
            <div className="space-y-1 text-sm">
              <p className="font-medium text-gray-700">Supporting evidence:</p>
              <ul className="list-disc list-inside text-gray-600 space-y-1">
                {hypothesis.supportingEvidence.map((e, i) => <li key={i}>{e}</li>)}
              </ul>
            </div>
          )}
          {hypothesis.contradictingEvidence.length > 0 && (
            <div className="mt-2 space-y-1 text-sm">
              <p className="font-medium text-gray-700">Contradicting evidence:</p>
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
  const [offlineNote, setOfflineNote] = useState<string | null>(null)
  const [suggestions, setSuggestions] = useState<DefenseSuggestion[]>([])
  const [faults, setFaults] = useState<FaultConfig[]>([])
  const [defenseBusy, setDefenseBusy] = useState<string | null>(null)
  const [defenseMsg, setDefenseMsg] = useState<string | null>(null)
  const [rps, setRps] = useState(10)
  const [burst, setBurst] = useState(20)

  useEffect(() => {
    if (!incidentId) return
    async function load() {
      try {
        setError(null)
        setLoading(true)
        const data = await fetchIncident(incidentId)
        setIncident(data)
        if (data.investigation) setInvestigation(data.investigation)
        try {
          const d = await fetchIncidentDefenses(incidentId)
          setSuggestions(d.suggestions)
          const rl = d.suggestions.find(s => s.action === 'enable_rate_limit')
          if (rl?.requestsPerSecond) setRps(rl.requestsPerSecond)
          if (rl?.burst) setBurst(rl.burst)
        } catch { /* defenses optional */ }
        try {
          setFaults(await fetchFaults())
        } catch { /* faults optional */ }
      } catch (err) {
        setError('Failed to load incident')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [incidentId])

  const reloadDefenses = async () => {
    try {
      const d = await fetchIncidentDefenses(incidentId)
      setSuggestions(d.suggestions)
    } catch { /* ignore */ }
    try {
      setFaults(await fetchFaults())
    } catch { /* ignore */ }
  }

  const handleInvestigate = async () => {
    if (!incidentId) return
    setInvestigating(true)
    setError(null)
    setOfflineNote(null)
    try {
      const result = await investigateIncident(incidentId)
      setInvestigation(result)
      if (isOfflineInvestigation(result.model)) {
        setOfflineNote('Ollama unreachable — showing offline rule-based analysis (what failed + when). Start Ollama for a full LLM review.')
      }
    } catch (err) {
      // Never leave the user with only "failed": render hardcoded
      // what/where/when analysis from the already-loaded incident.
      console.error(err)
      if (incident) {
        setInvestigation(buildLocalFallback(incident))
        setOfflineNote('Backend unreachable — showing local offline analysis (what failed + when).')
      } else {
        setError('Investigation failed with Gemma. Check that Ollama is running.')
      }
    } finally {
      setInvestigating(false)
    }
  }

  const handleDefense = async (s: DefenseSuggestion) => {
    if (!incidentId || defenseBusy) return
    setDefenseBusy(s.id)
    setDefenseMsg(null)
    try {
      if (s.action === 'enable_rate_limit') {
        await remediateWithDefense(incidentId, {
          action: 'enable_rate_limit',
          target: s.target,
          requestsPerSecond: rps,
          burst,
          reason: `Enable rate limiter on ${s.target} for ${incident?.type || 'incident'} defense`,
        })
        setDefenseMsg(`Rate limiter active on ${s.target}: ${rps} req/s, burst ${burst}. Flood traffic now gets HTTP 429.`)
      } else if (s.action === 'disable_defense') {
        await remediateWithDefense(incidentId, {
          action: 'disable_defense',
          target: s.target,
          reason: `Disable rate limiter on ${s.target} after recovery`,
        })
        setDefenseMsg(`Rate limiter disabled on ${s.target}.`)
      } else {
        setDefenseMsg('Use the Simulations panel or Remediate stop_fault action to remove the injected fault.')
      }
      await reloadDefenses()
    } catch (err) {
      setDefenseMsg(err instanceof Error ? err.message : 'Defense action failed')
    } finally {
      setDefenseBusy(null)
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
        <p className="text-red-600">{error || 'Incident not found'}</p>
        <Link to="/ops" className="mt-4 inline-block px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600">
          Back to dashboard
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <Link to="/ops" className="text-sm text-gray-500 hover:text-gray-700 mb-2 inline-block">← Back to dashboard</Link>
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
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Observed Metrics</h2>
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
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Baseline (Normal Behavior)</h2>
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
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Thresholds / Limits</h2>
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
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Collected Evidence</h2>
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
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Investigation with Gemma</h2>
            {!investigation ? (
              <button
                className="w-full px-4 py-3 bg-brand-500 text-white rounded-lg hover:bg-brand-600 disabled:opacity-50"
                onClick={handleInvestigate}
                disabled={investigating}
              >
                {investigating ? 'Investigating...' : 'Start Investigation'}
              </button>
            ) : (
              <div className="space-y-4">
                <div className={investigation.model && isOfflineInvestigation(investigation.model) ? 'bg-amber-50 border border-amber-200 rounded-lg p-3' : 'bg-green-50 border border-green-200 rounded-lg p-3'}>
                  <p className={investigation.model && isOfflineInvestigation(investigation.model) ? 'text-sm text-amber-800' : 'text-sm text-green-800'}>{investigation.model && isOfflineInvestigation(investigation.model) ? (<>Offline analysis with <strong>{investigation.model}</strong> — what failed + when, no Ollama needed</>) : (<>Investigation completed with <strong>{investigation.model}</strong></>)}</p>
                  <p className={investigation.model && isOfflineInvestigation(investigation.model) ? 'text-sm text-amber-700 mt-1' : 'text-sm text-green-700 mt-1'}>Confidence: {(investigation.confidence * 100).toFixed(0)}%</p>
                </div>
                {offlineNote && <p className="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded p-2">{offlineNote}</p>}
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Summary</p>
                  <p className="text-sm text-gray-600">{investigation.summary}</p>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Likely Root Cause</p>
                  <p className="text-sm text-gray-600 bg-gray-50 p-3 rounded">{investigation.rootCause}</p>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Hypotheses</p>
                  <div className="space-y-2">
                    {investigation.hypotheses.map((h) => (
                      <HypothesisCard key={h.rank} hypothesis={h} />
                    ))}
                  </div>
                </div>
                {investigation.missingEvidence.length > 0 && (
                  <div className="space-y-2">
                    <p className="font-medium text-gray-900">Missing Evidence</p>
                    <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                      {investigation.missingEvidence.map((e, i) => <li key={i}>{e}</li>)}
                    </ul>
                  </div>
                )}
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Recommended Steps</p>
                  <ol className="list-decimal list-inside text-sm text-gray-600 space-y-1">
                    {investigation.recommendedSteps.map((s, i) => <li key={i}>{s}</li>)}
                  </ol>
                </div>
                <div className="space-y-2">
                  <p className="font-medium text-gray-900">Recovery</p>
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
          <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
            <h2 className="text-lg font-semibold text-gray-900 mb-1">Defend this attack</h2>
            <p className="text-sm text-gray-500 mb-4">Per-attack-type controls. The rate limiter is enforced by the backend (excess requests get HTTP 429).</p>
            {suggestions.length === 0 && <p className="text-sm text-gray-500">Loading defense options…</p>}
            {suggestions.filter(s => s.action === 'enable_rate_limit').map(s => (
              <div key={s.id} className="border border-gray-200 rounded-lg p-4 mb-3">
                <div className="flex items-center justify-between mb-1">
                  <p className="font-medium text-gray-900">{s.primary ? 'Recommended: ' : ''}Rate limiter — {s.target}</p>
                  {s.active && <span className="px-2 py-0.5 bg-green-100 text-green-700 rounded-full text-xs font-medium">Active</span>}
                </div>
                <p className="text-sm text-gray-600 mb-3">{s.description}</p>
                <div className="grid grid-cols-2 gap-3 mb-3">
                  <label className="text-sm text-gray-600">Requests/sec
                    <input type="number" min={1} max={1000} value={rps} onChange={e => setRps(Math.max(1, +e.target.value || 1))} className="mt-1 w-full border border-gray-300 rounded px-2 py-1 text-gray-900" />
                  </label>
                  <label className="text-sm text-gray-600">Burst
                    <input type="number" min={1} max={1000} value={burst} onChange={e => setBurst(Math.max(1, +e.target.value || 1))} className="mt-1 w-full border border-gray-300 rounded px-2 py-1 text-gray-900" />
                  </label>
                </div>
                <button onClick={() => void handleDefense(s)} disabled={!!defenseBusy} className="w-full px-4 py-2 bg-brand-500 text-white rounded-lg hover:bg-brand-600 disabled:opacity-50">
                  {defenseBusy === s.id ? 'Applying…' : s.active ? 'Update rate limiter' : 'Enable rate limiter'}
                </button>
              </div>
            ))}
            {suggestions.filter(s => s.action !== 'enable_rate_limit').map(s => (
              <div key={s.id} className="border border-gray-200 rounded-lg p-4 mb-3">
                <p className="font-medium text-gray-900 mb-1">{s.label}</p>
                <p className="text-sm text-gray-600 mb-3">{s.description}</p>
                {s.action === 'stop_fault' && faults.length > 0 && (
                  <p className="text-xs text-gray-500 mb-2">Active faults: {faults.filter(f => f.enabled).map(f => `${f.type} (${f.id})`).join(', ') || 'none'}</p>
                )}
                <button onClick={() => void handleDefense(s)} disabled={!!defenseBusy} className="w-full px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 disabled:opacity-50">
                  {defenseBusy === s.id ? 'Applying…' : s.action === 'disable_defense' ? 'Disable rate limiter' : 'Stop injected fault (via Simulations)'}
                </button>
              </div>
            ))}
            {defenseMsg && <p className="text-sm text-gray-700 bg-gray-50 border border-gray-200 rounded p-2">{defenseMsg}</p>}
          </div>
        </div>
      </div>

      {/* Timeline */}
      <div className="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Timeline</h2>
        <div className="space-y-4">
          <div className="flex items-start gap-3">
            <div className="w-2 h-2 bg-red-500 rounded-full mt-2 flex-shrink-0" />
            <div>
              <p className="font-medium text-gray-900">Incident detected</p>
              <p className="text-sm text-gray-500">{new Date(incident.startedAt).toLocaleString('en-US')}</p>
            </div>
          </div>
          {incident.investigation && incident.investigation.createdAt && (() => {
            const createdAt = incident.investigation.createdAt!
            return (
              <div className="flex items-start gap-3">
                <div className="w-2 h-2 bg-brand-500 rounded-full mt-2 flex-shrink-0" />
                <div>
                  <p className="font-medium text-gray-900">Gemma investigation completed</p>
                  <p className="text-sm text-gray-500">{new Date(createdAt).toLocaleString('en-US')}</p>
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
                    <p className="font-medium text-gray-900">Incident resolved</p>
                    <p className="text-sm text-gray-500">{new Date(resolvedAt).toLocaleString('en-US')}</p>
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