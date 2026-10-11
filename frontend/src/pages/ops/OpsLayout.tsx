import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { useUiStore } from '@/stores/ui'
import { useOperations } from '@/hooks/useOperations'
import type { HealthResponse, OpsOverviewResponse } from '@/api/ops'
import './ops.css'

export function OpsIcon({ name }: { name: string }) {
 const paths: Record<string, string> = {
  overview: 'M3 3h7v7H3z M14 3h7v7h-7z M3 14h7v7H3z M14 14h7v7h-7z',
  simulations: 'M9 3h6 M10 3v6L4 19q-1 2 2 2h12q3 0 2-2L14 9V3 M8 15h8',
  incidents: 'M12 3 2 21h20L12 3z M12 9v5 M12 17v1',
  services: 'M9 2h6v6H9z M2 16h6v6H2z M16 16h6v6h-6z M12 8v5 M5 16v-3h14v3',
  settings: 'M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8 M12 2v3 M12 19v3 M2 12h3 M19 12h3 M5 5l2 2 M17 17l2 2 M5 19l2-2 M17 7l2-2',
  arrow: 'M5 12h14 M14 7l5 5-5 5',
  pulse: 'M2 12h5l3-7 4 14 3-7h5',
 }
 return <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d={paths[name] || paths.pulse} /></svg>
}
export default function OpsLayout() {
 const [collapsed, setCollapsed] = useState(false)
 const dark = useUiStore(s => s.darkMode)
 const toggleTheme = useUiStore(s => s.toggleDarkMode)
 const health = useOperations<HealthResponse>('/api/health', 15000)
 const access = useOperations<OpsOverviewResponse>('/api/ops/overview', 30000)
 const links = [['/', 'overview', 'Overview'], ['/ops/simulations', 'simulations', 'Simulations'], ['/ops/incidents', 'incidents', 'Incidents'], ['/ops/services', 'services', 'Services'], ['/ops/settings', 'settings', 'Settings']]
 return <div className={`ops-shell ${collapsed ? 'rail-collapsed' : ''}`}>
  <aside className="ops-rail">
   <NavLink to="/" className="ops-brand"><span className="brand-symbol"><OpsIcon name="pulse" /></span><span>faultline<span className="brand-live">LIVE</span></span></NavLink>
   <div className="rail-caption">OPERATIONS INTELLIGENCE</div>
   <nav aria-label="Operations">{links.map(([to, icon, label]) => <NavLink end={to === '/'} to={to} title={label} key={to}><OpsIcon name={icon} /><span>{label}</span></NavLink>)}</nav>
   <div className="rail-bottom"><div className="rail-note"><span className="status-dot" />One system. Every signal.<small>Go · PostgreSQL · Ollama</small></div><button className="rail-toggle" onClick={() => setCollapsed(c => !c)} aria-label={collapsed ? 'Expand navigation' : 'Collapse navigation'}>{collapsed ? '→' : '←'}<span>Collapse navigation</span></button></div>
  </aside>
  <div className="ops-workspace">
   <header className="ops-command"><div className="command-context"><span className="environment">{access.data ? access.data.demoMode ? 'Local Demo' : 'Authenticated operations' : 'Access unverified'}</span><span className="command-divider" /><span className={`connection ${health.error ? 'critical' : ''}`}><span className="status-dot" />{health.loading ? 'Connecting' : health.error ? 'API unavailable' : 'API connected'}</span></div><div className="command-actions"><span className="last-refresh">Last connection {health.lastSuccess ? new Date(health.lastSuccess).toLocaleTimeString() : '—'}</span><button title="Switch appearance" aria-label="Switch light or dark appearance" onClick={toggleTheme}>{dark ? '☀' : '☾'}</button></div></header>
   <main className="ops-main"><Outlet /></main>
   <footer className="ops-footer"><span>FAULTLINE LIVE / LOCAL OPERATIONS</span><span>Real measurements. Evidence before action.</span></footer>
  </div>
 </div>
}
