interface Props {
  rating: number
  size?: number
  showValue?: boolean
  count?: number
  className?: string
}

/** Read-only star rating with half-star support, rendered as SVG. */
export default function RatingStars({ rating, size = 16, showValue = false, count, className = '' }: Props) {
  const pct = Math.max(0, Math.min(100, (rating / 5) * 100))
  return (
    <span className={`inline-flex items-center gap-1.5 ${className}`}>
      <span className="relative inline-block" style={{ width: size * 5, height: size }} aria-hidden>
        <span className="absolute inset-0 flex text-gray-300 dark:text-gray-600">
          {[0, 1, 2, 3, 4].map((i) => (
            <Star key={i} size={size} />
          ))}
        </span>
        <span
          className="absolute inset-0 flex overflow-hidden text-warning-400"
          style={{ width: `${pct}%` }}
        >
          {[0, 1, 2, 3, 4].map((i) => (
            <Star key={i} size={size} filled />
          ))}
        </span>
      </span>
      {showValue && <span className="text-sm font-medium text-heading">{rating.toFixed(1)}</span>}
      {count != null && <span className="text-sm text-muted">({count})</span>}
    </span>
  )
}

function Star({ size, filled = false }: { size: number; filled?: boolean }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill={filled ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth={filled ? 0 : 1.5}>
      <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
    </svg>
  )
}
