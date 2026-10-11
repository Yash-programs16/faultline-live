import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { productsApi } from '../lib/api';
import { Loading } from '../components/Loading';
import { QuantitySelector } from '../components/QuantitySelector';
import type { Product } from '../types';

const PRODUCT_ICONS: Record<string, string> = {
  'Wireless Headphones': '🎧',
  'Mechanical Keyboard': '⌨️',
  'USB-C Hub': '🔌',
  'Ergonomic Mouse': '🖱️',
  'Laptop Stand': '💻',
  'Webcam Cover': '📷',
  'Monitor Light': '💡',
  'Cable Organizer': '📦',
};

export function ProductDetail() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    if (!id) return;
    const productId = parseInt(id, 10);
    async function fetchProduct() {
      try {
        setLoading(true);
        const response = await productsApi.get(productId);
        setProduct(response.product);
        setQuantity(1);
      } catch (err: unknown) {
        if (err instanceof Error && 'status' in err && (err as { status: number }).status === 404) {
          setError('Product not found');
        } else {
          setError('Failed to load product');
        }
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    fetchProduct();
  }, [id]);

  const handleAddToCart = async () => {
    if (!product) return;
    setAdding(true);
    try {
      const event = new CustomEvent('add-to-cart', { detail: { ...product, quantity } });
      window.dispatchEvent(event);
    } finally {
      setAdding(false);
    }
  };

  if (loading) {
    return (
      <div className="container">
        <Loading size="lg" text="Loading product..." />
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="container">
        <div className="alert alert-error">{error || 'Product not found'}</div>
        <Link to="/store" className="btn btn-primary">Back to Store</Link>
      </div>
    );
  }

  const icon = PRODUCT_ICONS[product.name] || '📦';
  const price = (product.price_cents / 100).toFixed(2);
  const inStock = product.stock > 0;

  return (
    <div className="container">
      <Link to="/store" className="btn btn-secondary" style={{ marginBottom: '1.5rem', display: 'inline-flex' }}>
        ← Back to Products
      </Link>

      <div style={{ display: 'grid', gap: '2rem', gridTemplateColumns: '1fr 1fr', maxWidth: '900px' }}>
        <div className="card" style={{ padding: '2rem', textAlign: 'center' }}>
          <div style={{ fontSize: '6rem', marginBottom: '1rem' }} role="img" aria-label={product.name}>
            {icon}
          </div>
        </div>

        <div className="card">
          <div className="card-body" style={{ padding: '2rem' }}>
            <h1 style={{ fontSize: '2rem', fontWeight: '700', marginBottom: '1rem' }}>{product.name}</h1>
            <p style={{ color: '#6b7280', marginBottom: '1.5rem', fontSize: '1.125rem' }}>{product.description}</p>
            
            <div style={{ fontSize: '2.5rem', fontWeight: '700', color: '#1a1a2e', marginBottom: '1.5rem' }}>
              ${price}
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <span className={`badge ${inStock ? 'badge-success' : 'badge-danger'}`}>
                {inStock ? `In Stock (${product.stock} available)` : 'Out of Stock'}
              </span>
            </div>

            {inStock && (
              <div style={{ marginBottom: '1.5rem' }}>
                <label className="form-label" htmlFor="quantity">Quantity</label>
                <QuantitySelector
                  value={quantity}
                  onChange={setQuantity}
                  max={product.stock}
                  style={{ maxWidth: '200px' }}
                />
              </div>
            )}

            <button
              className="btn btn-primary"
              style={{ width: '100%', padding: '1rem', fontSize: '1.125rem' }}
              onClick={handleAddToCart}
              disabled={!inStock || adding}
            >
              {adding ? 'Adding...' : inStock ? 'Add to Cart' : 'Out of Stock'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}