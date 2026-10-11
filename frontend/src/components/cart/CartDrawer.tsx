import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useCartStore } from '@/stores/cart'
import { useUiStore } from '@/stores/ui'
import { formatPrice } from '@/lib/format'
import IconBase from '@/components/ui/IconBase'
import QuantityStepper from '@/components/ui/QuantityStepper'
import EmptyState from '@/components/ui/EmptyState'

export default function CartDrawer() {
  const { t, i18n } = useTranslation()
  const open = useUiStore((s) => s.cartDrawerOpen)
  const close = useUiStore((s) => s.closeCartDrawer)

  const items = useCartStore((s) => s.items)
  const setQty = useCartStore((s) => s.setQty)
  const removeItem = useCartStore((s) => s.removeItem)
  const subtotal = useCartStore((s) => s.subtotal())

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && close()
    if (open) document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, close])

  return (
    <div className={`fixed inset-0 z-[60] ${open ? '' : 'pointer-events-none'}`} aria-hidden={!open}>
      <div
        className={`absolute inset-0 bg-gray-900/40 transition-opacity ${open ? 'opacity-100' : 'opacity-0'}`}
        onClick={close}
      />
      <aside
        className={`surface-card absolute right-0 top-0 flex h-full w-full max-w-md flex-col border-l shadow-theme-xl transition-transform duration-300 ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
        role="dialog"
        aria-label={t('cart.drawerTitle')}
      >
        <div className="flex items-center justify-between border-b border-themed px-5 py-4">
          <h2 className="flex items-center gap-2 text-lg font-semibold text-heading">
            <IconBase name="cart" size={20} />
            {t('cart.drawerTitle')}
          </h2>
          <button
            type="button"
            onClick={close}
            className="flex h-9 w-9 items-center justify-center rounded-full text-muted hover:text-heading"
            aria-label={t('common.close')}
          >
            <IconBase name="x" size={20} />
          </button>
        </div>

        {items.length === 0 ? (
          <div className="flex flex-1 items-center justify-center">
            <EmptyState
              icon="cart"
              title={t('cart.empty')}
              description={t('cart.emptyDesc')}
              action={
                <button
                  type="button"
                  onClick={close}
                  className="rounded-lg bg-brand-500 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-brand-600"
                >
                  {t('cart.continueShopping')}
                </button>
              }
            />
          </div>
        ) : (
          <>
            <div className="flex-1 space-y-4 overflow-y-auto px-5 py-4">
              {items.map((item) => (
                <div key={item.id} className="flex gap-3">
                  <Link to={`/produto/${item.slug}`} onClick={close} className="shrink-0">
                    <img
                      src={item.image}
                      alt={item.name}
                      className="h-20 w-20 rounded-xl border border-themed object-cover"
                    />
                  </Link>
                  <div className="flex flex-1 flex-col">
                    <div className="flex justify-between gap-2">
                      <Link
                        to={`/produto/${item.slug}`}
                        onClick={close}
                        className="line-clamp-2 text-sm font-medium text-heading hover:text-brand-600"
                      >
                        {item.name}
                      </Link>
                      <button
                        type="button"
                        onClick={() => removeItem(item.id)}
                        className="text-gray-400 hover:text-error-500"
                        aria-label={t('common.remove')}
                      >
                        <IconBase name="trash" size={16} />
                      </button>
                    </div>
                    {(item.color || item.size) && (
                      <p className="mt-0.5 text-xs text-muted">
                        {[item.size, item.color].filter(Boolean).join(' · ')}
                      </p>
                    )}
                    <div className="mt-auto flex items-center justify-between pt-2">
                      <QuantityStepper
                        value={item.qty}
                        onChange={(q) => setQty(item.id, q)}
                        max={item.stock}
                        size="sm"
                      />
                      <span className="text-sm font-semibold text-heading">
                        {formatPrice(item.price * item.qty, i18n.language)}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div className="border-t border-themed px-5 py-4">
              <div className="mb-4 flex items-center justify-between">
                <span className="text-sm text-muted">{t('cart.subtotal')}</span>
                <span className="text-lg font-bold text-heading">{formatPrice(subtotal, i18n.language)}</span>
              </div>
              <div className="flex flex-col gap-2">
                <Link
                  to="/checkout"
                  onClick={close}
                  className="flex h-11 items-center justify-center rounded-lg bg-brand-500 text-sm font-medium text-white transition hover:bg-brand-600"
                >
                  {t('cart.checkout')}
                </Link>
                <Link
                  to="/carrinho"
                  onClick={close}
                  className="flex h-11 items-center justify-center rounded-lg border border-themed text-sm font-medium text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
                >
                  {t('cart.viewCart')}
                </Link>
              </div>
            </div>
          </>
        )}
      </aside>
    </div>
  )
}
