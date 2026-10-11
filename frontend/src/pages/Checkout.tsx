import { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { checkoutApi } from '../lib/api';
import { Loading } from '../components/Loading';
import { useToast, ToastContainer } from '../components/Toast';
import type { CartValidateResponse, CartItem } from '../types';

interface StepIndicatorProps {
  step: 'payment' | 'processing' | 'complete' | 'error';
  targetStep: 'payment' | 'processing' | 'complete';
  label: string;
}

function StepIndicator({ step, targetStep, label }: StepIndicatorProps) {
  const isActive = step === targetStep;
  const isCompleted = (targetStep === 'payment' && ['processing', 'complete', 'error'].includes(step)) ||
    (targetStep === 'processing' && ['complete', 'error'].includes(step)) ||
    (targetStep === 'complete' && step === 'complete');

  const background = isActive ? '#dbeafe' : isCompleted ? '#dcfce7' : '#e5e7eb';
  const color = isActive ? '#1e40af' : isCompleted ? '#166534' : '#6b7280';

  return (
    <div className={`step ${isActive ? 'active' : isCompleted ? 'completed' : ''}`} style={{
      padding: '0.5rem 1rem',
      borderRadius: '9999px',
      background,
      color,
      fontWeight: 600
    }}>
      {label}
    </div>
  );
}

export function Checkout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { toasts, success, error: showError } = useToast();
  const [step, setStep] = useState<'payment' | 'processing' | 'complete' | 'error'>('payment');
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [orderId, setOrderId] = useState<string | null>(null);
  const [processingTime, setProcessingTime] = useState<number | null>(null);

  const cart = (location.state as { cart?: CartItem[] })?.cart || [];
  const validated = (location.state as { validated?: CartValidateResponse })?.validated;

  useEffect(() => {
    if (!validated?.valid || cart.length === 0) {
      navigate('/cart');
    }
  }, [cart, validated, navigate]);

  const subtotal = validated ? (validated.total_cents / 100).toFixed(2) : '0.00';

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!validated) return;

    setProcessing(true);
    setError(null);
    setStep('processing');

    try {
      const formData = new FormData(e.currentTarget);
      const paymentMethod = formData.get('payment_method') as string;

      const response = await checkoutApi.checkout(cart, paymentMethod);
      
      if (response.success) {
        setOrderId(response.order_id || 'ORD-' + Date.now());
        setProcessingTime(response.processing_time_ms || 0);
        setStep('complete');
        success('Order placed successfully!');
        localStorage.removeItem('faultline-cart');
      } else {
        setError(response.message);
        setStep('error');
        showError(response.message);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Checkout failed';
      setError(message);
      setStep('error');
      showError(message);
    } finally {
      setProcessing(false);
    }
  };

  const handleRetry = () => {
    setStep('payment');
    setError(null);
  };

  if (step === 'complete') {
    return (
      <div className="container">
        <ToastContainer toasts={toasts} />
        <div className="card" style={{ maxWidth: '600px', margin: '2rem auto' }}>
          <div className="card-body" style={{ textAlign: 'center', padding: '3rem' }}>
            <div style={{ fontSize: '4rem', marginBottom: '1rem' }}>✅</div>
            <h1 style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>Order Confirmed!</h1>
            <p style={{ color: '#6b7280', marginBottom: '1.5rem' }}>
              Thank you for your purchase. Your order has been processed.
            </p>
            <div style={{ background: '#f3f4f6', borderRadius: '0.5rem', padding: '1.5rem', marginBottom: '1.5rem', textAlign: 'left' }}>
              <div style={{ fontWeight: '600', marginBottom: '0.5rem' }}>Order ID: {orderId}</div>
              <div style={{ color: '#6b7280' }}>Processing time: {processingTime}ms</div>
            </div>
            <Link to="/store" className="btn btn-primary" style={{ padding: '1rem 2rem', fontSize: '1.125rem' }}>
              Continue Shopping
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <ToastContainer toasts={toasts} />
      <div style={{ maxWidth: '600px', margin: '0 auto' }}>
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem', justifyContent: 'center' }}>
          <StepIndicator step={step} targetStep="payment" label="1. Payment" />
          <StepIndicator step={step} targetStep="processing" label="2. Processing" />
          <StepIndicator step={step} targetStep="complete" label="3. Complete" />
        </div>

        {step === 'payment' && (
          <div className="card">
            <div className="card-header">
              <h2>Payment Details</h2>
            </div>
            <div className="card-body">
              <div style={{ marginBottom: '1.5rem', padding: '1rem', background: '#f9fafb', borderRadius: '0.5rem' }}>
                <div className="summary-row" style={{ fontWeight: 600, fontSize: '1.125rem' }}>
                  <span>Order Total</span>
                  <span>${subtotal}</span>
                </div>
              </div>

              <form onSubmit={handleSubmit} className="checkout-form">
                <div className="form-group">
                  <label className="form-label">Payment Method (Simulated)</label>
                  <select
                    name="payment_method"
                    className="form-input"
                    required
                    defaultValue="simulated_card"
                  >
                    <option value="simulated_card">Simulated Credit Card (Test)</option>
                    <option value="simulated_bank">Simulated Bank Transfer (Test)</option>
                    <option value="simulated_wallet">Simulated Digital Wallet (Test)</option>
                  </select>
                  <p className="error-text" style={{ marginTop: '0.25rem', color: '#6b7280', fontSize: '0.75rem' }}>
                    This is a simulation. No real payment is processed.
                  </p>
                </div>

                <button
                  type="submit"
                  className="btn btn-primary"
                  style={{ width: '100%', padding: '1rem', fontSize: '1.125rem' }}
                  disabled={processing}
                >
                  {processing ? <Loading size="sm" /> : `Pay $${subtotal}`}
                </button>
              </form>
            </div>
          </div>
        )}

        {step === 'processing' && (
          <div className="card" style={{ textAlign: 'center', padding: '3rem' }}>
            <Loading size="lg" text="Processing payment..." />
            <p style={{ marginTop: '1rem', color: '#6b7280' }}>
              This may take a moment. Please do not refresh the page.
            </p>
          </div>
        )}

        {step === 'error' && (
          <div className="card" style={{ textAlign: 'center', padding: '3rem' }}>
            <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>❌</div>
            <h2>Payment Failed</h2>
            <p style={{ color: '#6b7280', marginBottom: '1.5rem' }}>{error || 'An unknown error occurred'}</p>
            <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center' }}>
              <button className="btn btn-primary" onClick={handleRetry}>Try Again</button>
              <Link to="/cart" className="btn btn-secondary">Back to Cart</Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}