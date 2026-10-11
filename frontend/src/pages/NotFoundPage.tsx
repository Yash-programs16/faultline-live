import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import IconBase from '@/components/ui/IconBase'

export default function NotFoundPage() {
  const { t } = useTranslation()
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center px-4 text-center">
      <p className="text-title-xl font-bold text-brand-500 sm:text-title-2xl">404</p>
      <h1 className="mt-2 text-2xl font-bold text-heading sm:text-3xl">{t('notFound.title')}</h1>
      <p className="mx-auto mt-3 max-w-md text-sm text-muted">{t('notFound.subtitle')}</p>
      <Link
        to="/"
        className="mt-8 flex h-12 items-center gap-2 rounded-lg bg-brand-500 px-6 text-sm font-medium text-white transition hover:bg-brand-600"
      >
        <IconBase name="home" size={18} />
        {t('notFound.cta')}
      </Link>
    </div>
  )
}
