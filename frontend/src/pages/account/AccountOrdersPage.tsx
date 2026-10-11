import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useOrdersStore, type OrderStatus } from '@/stores/orders'
import { formatPrice } from '@/lib/format'
import IconBase from '@/components/ui/IconBase'
import EmptyState from '@/components/ui/EmptyState'

const STATUS_FLOW: OrderStatus[] = ['placed', 'paid', 'shipped', 'delivered']

const STATUS_COLOR: Record<OrderStatus, string> = {
  placed: 'bg-blue-light-50 text-blue-light-700 dark:bg-blue-light-500/15 dark:text-blue-light-500',
  paid: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-400',
  shipped: 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-500',
  delivered: 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500',
}

export default function AccountOrdersPage() {
  const { t, i18n } = useTranslation()
  const orders = useOrdersStore((s) => s.orders)
  const [expanded, setExpanded] = useState<string | null>(orders[0]?.id ?? null)

  if (orders.length === 0) {
    return (
      <div className="surface-card rounded-2xl border">
        <EmptyState
          icon="package"
          title={t('account.noOrders')}
          description={t('account.noOrdersDesc')}
          action={
            <Link to="/produtos" className="rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600">
              {t('cart.continueShopping')}
            </Link>
          }
        />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {orders.map((o) => {
        const statusIndex = STATUS_FLOW.indexOf(o.status)
        const isOpen = expanded === o.id
        return (
          <div key={o.id} className="surface-card overflow-hidden rounded-2xl border">
            <button
              type="button"
              onClick={() => setExpanded(isOpen ? null : o.id)}
              className="flex w-full items-center justify-between gap-4 p-5 text-left"
            >
              <div>
                <p className="font-semibold text-heading">{o.id}</p>
                <p className="text-xs text-muted">
                  {t('account.orderDate')}: {new Date(o.createdAt).toLocaleDateString(i18n.language)} ·{' '}
                  {t('account.orderItems', { count: o.items.length })}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${STATUS_COLOR[o.status]}`}>
                  {t(`account.orderStatus.${o.status}`)}
                </span>
                <span className="font-semibold text-heading">{formatPrice(o.total, i18n.language)}</span>
                <IconBase name={isOpen ? 'chevron-down' : 'chevron-right'} size={18} className="text-muted" />
              </div>
            </button>

            {isOpen && (
              <div className="border-t border-themed p-5">
                {/* Timeline */}
                <div className="mb-6">
                  <p className="mb-3 text-sm font-medium text-heading">{t('account.timeline')}</p>
                  <div className="flex items-center">
                    {STATUS_FLOW.map((s, i) => (
                      <div key={s} className="flex flex-1 items-center last:flex-none">
                        <div className="flex flex-col items-center gap-1.5">
                          <span
                            className={`flex h-8 w-8 items-center justify-center rounded-full text-xs ${
                              i <= statusIndex ? 'bg-success-500 text-white' : 'border border-themed text-muted'
                            }`}
                          >
                            {i <= statusIndex ? <IconBase name="check" size={14} /> : i + 1}
                          </span>
                          <span className="hidden text-[10px] text-muted sm:block">{t(`account.orderStatus.${s}`)}</span>
                        </div>
                        {i < STATUS_FLOW.length - 1 && (
                          <span className={`mx-2 h-px flex-1 ${i < statusIndex ? 'bg-success-500' : 'bg-gray-200 dark:bg-gray-700'}`} />
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                {/* Items */}
                <div className="space-y-3">
                  {o.items.map((item) => (
                    <div key={item.id} className="flex items-center gap-3">
                      <img src={item.image} alt={item.name} className="h-12 w-12 rounded-lg border border-themed object-cover" />
                      <div className="flex-1 text-sm">
                        <p className="text-heading">{item.name}</p>
                        <p className="text-xs text-muted">×{item.qty}</p>
                      </div>
                      <span className="text-sm text-heading">{formatPrice(item.price * item.qty, i18n.language)}</span>
                    </div>
                  ))}
                </div>
                <p className="mt-4 text-xs text-muted">
                  <IconBase name="map-pin" size={13} className="mr-1 inline" />
                  {o.shippingAddress && (
                    <>
                      {o.shippingAddress.recipient} — {o.shippingAddress.street}, {o.shippingAddress.number}{o.shippingAddress.complement ? `, ${o.shippingAddress.complement}` : ''} — {o.shippingAddress.district}, {o.shippingAddress.city}/{o.shippingAddress.state} — {o.shippingAddress.zip}
                    </>
                  )}
                </p>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
