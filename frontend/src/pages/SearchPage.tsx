import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import Breadcrumb from '@/components/ui/Breadcrumb'
import CatalogView from '@/components/catalog/CatalogView'

export default function SearchPage() {
  const { t } = useTranslation()
  const [params] = useSearchParams()
  const query = params.get('q') ?? ''

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('common.search') }]} />
      <h1 className="mb-8 mt-4 text-2xl font-bold text-heading sm:text-3xl">
        {t('search.title', { query })}
      </h1>
      <CatalogView query={query} key={query} />
    </div>
  )
}
