import { QuantitySelector } from './QuantitySelector';
import type { CartItemDetail } from '../types';

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

interface CartItemProps {
  item: CartItemDetail;
  onUpdateQuantity: (productId: number, quantity: number) => void;
  onRemove: (productId: number) => void;
}

export function CartItemComponent({ item, onUpdateQuantity, onRemove }: CartItemProps) {
  const icon = PRODUCT_ICONS[item.name] || '📦';
  const price = (item.price_cents / 100).toFixed(2);
  const subtotal = (item.subtotal_cents / 100).toFixed(2);

  return (
    <div className="cart-item">
      <div className="cart-item-image" aria-hidden="true">{icon}</div>
      <div className="cart-item-details">
        <div className="cart-item-title">{item.name}</div>
        <div className="cart-item-price">${price} each</div>
        <QuantitySelector
          value={item.quantity}
          onChange={(qty) => onUpdateQuantity(item.product_id, qty)}
          max={item.stock}
          aria-label={`Quantity for ${item.name}`}
        />
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '0.5rem' }}>
        <span style={{ fontWeight: 600, fontSize: '1.125rem' }}>${subtotal}</span>
        <button
          className="btn btn-secondary"
          style={{ padding: '0.375rem 0.75rem', fontSize: '0.75rem' }}
          onClick={() => onRemove(item.product_id)}
          aria-label={`Remove ${item.name} from cart`}
        >
          Remove
        </button>
      </div>
    </div>
  );
}