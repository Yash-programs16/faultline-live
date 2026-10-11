import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore, DEMO_CREDENTIALS } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import IconBase from '@/components/ui/IconBase'

export default function LoginPage() {
  const { t } = useTranslation()
  const [email, setEmail] = useState(DEMO_CREDENTIALS.email)
  const [password, setPassword] = useState(DEMO_CREDENTIALS.password)
  const [showPassword, setShowPassword] = useState(false)
  const [remember, setRemember] = useState(true)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const login = useAuthStore((s) => s.login)
  const notify = useUiStore((s) => s.notifySuccess)
  const navigate = useNavigate()
  const [params] = useSearchParams()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    const ok = await login(email, password)
    setLoading(false)
    if (!ok) {
      setError(t('auth.loginError'))
      return
    }
    notify(t('auth.loginSuccess'))
    navigate(params.get('redirect') || '/conta')
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-gradient-to-br from-brand-50 via-white to-brand-100 px-4 dark:from-gray-900 dark:via-gray-dark dark:to-gray-900">
      <div className="absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute -top-32 -right-32 h-96 w-96 rounded-full bg-brand-200/40 blur-3xl dark:bg-brand-500/10" />
        <div className="absolute -bottom-32 -left-32 h-96 w-96 rounded-full bg-warning-100/60 blur-3xl dark:bg-brand-700/10" />
      </div>

      <div className="w-full max-w-md">
        <div className="mb-6 flex flex-col items-center">
          <Link to="/" className="flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-500 text-white shadow-theme-md">
            <span className="text-2xl font-bold">L</span>
          </Link>
          <h1 className="mt-5 text-title-sm font-bold text-heading">{t('auth.loginTitle')}</h1>
          <p className="mt-2 text-sm text-muted">{t('auth.loginSubtitle')}</p>
        </div>

        <div className="surface-card rounded-2xl border p-6 shadow-theme-md md:p-8">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="mb-2 block text-sm font-medium text-heading" htmlFor="email">
                {t('auth.email')}
              </label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="mail" size={18} />
                </span>
                <input
                  id="email"
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="seu@email.com"
                  className="h-11 w-full rounded-lg border border-themed bg-gray-50 pl-10 pr-3 text-sm text-body outline-none focus:border-brand-300 focus-ring dark:bg-gray-900"
                />
              </div>
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-heading" htmlFor="password">
                {t('auth.password')}
              </label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="lock" size={18} />
                </span>
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••"
                  className="h-11 w-full rounded-lg border border-themed bg-gray-50 pl-10 pr-10 text-sm text-body outline-none focus:border-brand-300 focus-ring dark:bg-gray-900"
                />
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                  onClick={() => setShowPassword((s) => !s)}
                >
                  <IconBase name={showPassword ? 'eye-off' : 'eye'} size={18} />
                </button>
              </div>
            </div>

            <div className="flex items-center justify-between text-sm">
              <label className="flex items-center gap-2 text-body">
                <input
                  type="checkbox"
                  checked={remember}
                  onChange={(e) => setRemember(e.target.checked)}
                  className="h-4 w-4 rounded border-themed accent-brand-500"
                />
                {t('auth.remember')}
              </label>
              <a href="#" className="font-medium text-brand-500 hover:text-brand-600">
                {t('auth.forgot')}
              </a>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="flex h-11 w-full items-center justify-center rounded-lg bg-brand-500 text-sm font-medium text-white shadow-theme-xs transition hover:bg-brand-600 focus-ring disabled:cursor-not-allowed disabled:opacity-60"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                  {t('auth.loggingIn')}
                </span>
              ) : (
                t('auth.loginCta')
              )}
            </button>

            {error && (
              <p className="rounded-lg bg-error-50 px-4 py-2 text-center text-sm text-error-600 dark:bg-error-500/15 dark:text-error-500">
                {error}
              </p>
            )}
          </form>

          <div className="mt-6 rounded-lg border border-dashed border-themed bg-gray-50/60 p-3 text-xs text-muted dark:bg-gray-900/40">
            <p className="font-medium text-body">{t('auth.demoTitle')}</p>
            <p>{DEMO_CREDENTIALS.email} / {DEMO_CREDENTIALS.password}</p>
            <p className="mt-0.5">{t('auth.demoHint')}</p>
          </div>
        </div>

        <p className="mt-6 text-center text-sm text-muted">
          {t('auth.noAccount')}{' '}
          <Link to="/registro" className="font-medium text-brand-500 hover:text-brand-600">
            {t('auth.registerCta')}
          </Link>
        </p>
      </div>
    </div>
  )
}
