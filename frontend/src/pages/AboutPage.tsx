import { useTranslation } from 'react-i18next'
import Breadcrumb from '@/components/ui/Breadcrumb'
import IconBase from '@/components/ui/IconBase'

const STATS = [
  { key: 'about.stats.customers', value: '50k+' },
  { key: 'about.stats.products', value: '12k+' },
  { key: 'about.stats.countries', value: '18' },
  { key: 'about.stats.rating', value: '4.8★' },
]

const VALUES = [
  { icon: 'shield-check', title: 'about.values.quality', desc: 'about.values.qualityDesc' },
  { icon: 'users', title: 'about.values.trust', desc: 'about.values.trustDesc' },
  { icon: 'truck', title: 'about.values.speed', desc: 'about.values.speedDesc' },
]

export default function AboutPage() {
  const { t } = useTranslation()
  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('common.about') }]} />

      <div className="mt-8 text-center">
        <h1 className="text-3xl font-bold text-heading sm:text-title-md">{t('about.title')}</h1>
        <p className="mx-auto mt-3 max-w-xl text-base text-muted">{t('about.subtitle')}</p>
      </div>

      <div className="mt-12 grid grid-cols-2 gap-4 sm:grid-cols-4">
        {STATS.map((s) => (
          <div key={s.key} className="surface-card rounded-2xl border p-6 text-center">
            <p className="text-2xl font-bold text-brand-600 sm:text-3xl">{s.value}</p>
            <p className="mt-1 text-sm text-muted">{t(s.key)}</p>
          </div>
        ))}
      </div>

      <div className="mt-12 grid gap-8 md:grid-cols-2">
        <div className="surface-card rounded-2xl border p-7">
          <h2 className="text-xl font-bold text-heading">{t('about.storyTitle')}</h2>
          <p className="mt-3 text-sm leading-relaxed text-body">{t('about.story')}</p>
        </div>
        <div className="surface-card rounded-2xl border p-7">
          <h2 className="text-xl font-bold text-heading">{t('about.missionTitle')}</h2>
          <p className="mt-3 text-sm leading-relaxed text-body">{t('about.mission')}</p>
        </div>
      </div>

      <div className="mt-12">
        <h2 className="mb-6 text-center text-2xl font-bold text-heading">{t('about.valuesTitle')}</h2>
        <div className="grid gap-4 sm:grid-cols-3">
          {VALUES.map((v) => (
            <div key={v.icon} className="surface-card rounded-2xl border p-6 text-center">
              <span className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-50 text-brand-500 dark:bg-brand-500/15">
                <IconBase name={v.icon} size={26} />
              </span>
              <h3 className="mt-4 font-semibold text-heading">{t(v.title)}</h3>
              <p className="mt-1 text-sm text-muted">{t(v.desc)}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
