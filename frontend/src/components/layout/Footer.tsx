import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { categories } from '@/data/categories'
import { useUiStore } from '@/stores/ui'
import IconBase from '@/components/ui/IconBase'

const SUPPORT_LINKS = ['footer.help', 'footer.shipping', 'footer.returns', 'footer.faq']
const COMPANY_LINKS = ['common.about', 'footer.careers', 'footer.press', 'footer.privacy', 'footer.terms']
const SOCIAL = ['globe', 'mail', 'users', 'star']
const PAYMENTS = ['credit-card', 'dollar', 'gift', 'shield-check']

export default function Footer() {
  const { t } = useTranslation()
  const notify = useUiStore((s) => s.notifySuccess)
  const [email, setEmail] = useState('')

  function subscribe(e: FormEvent) {
    e.preventDefault()
    if (!email.trim()) return
    notify(t('home.newsletterSuccess'))
    setEmail('')
  }

  return (
    <footer className="border-t border-themed bg-white dark:bg-gray-dark">
      <div className="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-5">
          <div className="col-span-2">
            <div className="flex items-center gap-2.5">
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand-500 text-lg font-bold text-white shadow-theme-md">
                L
              </span>
              <span className="text-lg font-bold text-heading">{t('brand')}</span>
            </div>
            <p className="mt-4 max-w-xs text-sm text-muted">{t('footer.tagline')}</p>
            <div className="mt-5 flex gap-2">
              {SOCIAL.map((icon) => (
                <a
                  key={icon}
                  href="#"
                  aria-label={icon}
                  className="flex h-10 w-10 items-center justify-center rounded-full border border-themed text-muted transition hover:bg-gray-50 hover:text-brand-600 dark:hover:bg-gray-800"
                >
                  <IconBase name={icon} size={18} />
                </a>
              ))}
            </div>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-heading">{t('footer.shop')}</h3>
            <ul className="mt-4 space-y-3">
              {categories.slice(0, 5).map((c) => (
                <li key={c.slug}>
                  <Link to={`/categoria/${c.slug}`} className="text-sm text-muted transition hover:text-brand-600">
                    {t(`categories.${c.slug}.name`)}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-heading">{t('footer.support')}</h3>
            <ul className="mt-4 space-y-3">
              {SUPPORT_LINKS.map((key) => (
                <li key={key}>
                  <a href="#" className="text-sm text-muted transition hover:text-brand-600">
                    {t(key)}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h3 className="text-sm font-semibold text-heading">{t('footer.company')}</h3>
            <ul className="mt-4 space-y-3">
              {COMPANY_LINKS.map((key) => (
                <li key={key}>
                  {key === 'common.about' ? (
                    <Link to="/sobre" className="text-sm text-muted transition hover:text-brand-600">
                      {t(key)}
                    </Link>
                  ) : (
                    <a href="#" className="text-sm text-muted transition hover:text-brand-600">
                      {t(key)}
                    </a>
                  )}
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* newsletter */}
        <div className="mt-12 grid gap-4 rounded-2xl border border-themed bg-gray-50 p-6 dark:bg-gray-900/40 md:grid-cols-2 md:items-center">
          <div>
            <h3 className="text-base font-semibold text-heading">{t('footer.newsletterTitle')}</h3>
            <p className="mt-1 text-sm text-muted">{t('footer.newsletterDesc')}</p>
          </div>
          <form onSubmit={subscribe} className="flex gap-2">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t('home.newsletterPlaceholder')}
              className="h-11 flex-1 rounded-lg border border-themed bg-white px-4 text-sm text-body outline-none focus:border-brand-300 focus-ring dark:bg-gray-900"
            />
            <button
              type="submit"
              className="h-11 rounded-lg bg-brand-500 px-5 text-sm font-medium text-white transition hover:bg-brand-600"
            >
              {t('home.newsletterCta')}
            </button>
          </form>
        </div>

        <div className="mt-10 flex flex-col items-center justify-between gap-4 border-t border-themed pt-6 sm:flex-row">
          <p className="text-sm text-muted">{t('footer.rights')}</p>
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted">{t('footer.payments')}:</span>
            {PAYMENTS.map((icon) => (
              <span
                key={icon}
                className="flex h-8 w-10 items-center justify-center rounded-md border border-themed text-muted"
              >
                <IconBase name={icon} size={16} />
              </span>
            ))}
          </div>
        </div>
      </div>
    </footer>
  )
}
