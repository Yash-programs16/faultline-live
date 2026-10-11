import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import IconBase from '@/components/ui/IconBase'

export default function RegisterPage() {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const register = useAuthStore((s) => s.register)
  const notify = useUiStore((s) => s.notifySuccess)
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    if (password !== confirm) {
      setError(t('auth.passwordMismatch'))
      return
    }
    setLoading(true)
    const ok = await register(name, email, password)
    setLoading(false)
    if (!ok) {
      setError(t('auth.loginError'))
      return
    }
    notify(t('auth.registerSuccess'))
    navigate('/conta')
  }

  const inputCls =
    'h-11 w-full rounded-lg border border-themed bg-gray-50 pl-10 pr-3 text-sm text-body outline-none focus:border-brand-300 focus-ring dark:bg-gray-900'

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-gradient-to-br from-brand-50 via-white to-brand-100 px-4 py-10 dark:from-gray-900 dark:via-gray-dark dark:to-gray-900">
      <div className="w-full max-w-md">
        <div className="mb-6 flex flex-col items-center">
          <Link to="/" className="flex h-14 w-14 items-center justify-center rounded-2xl bg-brand-500 text-white shadow-theme-md">
            <span className="text-2xl font-bold">L</span>
          </Link>
          <h1 className="mt-5 text-title-sm font-bold text-heading">{t('auth.registerTitle')}</h1>
          <p className="mt-2 text-sm text-muted">{t('auth.registerSubtitle')}</p>
        </div>

        <div className="surface-card rounded-2xl border p-6 shadow-theme-md md:p-8">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="mb-2 block text-sm font-medium text-heading">{t('auth.name')}</label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="user" size={18} />
                </span>
                <input type="text" required value={name} onChange={(e) => setName(e.target.value)} className={inputCls} />
              </div>
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-heading">{t('auth.email')}</label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="mail" size={18} />
                </span>
                <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className={inputCls} />
              </div>
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-heading">{t('auth.password')}</label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="lock" size={18} />
                </span>
                <input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} className={inputCls} />
              </div>
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-heading">{t('auth.confirmPassword')}</label>
              <div className="relative">
                <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="lock" size={18} />
                </span>
                <input type="password" required value={confirm} onChange={(e) => setConfirm(e.target.value)} className={inputCls} />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="flex h-11 w-full items-center justify-center rounded-lg bg-brand-500 text-sm font-medium text-white shadow-theme-xs transition hover:bg-brand-600 focus-ring disabled:opacity-60"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                  {t('auth.registering')}
                </span>
              ) : (
                t('auth.registerCta')
              )}
            </button>

            {error && (
              <p className="rounded-lg bg-error-50 px-4 py-2 text-center text-sm text-error-600 dark:bg-error-500/15 dark:text-error-500">
                {error}
              </p>
            )}
          </form>
        </div>

        <p className="mt-6 text-center text-sm text-muted">
          {t('auth.hasAccount')}{' '}
          <Link to="/login" className="font-medium text-brand-500 hover:text-brand-600">
            {t('auth.loginCta')}
          </Link>
        </p>
      </div>
    </div>
  )
}
