import { useTranslation } from 'react-i18next'
import { formatPrice, discountPercent } from '@/lib/format'

interface Props {
  price: number
  priceOld?: number
  size?: 'sm' | 'md' | 'lg'
  showBadge?: boolean
}

const SIZES = {
  sm: 'text-base',
  md: 'text-lg',
  lg: 'text-2xl',
}

export default function PriceTag({ price, priceOld, size = 'md', showBadge = false }: Props) {
  const { i18n } = useTranslation()
  const lang = i18n.language
  const pct = discountPercent(price, priceOld)

  return (
    <div className="flex flex-wrap items-baseline gap-2">
      <span className={`font-bold text-heading ${SIZES[size]}`}>{formatPrice(price, lang)}</span>
      {priceOld && priceOld > price && (
        <span className="text-sm text-muted line-through">{formatPrice(priceOld, lang)}</span>
      )}
      {showBadge && pct > 0 && (
        <span className="rounded-full bg-error-50 px-2 py-0.5 text-xs font-semibold text-error-600 dark:bg-error-500/15 dark:text-error-500">
          -{pct}%
        </span>
      )}
    </div>
  )
}
