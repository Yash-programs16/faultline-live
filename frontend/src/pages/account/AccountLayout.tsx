import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import Breadcrumb from '@/components/ui/Breadcrumb'
import IconBase from '@/components/ui/IconBase'

const LINKS = [
  { to: '/conta', icon: 'grid', key: 'account.overview', end: true },
  { to: '/conta/pedidos', icon: 'package', key: 'account.orders', end: false },
  { to: '/conta/enderecos', icon: 'map-pin', key: 'account.addresses', end: false },
  { to: '/conta/favoritos', icon: 'heart', key: 'account.wishlist', end: false },
]

export default function AccountLayout() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const notify = useUiStore((s) => s.notify)

  function handleLogout() {
    logout()
    notify(t('account.logoutSuccess'), 'info')
    navigate('/store')
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('account.title') }]} />
      <h1 className="mb-8 mt-4 text-2xl font-bold text-heading sm:text-3xl">{t('account.title')}</h1>

      <div className="grid gap-8 lg:grid-cols-[260px_1fr]">
        <aside>
          <div className="surface-card rounded-2xl border p-5">
            <div className="flex items-center gap-3 border-b border-themed pb-4">
              <span className="flex h-12 w-12 items-center justify-center rounded-full bg-brand-500 text-lg font-semibold text-white">
                {user?.name?.charAt(0).toUpperCase() ?? 'U'}
              </span>
              <div className="min-w-0">
                <p className="truncate font-medium text-heading">{user?.name}</p>
                <p className="truncate text-xs text-muted">{user?.email}</p>
              </div>
            </div>
            <nav className="mt-4 space-y-1">
              {LINKS.map((l) => (
                <NavLink
                  key={l.to}
                  to={l.to}
                  end={l.end}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                      isActive
                        ? 'bg-brand-50 text-brand-600 dark:bg-brand-500/15'
                        : 'text-body hover:bg-gray-50 dark:hover:bg-gray-800'
                    }`
                  }
                >
                  <IconBase name={l.icon} size={18} />
                  {t(l.key)}
                </NavLink>
              ))}
              <button
                type="button"
                onClick={handleLogout}
                className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-error-600 transition hover:bg-error-50 dark:hover:bg-error-500/10"
              >
                <IconBase name="logout" size={18} />
                {t('common.logout')}
              </button>
            </nav>
          </div>
        </aside>

        <div>
          <Outlet />
        </div>
      </div>
    </div>
  )
}
