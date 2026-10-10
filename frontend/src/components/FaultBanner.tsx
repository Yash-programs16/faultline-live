import type { FaultState } from '../types';

interface FaultBannerProps {
  fault: FaultState | undefined;
  onDismiss: () => void;
}

export function FaultBanner({ fault, onDismiss }: FaultBannerProps) {
  if (!fault || !fault.active) return null;

  const isLatency = fault.scenario === 'latency';
  const delay = isLatency ? (fault.params.delay_ms as number) : 0;
  const rate = !isLatency ? (fault.params.failure_rate as number) : 0;

  return (
    <div className={`fault-banner ${fault.active ? 'active' : ''}`} role="alert">
      <div>
        <strong>⚠️ Active Fault Injection: </strong>
        {isLatency
          ? `Payment latency injection active (+${delay}ms delay)`
          : `Payment failure injection active (${Math.round(rate * 100)}% failure rate)`}
      </div>
      <button className="btn btn-secondary" onClick={onDismiss} aria-label="Dismiss fault notification">
        Dismiss
      </button>
    </div>
  );
}