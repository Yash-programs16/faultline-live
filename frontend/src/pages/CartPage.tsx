import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useCartStore, FREE_SHIPPING } from '@/stores/cart'
import { useUiStore } from '@/stores/ui'
import { formatPrice } from '@/lib/format'
import Breadcrumb from '@/components/ui/Breadcrumb'
import QuantityStepper from '@/components/ui/QuantityStepper'
import EmptyState from '@/components/ui/EmptyState'
import IconBase from '@/components/ui/IconBase'

export default function CartPage() {
  const { t, i18n } = useTranslation()
  const lang = i18n.language
  const notify = useUiStore((s) => s.notify)

  const items = useCartStore((s) => s.items)
  const setQty = useCartStore((s) => s.setQty)
  const removeItem = useCartStore((s) => s.removeItem)
  const applyCoupon = useCartStore((s) => s.applyCoupon)
  const removeCoupon = useCartStore((s) => s.removeCoupon)
  const coupon = useCartStore((s) => s.coupon)
  const subtotal = useCartStore((s) => s.subtotal())
  const discount = useCartStore((s) => s.discount())
  const shipping = useCartStore((s) => s.shipping())
  const total = useCartStore((s) => s.total())

  const [code, setCode] = useState('')

  function handleApply() {
    if (!code.trim()) return
    if (applyCoupon(code)) {
      notify(t('cart.couponApplied', { code: code.trim().toUpperCase() }), 'success')
      setCode('')
    } else {
      notify(t('cart.couponInvalid'), 'error')
    }
  }

  const remainingFree = FREE_SHIPPING - subtotal

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb items={[{ label: t('breadcrumb.home'), to: '/' }, { label: t('cart.title') }]} />
      <h1 className="mb-8 mt-4 text-2xl font-bold text-heading sm:text-3xl">{t('cart.title')}</h1>

      {items.length === 0 ? (
        <EmptyState
          icon="cart"
          title={t('cart.empty')}
          description={t('cart.emptyDesc')}
          action={
            <Link
              to="/produtos"
              className="rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white transition hover:bg-brand-600"
            >
              {t('cart.continueShopping')}
            </Link>
          }
        />
      ) : (
        <div className="grid gap-8 lg:grid-cols-[1fr_360px]">
          {/* Items */}
          <div className="space-y-4">
            {remainingFree > 0 && (
              <div className="flex items-center gap-2 rounded-xl border border-themed bg-brand-50 px-4 py-3 text-sm text-brand-700 dark:bg-brand-500/10 dark:text-brand-400">
                <IconBase name="truck" size={18} />
                {t('cart.freeShippingHint', { amount: formatPrice(remainingFree, lang) })}
              </div>
            )}
            {items.map((item) => (
              <div key={item.id} className="surface-card flex gap-4 rounded-2xl border p-4">
                <Link to={`/produto/${item.slug}`} className="shrink-0">
                  <img
                    src={item.image}
                    alt={item.name}
                    className="h-24 w-24 rounded-xl border border-themed object-cover sm:h-28 sm:w-28"
                  />
                </Link>
                <div className="flex flex-1 flex-col">
                  <div className="flex justify-between gap-3">
                    <div>
                      <Link to={`/produto/${item.slug}`} className="font-medium text-heading hover:text-brand-600">
                        {item.name}
                      </Link>
                      {(item.color || item.size) && (
                        <p className="mt-1 text-xs text-muted">
                          {[item.size, item.color].filter(Boolean).join(' · ')}
                        </p>
                      )}
                    </div>
                    <button
                      type="button"
                      onClick={() => {
                        removeItem(item.id)
                        notify(t('cart.removed'), 'info')
                      }}
                      className="h-fit text-gray-400 hover:text-error-500"
                      aria-label={t('common.remove')}
                    >
                      <IconBase name="trash" size={18} />
                    </button>
                  </div>
                  <div className="mt-auto flex items-center justify-between pt-3">
                    <QuantityStepper value={item.qty} onChange={(q) => setQty(item.id, q)} max={item.stock} />
                    <span className="font-semibold text-heading">{formatPrice(item.price * item.qty, lang)}</span>
                  </div>
                </div>
              </div>
            ))}
            <Link
              to="/produtos"
              className="inline-flex items-center gap-2 text-sm font-medium text-brand-600 hover:text-brand-700"
            >
              <IconBase name="arrow-left" size={16} />
              {t('cart.continueShopping')}
            </Link>
          </div>

          {/* Summary */}
          <div>
            <div className="surface-card sticky top-32 rounded-2xl border p-6">
              <h2 className="text-lg font-semibold text-heading">{t('cart.summary')}</h2>

              {/* Coupon */}
              <div className="mt-5">
                <label className="mb-1.5 block text-sm font-medium text-heading">{t('cart.coupon')}</label>
                {coupon ? (
                  <div className="flex items-center justify-between rounded-lg border border-success-200 bg-success-50 px-3 py-2.5 text-sm dark:border-success-500/30 dark:bg-success-500/10">
                    <span className="font-medium text-success-700 dark:text-success-500">{coupon.code}</span>
                    <button type="button" onClick={removeCoupon} className="text-success-700 hover:text-error-500">
                      <IconBase name="x" size={16} />
                    </button>
                  </div>
                ) : (
                  <div className="flex gap-2">
                    <input
                      value={code}
                      onChange={(e) => setCode(e.target.value)}
                      placeholder={t('cart.couponPlaceholder')}
                      className="h-10 flex-1 rounded-lg border border-themed bg-transparent px-3 text-sm text-body outline-none focus:border-brand-300 focus-ring"
                    />
                    <button
                      type="button"
                      onClick={handleApply}
                      className="h-10 rounded-lg border border-themed px-4 text-sm font-medium text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
                    >
                      {t('cart.couponApply')}
                    </button>
                  </div>
                )}
                <p className="mt-1.5 text-xs text-muted">{t('cart.couponHint')}</p>
              </div>

              <div className="mt-5 space-y-3 border-t border-themed pt-5 text-sm">
                <div className="flex justify-between text-body">
                  <span>{t('cart.subtotal')}</span>
                  <span>{formatPrice(subtotal, lang)}</span>
                </div>
                {discount > 0 && (
                  <div className="flex justify-between text-success-600">
                    <span>{t('cart.discount')}</span>
                    <span>-{formatPrice(discount, lang)}</span>
                  </div>
                )}
                <div className="flex justify-between text-body">
                  <span>{t('cart.shipping')}</span>
                  <span>{shipping === 0 ? t('cart.shippingFree') : formatPrice(shipping, lang)}</span>
                </div>
                <div className="flex justify-between border-t border-themed pt-3 text-base font-bold text-heading">
                  <span>{t('cart.total')}</span>
                  <span>{formatPrice(total, lang)}</span>
                </div>
              </div>

              <Link
                to="/checkout"
                className="mt-6 flex h-12 items-center justify-center gap-2 rounded-lg bg-brand-500 text-sm font-medium text-white transition hover:bg-brand-600"
              >
                {t('cart.checkout')}
                <IconBase name="arrow-right" size={18} />
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
