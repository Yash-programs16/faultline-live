interface QuantitySelectorProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  disabled?: boolean;
  style?: React.CSSProperties;
}

export function QuantitySelector({ value, onChange, min = 1, max = 99, disabled = false, style }: QuantitySelectorProps) {
  const increment = () => {
    if (value < max) onChange(value + 1);
  };

  const decrement = () => {
    if (value > min) onChange(value - 1);
  };

  return (
    <div className="quantity-selector" style={style}>
      <button
        className="quantity-btn"
        onClick={decrement}
        disabled={disabled || value <= min}
        aria-label="Decrease quantity"
      >
        −
      </button>
      <span className="quantity-value" aria-live="polite">{value}</span>
      <button
        className="quantity-btn"
        onClick={increment}
        disabled={disabled || value >= max}
        aria-label="Increase quantity"
      >
        +
      </button>
    </div>
  );
}