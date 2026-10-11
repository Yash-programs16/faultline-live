import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useOrdersStore } from '@/stores/orders'
import { formatPrice } from '@/lib/format'
import IconBase from '@/components/ui/IconBase'
import EmptyState from '@/components/ui/EmptyState'

export default function OrderConfirmedPage() {
  const { t, i18n } = useTranslation()
  const lastOrderId = useOrdersStore((s) => s.lastOrderId)
  const order = useOrdersStore((s) => (lastOrderId ? s.getOrder(lastOrderId) : undefined))

  if (!order) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
        <EmptyState
          icon="package"
          title={t('confirmation.noOrder')}
          action={
            <Link to="/produtos" className="rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600">
              {t('confirmation.keepShopping')}
            </Link>
          }
        />
      </div>
    )
  }

  const steps = ['step1', 'step2', 'step3'] as const

  return (
    <div className="mx-auto max-w-3xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="text-center">
        <span className="inline-flex h-20 w-20 items-center justify-center rounded-full bg-success-50 text-success-500 dark:bg-success-500/15">
          <IconBase name="check" size={40} strokeWidth={2.5} />
        </span>
        <h1 className="mt-6 text-2xl font-bold text-heading sm:text-3xl">{t('confirmation.title')}</h1>
        <p className="mx-auto mt-3 max-w-md text-sm text-muted">{t('confirmation.subtitle')}</p>
      </div>

      <div className="surface-card mt-8 rounded-2xl border p-6">
        <div className="flex items-center justify-between border-b border-themed pb-4">
          <span className="text-sm text-muted">{t('confirmation.orderNumber')}</span>
          <span className="text-lg font-bold text-brand-600">{order.id}</span>
        </div>

        <div className="mt-5 space-y-3">
          {order.items.map((item) => (
            <div key={item.id} className="flex items-center gap-3">
              <img src={item.image} alt={item.name} className="h-14 w-14 rounded-lg border border-themed object-cover" />
              <div className="flex-1 text-sm">
                <p className="text-heading">{item.name}</p>
                <p className="text-xs text-muted">×{item.qty}</p>
              </div>
              <span className="text-sm font-medium text-heading">{formatPrice(item.price * item.qty, i18n.language)}</span>
            </div>
          ))}
        </div>

        <div className="mt-5 flex justify-between border-t border-themed pt-4 text-base font-bold text-heading">
          <span>{t('cart.total')}</span>
          <span>{formatPrice(order.total, i18n.language)}</span>
        </div>
      </div>

      <div className="surface-card mt-6 rounded-2xl border p-6">
        <h2 className="text-base font-semibold text-heading">{t('confirmation.nextSteps')}</h2>
        <ol className="mt-4 space-y-4">
          {steps.map((s, i) => (
            <li key={s} className="flex items-start gap-3">
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-50 text-xs font-semibold text-brand-600 dark:bg-brand-500/15">
                {i + 1}
              </span>
              <span className="text-sm text-body">{t(`confirmation.${s}`)}</span>
            </li>
          ))}
        </ol>
      </div>

      <div className="mt-8 flex flex-col gap-3 sm:flex-row">
        <Link
          to="/conta/pedidos"
          className="flex h-12 flex-1 items-center justify-center rounded-lg bg-brand-500 text-sm font-medium text-white transition hover:bg-brand-600"
        >
          {t('confirmation.viewOrder')}
        </Link>
        <Link
          to="/produtos"
          className="flex h-12 flex-1 items-center justify-center rounded-lg border border-themed text-sm font-medium text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
        >
          {t('confirmation.keepShopping')}
        </Link>
      </div>
    </div>
  )
}
