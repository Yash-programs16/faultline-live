import { useState, useEffect, useCallback } from 'react';
import { metricsApi, faultsApi } from '../lib/api';
import { FaultBanner } from '../components/FaultBanner';
import { MetricCard } from '../components/MetricCard';
import { LineChart } from '../components/LineChart';
import { Loading } from '../components/Loading';
import { useToast, ToastContainer } from '../components/Toast';
import type { MetricWindow, FaultState, FaultRequest } from '../types';

const POLL_INTERVAL = 2000;
const MAX_DATA_POINTS = 60;

export function Dashboard() {
  const { toasts, success, error: showError } = useToast();
  const [windows, setWindows] = useState<MetricWindow[]>([]);
  const [fault, setFault] = useState<FaultState | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [polling, setPolling] = useState(true);
  const [faultForm, setFaultForm] = useState<FaultRequest>({
    scenario: 'latency',
    params: { delay_ms: 2000, duration_seconds: 60 },
  });
  const [activating, setActivating] = useState(false);
  const [deactivating, setDeactivating] = useState(false);

  const fetchMetrics = useCallback(async () => {
    try {
      const response = await metricsApi.live();
      setWindows((prev: MetricWindow[]) => {
        const newWindows = response.windows;
        const combined = [...prev, ...newWindows];
        const unique = combined.reduce((acc: Map<string, MetricWindow>, w: MetricWindow) => {
          const key = `${w.service}-${w.window_start}`;
          if (!acc.has(key)) acc.set(key, w);
          return acc;
        }, new Map<string, MetricWindow>());
        return Array.from(unique.values())
          .sort((a: MetricWindow, b: MetricWindow) => new Date(a.window_start).getTime() - new Date(b.window_start).getTime())
          .slice(-MAX_DATA_POINTS);
      });
      if (response.current_fault) {
        setFault(response.current_fault);
      }
    } catch (err) {
      console.error('Failed to fetch metrics:', err);
    }
  }, []);

  useEffect(() => {
    fetchMetrics();
    setLoading(false);

    if (!polling) return;

    const interval = setInterval(fetchMetrics, POLL_INTERVAL);
    return () => clearInterval(interval);
  }, [fetchMetrics, polling]);

  const activateFault = async () => {
    setActivating(true);
    try {
      const response = await faultsApi.create(faultForm.scenario, faultForm.params);
      if (response.success && response.fault) {
        setFault(response.fault);
        success(`Fault activated: ${faultForm.scenario}`);
      } else {
        showError(response.message || 'Failed to activate fault');
      }
    } catch (err) {
      showError('Failed to activate fault');
      console.error(err);
    } finally {
      setActivating(false);
    }
  };

  const deactivateFault = async () => {
    setDeactivating(true);
    try {
      const response = await faultsApi.remove();
      if (response.success) {
        setFault(undefined);
        success('Fault deactivated');
      } else {
        showError(response.message || 'Failed to deactivate fault');
      }
    } catch (err) {
      showError('Failed to deactivate fault');
      console.error(err);
    } finally {
      setDeactivating(false);
    }
  };

  const checkoutWindows = windows.filter((w) => w.service === 'checkout');
  const paymentWindows = windows.filter((w) => w.service === 'payment');
  const apiWindows = windows.filter((w) => w.service === 'api');

  const latestCheckout = checkoutWindows[checkoutWindows.length - 1];

  const checkoutLatencyData = checkoutWindows.map((w) => ({
    timestamp: w.window_start,
    value: w.mean_latency_ms,
  }));
  const checkoutP95Data = checkoutWindows.map((w) => ({
    timestamp: w.window_start,
    value: w.p95_latency_ms,
  }));
  const errorRateData = checkoutWindows.map((w) => ({
    timestamp: w.window_start,
    value: w.error_rate * 100,
  }));
  const requestRateData = checkoutWindows.map((w) => ({
    timestamp: w.window_start,
    value: w.requests_per_second,
  }));

  return (
    <div className="container">
      <ToastContainer toasts={toasts} />

      <header style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: '700', marginBottom: '0.5rem' }}>Operational Dashboard</h1>
        <p style={{ color: '#6b7280' }}>Real-time telemetry from Faultline Live backend</p>
      </header>

      <FaultBanner fault={fault} onDismiss={() => setFault(undefined)} />

      <div className="grid grid-4" style={{ marginBottom: '2rem' }}>
        <MetricCard
          label="Checkout Latency (mean)"
          value={latestCheckout ? `${latestCheckout.mean_latency_ms.toFixed(0)}ms` : 'No data'}
          status={latestCheckout && latestCheckout.mean_latency_ms > 1000 ? 'warning' : 'ok'}
        />
        <MetricCard
          label="Checkout Latency (p95)"
          value={latestCheckout ? `${latestCheckout.p95_latency_ms.toFixed(0)}ms` : 'No data'}
          status={latestCheckout && latestCheckout.p95_latency_ms > 2000 ? 'warning' : 'ok'}
        />
        <MetricCard
          label="Error Rate"
          value={latestCheckout ? `${(latestCheckout.error_rate * 100).toFixed(1)}%` : 'No data'}
          status={latestCheckout && latestCheckout.error_rate > 0.05 ? 'danger' : 'ok'}
        />
        <MetricCard
          label="Request Rate"
          value={latestCheckout ? `${latestCheckout.requests_per_second.toFixed(1)}/s` : 'No data'}
        />
      </div>

      <div className="grid grid-2" style={{ marginBottom: '2rem', gap: '1.5rem' }}>
        <div className="card">
          <div className="card-header">
            <h3>Checkout Latency</h3>
          </div>
          <div className="card-body">
            {checkoutLatencyData.length === 0 ? (
              <div className="empty-state" style={{ padding: '2rem' }}>No checkout telemetry yet</div>
            ) : (
              <>
                <LineChart
                  data={checkoutLatencyData}
                  label="Mean Latency (ms)"
                  color="#2563eb"
                  yAxisLabel="ms"
                />
                <LineChart
                  data={checkoutP95Data}
                  label="P95 Latency (ms)"
                  color="#ef4444"
                  height={180}
                  yAxisLabel="ms"
                />
              </>
            )}
          </div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3>Error Rate & Request Rate</h3>
          </div>
          <div className="card-body">
            {errorRateData.length === 0 ? (
              <div className="empty-state" style={{ padding: '2rem' }}>No checkout telemetry yet</div>
            ) : (
              <>
                <LineChart
                  data={errorRateData}
                  label="Error Rate (%)"
                  color="#ef4444"
                  yAxisLabel="%"
                />
                <LineChart
                  data={requestRateData}
                  label="Requests/sec"
                  color="#16a34a"
                  height={180}
                  yAxisLabel="req/s"
                />
              </>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-2" style={{ marginBottom: '2rem', gap: '1.5rem' }}>
        <div className="card">
          <div className="card-header">
            <h3>Payment Service Latency</h3>
          </div>
          <div className="card-body">
            {paymentWindows.length === 0 ? (
              <div className="empty-state" style={{ padding: '2rem' }}>No payment telemetry yet</div>
            ) : (
              <LineChart
                data={paymentWindows.map((w) => ({ timestamp: w.window_start, value: w.mean_latency_ms }))}
                label="Mean Latency (ms)"
                color="#f59e0b"
                yAxisLabel="ms"
              />
            )}
          </div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3>API Request Rate</h3>
          </div>
          <div className="card-body">
            {apiWindows.length === 0 ? (
              <div className="empty-state" style={{ padding: '2rem' }}>No API telemetry yet</div>
            ) : (
              <LineChart
                data={apiWindows.map((w) => ({ timestamp: w.window_start, value: w.requests_per_second }))}
                label="Requests/sec"
                color="#8b5cf6"
                yAxisLabel="req/s"
              />
            )}
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-header">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h3>Fault Injection Controls</h3>
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 'normal' }}>
              <input
                type="checkbox"
                checked={polling}
                onChange={(e) => setPolling(e.target.checked)}
              />
              Auto-refresh ({POLL_INTERVAL / 1000}s)
            </label>
          </div>
        </div>
        <div className="card-body">
          <div style={{ display: 'grid', gap: '1.5rem', gridTemplateColumns: '1fr 1fr', maxWidth: '900px' }}>
            <div>
              <h4 style={{ marginBottom: '1rem' }}>Activate Fault</h4>
              <div className="form-group">
                <label className="form-label">Scenario</label>
                <select
                  className="form-input"
                  value={faultForm.scenario}
                  onChange={(e) => setFaultForm({ ...faultForm, scenario: e.target.value as 'latency' | 'failure' })}
                  disabled={fault?.active || activating}
                >
                  <option value="latency">Payment Latency Injection</option>
                  <option value="failure">Payment Failure Injection</option>
                </select>
              </div>

              {faultForm.scenario === 'latency' && (
                <div className="form-group">
                  <label className="form-label">Delay (ms)</label>
                  <input
                    type="number"
                    className="form-input"
                    value={faultForm.params.delay_ms || 2000}
                    onChange={(e) => setFaultForm({ ...faultForm, params: { ...faultForm.params, delay_ms: parseInt(e.target.value, 10) } })}
                    min={100}
                    max={10000}
                    disabled={fault?.active || activating}
                  />
                  <p style={{ fontSize: '0.75rem', color: '#6b7280' }}>Max 10000ms</p>
                </div>
              )}

              {faultForm.scenario === 'failure' && (
                <div className="form-group">
                  <label className="form-label">Failure Rate (0-1)</label>
                  <input
                    type="number"
                    className="form-input"
                    value={faultForm.params.failure_rate || 0.5}
                    onChange={(e) => setFaultForm({ ...faultForm, params: { ...faultForm.params, failure_rate: parseFloat(e.target.value) } })}
                    min={0.01}
                    max={1}
                    step={0.01}
                    disabled={fault?.active || activating}
                  />
                  <p style={{ fontSize: '0.75rem', color: '#6b7280' }}>Max 1.0 (100%)</p>
                </div>
              )}

              <div className="form-group">
                <label className="form-label">Duration (seconds)</label>
                <input
                  type="number"
                  className="form-input"
                  value={faultForm.params.duration_seconds || 60}
                  onChange={(e) => setFaultForm({ ...faultForm, params: { ...faultForm.params, duration_seconds: parseInt(e.target.value, 10) } })}
                  min={10}
                  max={300}
                  disabled={fault?.active || activating}
                />
                <p style={{ fontSize: '0.75rem', color: '#6b7280' }}>Max 300s (5 minutes)</p>
              </div>

              <button
                className="btn btn-danger"
                onClick={activateFault}
                disabled={fault?.active || activating}
                style={{ width: '100%' }}
              >
                {activating ? 'Activating...' : fault?.active ? 'Fault Active' : 'Activate Fault'}
              </button>
            </div>

            <div>
              <h4 style={{ marginBottom: '1rem' }}>Current Fault Status</h4>
              {fault?.active ? (
                <div style={{ padding: '1rem', background: '#fef3c7', borderRadius: '0.5rem', border: '1px solid #fcd34d' }}>
                  <div style={{ fontWeight: 600, marginBottom: '0.5rem' }}>
                    {fault.scenario === 'latency' ? '⚡ Latency Injection' : '💥 Failure Injection'}
                  </div>
                  <div style={{ fontSize: '0.875rem', color: '#92400e' }}>
                    {fault.scenario === 'latency'
                      ? `Delay: ${fault.params.delay_ms}ms`
                      : `Failure rate: ${Math.round((fault.params.failure_rate as number) * 100)}%`}
                  </div>
                  <div style={{ fontSize: '0.875rem', color: '#92400e', marginTop: '0.25rem' }}>
                    Activated: {new Date(fault.activated_at).toLocaleTimeString()}
                  </div>
                  {fault.expires_at && (
                    <div style={{ fontSize: '0.875rem', color: '#92400e', marginTop: '0.25rem' }}>
                      Expires: {new Date(fault.expires_at).toLocaleTimeString()}
                    </div>
                  )}
                  <button
                    className="btn btn-primary"
                    onClick={deactivateFault}
                    disabled={deactivating}
                    style={{ marginTop: '1rem', width: '100%' }}
                  >
                    {deactivating ? 'Deactivating...' : 'Deactivate Fault'}
                  </button>
                </div>
              ) : (
                <div style={{ padding: '1rem', background: '#f3f4f6', borderRadius: '0.5rem', textAlign: 'center', color: '#6b7280' }}>
                  No active fault injection
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {loading && <Loading text="Loading dashboard..." />}
    </div>
  );
}