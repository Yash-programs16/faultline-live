import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useWishlistStore } from '@/stores/wishlist'
import { products } from '@/data/products'
import ProductCard from '@/components/product/ProductCard'
import EmptyState from '@/components/ui/EmptyState'

export default function AccountWishlistPage() {
  const { t } = useTranslation()
  const ids = useWishlistStore((s) => s.ids)
  const items = products.filter((p) => ids.includes(p.id))

  return (
    <div className="space-y-5">
      <h2 className="text-lg font-semibold text-heading">{t('account.wishlist')}</h2>
      {items.length === 0 ? (
        <div className="surface-card rounded-2xl border">
          <EmptyState
            icon="heart"
            title={t('account.emptyWishlist')}
            description={t('account.emptyWishlistDesc')}
            action={
              <Link to="/produtos" className="rounded-lg bg-brand-500 px-6 py-3 text-sm font-medium text-white hover:bg-brand-600">
                {t('cart.continueShopping')}
              </Link>
            }
          />
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
          {items.map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
      )}
    </div>
  )
}
