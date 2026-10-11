import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { SortKey } from '@/data/types'
import Breadcrumb from '@/components/ui/Breadcrumb'
import CatalogView from '@/components/catalog/CatalogView'

export default function CatalogPage() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const sort = (params.get('sort') as SortKey) || 'relevance'

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('catalog.title') }]} />
      <h1 className="mb-8 mt-4 text-2xl font-bold text-heading sm:text-3xl">{t('catalog.title')}</h1>
      <CatalogView initialSort={sort} />
    </div>
  )
}
