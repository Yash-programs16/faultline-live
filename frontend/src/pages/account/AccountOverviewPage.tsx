import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth'
import { useOrdersStore } from '@/stores/orders'
import { useWishlistStore } from '@/stores/wishlist'
import { formatPrice } from '@/lib/format'
import IconBase from '@/components/ui/IconBase'
import EmptyState from '@/components/ui/EmptyState'

export default function AccountOverviewPage() {
  const { t, i18n } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const orders = useOrdersStore((s) => s.orders)
  const wishCount = useWishlistStore((s) => s.ids.length)

  const stats = [
    { icon: 'package', label: 'account.stats.orders', value: orders.length, to: '/conta/pedidos' },
    { icon: 'heart', label: 'account.stats.wishlist', value: wishCount, to: '/conta/favoritos' },
    { icon: 'map-pin', label: 'account.stats.addresses', value: user?.addresses.length ?? 0, to: '/conta/enderecos' },
  ]

  return (
    <div className="space-y-6">
      <div className="surface-card rounded-2xl border bg-gradient-to-r from-brand-500 to-brand-600 p-6 text-white">
        <h2 className="text-xl font-bold">{t('account.hello', { name: user?.name })}</h2>
        <p className="mt-1 text-sm text-white/80">{t('account.memberSince')}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {stats.map((s) => (
          <Link
            key={s.label}
            to={s.to}
            className="surface-card flex items-center gap-4 rounded-2xl border p-5 transition hover:shadow-theme-md"
          >
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-50 text-brand-500 dark:bg-brand-500/15">
              <IconBase name={s.icon} size={24} />
            </span>
            <div>
              <p className="text-2xl font-bold text-heading">{s.value}</p>
              <p className="text-sm text-muted">{t(s.label)}</p>
            </div>
          </Link>
        ))}
      </div>

      <div className="surface-card rounded-2xl border p-6">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="font-semibold text-heading">{t('account.recentOrders')}</h3>
          <Link to="/conta/pedidos" className="text-sm font-medium text-brand-600 hover:text-brand-700">
            {t('common.viewAll')}
          </Link>
        </div>
        {orders.length === 0 ? (
          <EmptyState icon="package" title={t('account.noOrders')} description={t('account.noOrdersDesc')} />
        ) : (
          <div className="space-y-3">
            {orders.slice(0, 3).map((o) => (
              <div key={o.id} className="flex items-center justify-between rounded-xl border border-themed p-4">
                <div>
                  <p className="font-medium text-heading">{o.id}</p>
                  <p className="text-xs text-muted">{new Date(o.createdAt).toLocaleDateString(i18n.language)}</p>
                </div>
                <span className="font-semibold text-heading">{formatPrice(o.total, i18n.language)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
