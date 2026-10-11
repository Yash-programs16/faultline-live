import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { products as allProducts } from '@/data/products'
import { filterProducts, sortProducts, paginate } from '@/data/query'
import type { SortKey } from '@/data/types'
import ProductCard from '@/components/product/ProductCard'
import FilterSidebar, { defaultFilters, type CatalogFilterState } from './FilterSidebar'
import SortDropdown from './SortDropdown'
import Pagination from '@/components/ui/Pagination'
import EmptyState from '@/components/ui/EmptyState'
import IconBase from '@/components/ui/IconBase'

interface Props {
  /** restrict to a single category slug (category page) */
  fixedCategory?: string
  /** text query (search page) */
  query?: string
  initialSort?: SortKey
}

const PER_PAGE = 9

export default function CatalogView({ fixedCategory, query, initialSort = 'relevance' }: Props) {
  const { t } = useTranslation()
  const [filters, setFilters] = useState<CatalogFilterState>({
    ...defaultFilters,
    category: fixedCategory ?? '',
  })
  const [sort, setSort] = useState<SortKey>(initialSort)
  const [page, setPage] = useState(1)
  const [view, setView] = useState<'grid' | 'list'>('grid')
  const [mobileFilters, setMobileFilters] = useState(false)

  const filtered = useMemo(() => {
    const list = filterProducts(allProducts, {
      query,
      category: fixedCategory ?? (filters.category || undefined),
      brands: filters.brands,
      maxPrice: filters.maxPrice,
      minRating: filters.minRating || undefined,
    })
    return sortProducts(list, sort)
  }, [filters, sort, fixedCategory, query])

  const totalPages = Math.ceil(filtered.length / PER_PAGE)
  const current = Math.min(page, Math.max(1, totalPages))
  const pageItems = paginate(filtered, current, PER_PAGE)

  const handleFilterChange = (next: CatalogFilterState) => {
    setFilters(next)
    setPage(1)
  }

  return (
    <div className="grid gap-8 lg:grid-cols-[260px_1fr]">
      {/* Sidebar (desktop) */}
      <aside className="hidden lg:block">
        <div className="surface-card sticky top-32 rounded-2xl border p-5">
          <FilterSidebar filters={filters} onChange={handleFilterChange} lockCategory={!!fixedCategory} />
        </div>
      </aside>

      <div>
        {/* Toolbar */}
        <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-muted">
            {t('catalog.showingResults', { shown: pageItems.length, total: filtered.length })}
          </p>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setMobileFilters(true)}
              className="flex h-10 items-center gap-2 rounded-lg border border-themed px-3.5 text-sm font-medium text-body lg:hidden"
            >
              <IconBase name="sliders" size={16} />
              {t('catalog.filters')}
            </button>
            <div className="hidden items-center rounded-lg border border-themed sm:flex">
              <button
                type="button"
                onClick={() => setView('grid')}
                aria-label={t('catalog.gridView')}
                className={`flex h-10 w-10 items-center justify-center rounded-l-lg ${view === 'grid' ? 'bg-brand-50 text-brand-600 dark:bg-brand-500/15' : 'text-muted'}`}
              >
                <IconBase name="grid" size={18} />
              </button>
              <button
                type="button"
                onClick={() => setView('list')}
                aria-label={t('catalog.listView')}
                className={`flex h-10 w-10 items-center justify-center rounded-r-lg ${view === 'list' ? 'bg-brand-50 text-brand-600 dark:bg-brand-500/15' : 'text-muted'}`}
              >
                <IconBase name="list" size={18} />
              </button>
            </div>
            <SortDropdown value={sort} onChange={(s) => { setSort(s); setPage(1) }} />
          </div>
        </div>

        {/* Grid / list */}
        {pageItems.length === 0 ? (
          <EmptyState icon="search" title={t('catalog.noResults')} description={t('catalog.noResultsDesc')} />
        ) : view === 'grid' ? (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            {pageItems.map((p) => (
              <ProductCard key={p.id} product={p} />
            ))}
          </div>
        ) : (
          <div className="space-y-4">
            {pageItems.map((p) => (
              <ProductCard key={p.id} product={p} layout="list" />
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="mt-10">
            <Pagination page={current} totalPages={totalPages} onChange={setPage} />
          </div>
        )}
      </div>

      {/* Mobile filter drawer */}
      {mobileFilters && (
        <div className="fixed inset-0 z-[60] lg:hidden">
          <div className="absolute inset-0 bg-gray-900/40" onClick={() => setMobileFilters(false)} />
          <div className="surface-card absolute left-0 top-0 h-full w-80 max-w-[85%] overflow-y-auto border-r p-5 shadow-theme-xl">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="font-semibold text-heading">{t('catalog.filters')}</h2>
              <button type="button" onClick={() => setMobileFilters(false)} className="text-muted hover:text-heading">
                <IconBase name="x" size={20} />
              </button>
            </div>
            <FilterSidebar filters={filters} onChange={handleFilterChange} lockCategory={!!fixedCategory} />
          </div>
        </div>
      )}
    </div>
  )
}
