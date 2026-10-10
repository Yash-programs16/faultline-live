interface MetricCardProps {
  label: string;
  value: string | number;
  status?: 'ok' | 'warning' | 'danger';
  subtext?: string;
}

export function MetricCard({ label, value, status = 'ok', subtext }: MetricCardProps) {
  return (
    <div className="metric-card">
      <div className="metric-label">{label}</div>
      <div className={`metric-value ${status === 'warning' || status === 'danger' ? 'warning' : 'ok'}`}>
        {value}
      </div>
      {subtext && <div style={{ fontSize: '0.75rem', color: '#9ca3af', marginTop: '0.25rem' }}>{subtext}</div>}
    </div>
  );
}