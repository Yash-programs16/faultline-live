import { useState, useEffect } from 'react';
import { productsApi } from '../lib/api';
import { ProductCard } from '../components/ProductCard';
import { Loading } from '../components/Loading';
import type { Product } from '../types';

export function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchProducts() {
      try {
        setLoading(true);
        const response = await productsApi.list();
        setProducts(response.products);
      } catch (err) {
        setError('Failed to load products. Please check if the backend is running.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    }
    fetchProducts();
  }, []);

  const handleAddToCart = (product: Product) => {
    const event = new CustomEvent('add-to-cart', { detail: product });
    window.dispatchEvent(event);
  };

  if (loading) {
    return (
      <div className="container">
        <Loading size="lg" text="Loading products..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container">
        <div className="alert alert-error">{error}</div>
        <button className="btn btn-primary" onClick={() => window.location.reload()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="container">
      <section style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: '700', marginBottom: '0.5rem' }}>Our Products</h1>
        <p style={{ color: '#6b7280' }}>Browse our selection of quality tech accessories</p>
      </section>

      {products.length === 0 ? (
        <div className="empty-state">
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>📦</div>
          <h3>No products available</h3>
          <p>Check back later or contact support.</p>
        </div>
      ) : (
        <div className="grid grid-4" style={{ marginTop: '1.5rem' }}>
          {products.map((product) => (
            <ProductCard key={product.id} product={product} onAddToCart={handleAddToCart} />
          ))}
        </div>
      )}
    </div>
  );
}