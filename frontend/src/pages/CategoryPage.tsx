import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { findCategory } from '@/data/categories'
import Breadcrumb from '@/components/ui/Breadcrumb'
import CatalogView from '@/components/catalog/CatalogView'
import NotFoundPage from './NotFoundPage'

export default function CategoryPage() {
  const { t } = useTranslation()
  const { slug = '' } = useParams()
  const category = findCategory(slug)

  if (!category) return <NotFoundPage />

  const name = t(`categories.${slug}.name`)

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb
        items={[
          { label: t('breadcrumb.home'), to: '/' },
          { label: t('common.products'), to: '/produtos' },
          { label: name },
        ]}
      />
      <div className="mb-8 mt-4">
        <h1 className="text-2xl font-bold text-heading sm:text-3xl">{name}</h1>
        <p className="mt-1 text-sm text-muted">{t(`categories.${slug}.desc`)}</p>
      </div>
      <CatalogView fixedCategory={slug} key={slug} />
    </div>
  )
}
