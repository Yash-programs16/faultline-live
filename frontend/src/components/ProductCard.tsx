import { Link } from 'react-router-dom';
import type { Product } from '../types';

interface ProductCardProps {
  product: Product;
  onAddToCart: (product: Product) => void;
}

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

export function ProductCard({ product, onAddToCart }: ProductCardProps) {
  const icon = PRODUCT_ICONS[product.name] || '📦';
  const price = (product.price_cents / 100).toFixed(2);
  const inStock = product.stock > 0;

  return (
    <article className="card product-card">
      <Link to={`/products/${product.id}`} className="product-image" aria-label={`View ${product.name}`}>
        <span role="img" aria-label={product.name}>{icon}</span>
      </Link>
      <div className="product-info">
        <Link to={`/products/${product.id}`}>
          <h3 className="product-title">{product.name}</h3>
        </Link>
        <p className="product-description">{product.description}</p>
        <div className="product-footer">
          <span className="product-price">${price}</span>
          <button
            className="btn btn-primary"
            onClick={() => onAddToCart(product)}
            disabled={!inStock}
            aria-label={inStock ? `Add ${product.name} to cart` : `${product.name} is out of stock`}
          >
            {inStock ? 'Add to Cart' : 'Out of Stock'}
          </button>
        </div>
      </div>
    </article>
  );
}