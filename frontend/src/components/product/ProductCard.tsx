import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { Product } from '@/data/types'
import { productImages } from '@/data/products'
import { discountPercent } from '@/lib/format'
import { useCartStore } from '@/stores/cart'
import { useWishlistStore } from '@/stores/wishlist'
import { useUiStore } from '@/stores/ui'
import PriceTag from '@/components/ui/PriceTag'
import RatingStars from '@/components/ui/RatingStars'
import IconBase from '@/components/ui/IconBase'

interface Props {
  product: Product
  layout?: 'grid' | 'list'
}

export default function ProductCard({ product, layout = 'grid' }: Props) {
  const { t } = useTranslation()
  const addItem = useCartStore((s) => s.addItem)
  const openCart = useUiStore((s) => s.openCartDrawer)
  const notify = useUiStore((s) => s.notifySuccess)
  const toggleWish = useWishlistStore((s) => s.toggle)
  const isWished = useWishlistStore((s) => s.ids.includes(product.id))

  const image = productImages(product.slug, 1)[0]
  const pct = discountPercent(product.price, product.priceOld)
  const outOfStock = product.stock === 0

  function quickAdd(e: React.MouseEvent) {
    e.preventDefault()
    if (outOfStock) return
    addItem(
      {
        productId: product.id,
        slug: product.slug,
        name: product.name,
        price: product.price,
        image,
        color: product.variants.colors[0],
        size: product.variants.sizes[0],
        stock: product.stock,
      },
      1
    )
    notify(t('product.addedToCart'))
    openCart()
  }

  function wish(e: React.MouseEvent) {
    e.preventDefault()
    toggleWish(product.id)
  }

  if (layout === 'list') {
    return (
      <Link
        to={`/produto/${product.slug}`}
        className="surface-card group flex gap-4 rounded-2xl border p-4 transition hover:shadow-theme-md"
      >
        <div className="relative h-32 w-32 shrink-0 overflow-hidden rounded-xl bg-gray-100 dark:bg-gray-800">
          <img src={image} alt={product.name} loading="lazy" className="h-full w-full object-cover" />
          {pct > 0 && (
            <span className="absolute left-2 top-2 rounded-full bg-error-500 px-2 py-0.5 text-xs font-semibold text-white">
              -{pct}%
            </span>
          )}
        </div>
        <div className="flex flex-1 flex-col">
          <p className="text-xs font-medium uppercase tracking-wide text-muted">{product.brand}</p>
          <h3 className="mt-0.5 font-semibold text-heading group-hover:text-brand-600">{product.name}</h3>
          <RatingStars rating={product.rating} count={product.reviewsCount} className="mt-1.5" />
          <p className="mt-2 line-clamp-2 text-sm text-muted">{product.description}</p>
          <div className="mt-auto flex items-center justify-between pt-3">
            <PriceTag price={product.price} priceOld={product.priceOld} />
            <button
              type="button"
              onClick={quickAdd}
              disabled={outOfStock}
              className="flex h-10 items-center gap-2 rounded-lg bg-brand-500 px-4 text-sm font-medium text-white transition hover:bg-brand-600 disabled:opacity-50"
            >
              <IconBase name="cart" size={16} />
              {outOfStock ? t('product.outOfStock') : t('product.addToCart')}
            </button>
          </div>
        </div>
      </Link>
    )
  }

  return (
    <Link
      to={`/produto/${product.slug}`}
      className="surface-card group flex flex-col overflow-hidden rounded-2xl border transition hover:shadow-theme-md"
    >
      <div className="relative aspect-square overflow-hidden bg-gray-100 dark:bg-gray-800">
        <img
          src={image}
          alt={product.name}
          loading="lazy"
          className="h-full w-full object-cover transition duration-500 group-hover:scale-105"
        />
        <div className="absolute left-3 top-3 flex flex-col gap-1.5">
          {pct > 0 && (
            <span className="rounded-full bg-error-500 px-2.5 py-1 text-xs font-semibold text-white">
              -{pct}%
            </span>
          )}
          {product.isNew && (
            <span className="rounded-full bg-success-500 px-2.5 py-1 text-xs font-semibold text-white">
              {t('common.new')}
            </span>
          )}
          {product.bestseller && !product.isNew && (
            <span className="rounded-full bg-brand-500 px-2.5 py-1 text-xs font-semibold text-white">
              {t('common.bestseller')}
            </span>
          )}
        </div>
        <button
          type="button"
          onClick={wish}
          aria-label={isWished ? t('product.removeFromWishlist') : t('product.addToWishlist')}
          className="absolute right-3 top-3 flex h-9 w-9 items-center justify-center rounded-full bg-white/90 text-gray-600 shadow-theme-sm backdrop-blur transition hover:text-error-500 dark:bg-gray-dark/90 dark:text-gray-300"
        >
          <IconBase name={isWished ? 'heart-filled' : 'heart'} size={18} className={isWished ? 'text-error-500' : ''} />
        </button>
        {outOfStock && (
          <div className="absolute inset-0 flex items-center justify-center bg-white/60 dark:bg-gray-900/60">
            <span className="rounded-full bg-gray-900/80 px-3 py-1 text-xs font-semibold text-white">
              {t('product.outOfStock')}
            </span>
          </div>
        )}
        {!outOfStock && (
          <button
            type="button"
            onClick={quickAdd}
            className="absolute inset-x-3 bottom-3 flex h-10 translate-y-2 items-center justify-center gap-2 rounded-lg bg-brand-500 text-sm font-medium text-white opacity-0 shadow-theme-md transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100"
          >
            <IconBase name="cart" size={16} />
            {t('product.addToCart')}
          </button>
        )}
      </div>
      <div className="flex flex-1 flex-col p-4">
        <p className="text-xs font-medium uppercase tracking-wide text-muted">{product.brand}</p>
        <h3 className="mt-1 line-clamp-2 text-sm font-semibold text-heading group-hover:text-brand-600">
          {product.name}
        </h3>
        <RatingStars rating={product.rating} count={product.reviewsCount} size={14} className="mt-2" />
        <div className="mt-3">
          <PriceTag price={product.price} priceOld={product.priceOld} size="sm" />
        </div>
      </div>
    </Link>
  )
}
