import { useTranslation } from 'react-i18next'
import { categories } from '@/data/categories'
import { brands, priceBounds } from '@/data/products'
import RatingStars from '@/components/ui/RatingStars'

export interface CatalogFilterState {
  category: string
  brands: string[]
  maxPrice: number
  minRating: number
}

export const defaultFilters: CatalogFilterState = {
  category: '',
  brands: [],
  maxPrice: priceBounds.max,
  minRating: 0,
}

interface Props {
  filters: CatalogFilterState
  onChange: (next: CatalogFilterState) => void
  /** when true the category facet is hidden (category page) */
  lockCategory?: boolean
}

export default function FilterSidebar({ filters, onChange, lockCategory = false }: Props) {
  const { t, i18n } = useTranslation()

  const toggleBrand = (brand: string) => {
    const next = filters.brands.includes(brand)
      ? filters.brands.filter((b) => b !== brand)
      : [...filters.brands, brand]
    onChange({ ...filters, brands: next })
  }

  const fmt = (v: number) =>
    new Intl.NumberFormat(i18n.language === 'pt-BR' ? 'pt-BR' : i18n.language, {
      style: 'currency',
      currency: i18n.language === 'en' ? 'USD' : i18n.language === 'es' ? 'EUR' : 'BRL',
      maximumFractionDigits: 0,
    }).format(v)

  return (
    <div className="space-y-7">
      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-heading">{t('catalog.filters')}</h2>
        <button
          type="button"
          onClick={() => onChange({ ...defaultFilters, category: lockCategory ? filters.category : '' })}
          className="text-sm font-medium text-brand-600 hover:text-brand-700"
        >
          {t('catalog.clearFilters')}
        </button>
      </div>

      {!lockCategory && (
        <div>
          <h3 className="mb-3 text-sm font-semibold text-heading">{t('catalog.category')}</h3>
          <div className="space-y-1">
            <button
              type="button"
              onClick={() => onChange({ ...filters, category: '' })}
              className={`block w-full rounded-lg px-3 py-1.5 text-left text-sm transition ${
                filters.category === ''
                  ? 'bg-brand-50 font-medium text-brand-600 dark:bg-brand-500/15'
                  : 'text-body hover:bg-gray-50 dark:hover:bg-gray-800'
              }`}
            >
              {t('catalog.allCategories')}
            </button>
            {categories.map((c) => (
              <button
                key={c.slug}
                type="button"
                onClick={() => onChange({ ...filters, category: c.slug })}
                className={`block w-full rounded-lg px-3 py-1.5 text-left text-sm transition ${
                  filters.category === c.slug
                    ? 'bg-brand-50 font-medium text-brand-600 dark:bg-brand-500/15'
                    : 'text-body hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
              >
                {t(`categories.${c.slug}.name`)}
              </button>
            ))}
          </div>
        </div>
      )}

      <div>
        <h3 className="mb-3 text-sm font-semibold text-heading">{t('catalog.price')}</h3>
        <input
          type="range"
          min={priceBounds.min}
          max={priceBounds.max}
          step={10}
          value={filters.maxPrice}
          onChange={(e) => onChange({ ...filters, maxPrice: Number(e.target.value) })}
          className="w-full"
        />
        <div className="mt-2 flex justify-between text-sm text-muted">
          <span>{fmt(priceBounds.min)}</span>
          <span className="font-medium text-heading">{fmt(filters.maxPrice)}</span>
        </div>
      </div>

      <div>
        <h3 className="mb-3 text-sm font-semibold text-heading">{t('catalog.brand')}</h3>
        <div className="max-h-56 space-y-2 overflow-y-auto pr-1">
          {brands.map((b) => (
            <label key={b} className="flex cursor-pointer items-center gap-2.5 text-sm text-body">
              <input
                type="checkbox"
                checked={filters.brands.includes(b)}
                onChange={() => toggleBrand(b)}
                className="h-4 w-4 rounded border-themed accent-brand-500"
              />
              {b}
            </label>
          ))}
        </div>
      </div>

      <div>
        <h3 className="mb-3 text-sm font-semibold text-heading">{t('catalog.rating')}</h3>
        <div className="space-y-1">
          {[4, 3, 2].map((r) => (
            <button
              key={r}
              type="button"
              onClick={() => onChange({ ...filters, minRating: filters.minRating === r ? 0 : r })}
              className={`flex w-full items-center gap-2 rounded-lg px-3 py-1.5 text-left text-sm transition ${
                filters.minRating === r
                  ? 'bg-brand-50 dark:bg-brand-500/15'
                  : 'hover:bg-gray-50 dark:hover:bg-gray-800'
              }`}
            >
              <RatingStars rating={r} size={14} />
              <span className="text-muted">{t('catalog.ratingUp', { count: r })}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
