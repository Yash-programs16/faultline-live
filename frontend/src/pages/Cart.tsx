import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { cartApi } from '../lib/api';
import { CartItemComponent } from '../components/CartItem';
import { useToast, ToastContainer } from '../components/Toast';
import type { CartItem, CartValidateResponse } from '../types';

const CART_STORAGE_KEY = 'faultline-cart';

export function Cart() {
  const navigate = useNavigate();
  const { toasts, success, error: showError } = useToast();
  const [cart, setCart] = useState<CartItem[]>([]);
  const [validated, setValidated] = useState<CartValidateResponse | null>(null);
  const [validating, setValidating] = useState(false);
  const [checkoutLoading, setCheckoutLoading] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem(CART_STORAGE_KEY);
    if (stored) {
      try {
        setCart(JSON.parse(stored));
      } catch {
        setCart([]);
      }
    }
  }, []);

  const saveCart = (newCart: CartItem[]) => {
    setCart(newCart);
    localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(newCart));
    setValidated(null);
  };

  const updateQuantity = (productId: number, quantity: number) => {
    if (quantity <= 0) {
      removeItem(productId);
      return;
    }
    saveCart(cart.map((item) => (item.product_id === productId ? { ...item, quantity } : item)));
  };

  const removeItem = (productId: number) => {
    saveCart(cart.filter((item) => item.product_id !== productId));
  };

  const addItem = (product: { id: number; stock: number }, quantity: number = 1) => {
    const existing = cart.find((item) => item.product_id === product.id);
    if (existing) {
      const newQty = Math.min(existing.quantity + quantity, product.stock);
      updateQuantity(product.id, newQty);
    } else {
      saveCart([...cart, { product_id: product.id, quantity: Math.min(quantity, product.stock) }]);
    }
    success('Added to cart');
  };

  const handleValidate = async () => {
    if (cart.length === 0) return;
    setValidating(true);
    try {
      const response = await cartApi.validate(cart);
      setValidated(response);
      if (!response.valid) {
        showError(response.errors.join(', '));
      } else {
        success('Cart validated successfully');
      }
    } catch (err) {
      showError('Validation failed');
      console.error(err);
    } finally {
      setValidating(false);
    }
  };

  const handleCheckout = async () => {
    if (!validated?.valid || cart.length === 0) return;
    setCheckoutLoading(true);
    try {
      const response = await cartApi.validate(cart);
      if (!response.valid) {
        showError('Cart validation failed before checkout');
        return;
      }
      
      navigate('/checkout', { state: { cart, validated: response } });
    } catch (err) {
      showError('Checkout failed');
      console.error(err);
    } finally {
      setCheckoutLoading(false);
    }
  };

  useEffect(() => {
    const handleAddToCart = (e: CustomEvent) => {
      const { id, stock } = e.detail;
      addItem({ id, stock });
    };
    window.addEventListener('add-to-cart', handleAddToCart as EventListener);
    return () => window.removeEventListener('add-to-cart', handleAddToCart as EventListener);
  }, []);

  const subtotal = validated
    ? (validated.total_cents / 100).toFixed(2)
    : cart.reduce((sum, item) => sum + item.quantity * 0, 0).toFixed(2);

  return (
    <div className="container">
      <h1 style={{ marginBottom: '1.5rem' }}>Shopping Cart</h1>

      <ToastContainer toasts={toasts} />

      {cart.length === 0 ? (
        <div className="empty-state">
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>🛒</div>
          <h3>Your cart is empty</h3>
          <p>Add some products to get started.</p>
          <Link to="/" className="btn btn-primary" style={{ marginTop: '1rem', display: 'inline-block' }}>
            Continue Shopping
          </Link>
        </div>
      ) : (
        <>
          <div className="card">
            <div className="card-header">
              <h2 style={{ fontSize: '1.25rem' }}>Cart Items ({cart.length})</h2>
            </div>
            {cart.map((item) => {
              const validatedItem = validated?.items.find((i) => i.product_id === item.product_id);
              return (
                <CartItemComponent
                  key={item.product_id}
                  item={{
                    product_id: item.product_id,
                    name: validatedItem?.name || `Product ${item.product_id}`,
                    price_cents: validatedItem?.price_cents || 0,
                    quantity: item.quantity,
                    stock: validatedItem?.stock || 0,
                    subtotal_cents: validatedItem?.subtotal_cents || 0,
                  }}
                  onUpdateQuantity={updateQuantity}
                  onRemove={removeItem}
                />
              );
            })}
          </div>

          <div className="card" style={{ marginTop: '1.5rem', maxWidth: '500px' }}>
            <div className="cart-summary">
              <h3 style={{ marginBottom: '1rem' }}>Order Summary</h3>
              <div className="summary-row">
                <span>Subtotal</span>
                <span>${subtotal}</span>
              </div>
              <div className="summary-row summary-total">
                <span>Total</span>
                <span>${subtotal}</span>
              </div>

              <button
                className="btn btn-primary"
                style={{ width: '100%', marginTop: '1.5rem', padding: '1rem' }}
                onClick={handleValidate}
                disabled={validating}
              >
                {validating ? 'Validating...' : 'Validate Cart'}
              </button>

              {validated && (
                <>
                  <div
                    className={`alert ${validated.valid ? 'alert-success' : 'alert-error'}`}
                    style={{ marginTop: '1rem' }}
                  >
                    {validated.valid
                      ? 'Cart is valid and ready for checkout'
                      : validated.errors.join(', ')}
                  </div>
                  {validated.valid && (
                    <button
                      className="btn btn-primary"
                      style={{ width: '100%', marginTop: '1rem', padding: '1rem' }}
                      onClick={handleCheckout}
                      disabled={checkoutLoading}
                    >
                      {checkoutLoading ? 'Proceeding...' : 'Proceed to Checkout'}
                    </button>
                  )}
                </>
              )}
            </div>
          </div>

          <Link to="/" className="btn btn-secondary" style={{ marginTop: '1.5rem', display: 'inline-block' }}>
            ← Continue Shopping
          </Link>
        </>
      )}
    </div>
  );
}