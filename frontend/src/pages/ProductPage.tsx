import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { findProduct, productImages, relatedProducts } from '@/data/products'
import { useCartStore } from '@/stores/cart'
import { useWishlistStore } from '@/stores/wishlist'
import { useUiStore } from '@/stores/ui'
import Breadcrumb from '@/components/ui/Breadcrumb'
import PriceTag from '@/components/ui/PriceTag'
import RatingStars from '@/components/ui/RatingStars'
import QuantityStepper from '@/components/ui/QuantityStepper'
import ProductCard from '@/components/product/ProductCard'
import SectionHeader from '@/components/ui/SectionHeader'
import IconBase from '@/components/ui/IconBase'
import NotFoundPage from './NotFoundPage'

type Tab = 'description' | 'specs' | 'reviews'

export default function ProductPage() {
  const { t } = useTranslation()
  const { slug = '' } = useParams()
  const product = findProduct(slug)

  const addItem = useCartStore((s) => s.addItem)
  const openCart = useUiStore((s) => s.openCartDrawer)
  const notify = useUiStore((s) => s.notifySuccess)
  const toggleWish = useWishlistStore((s) => s.toggle)
  const isWished = useWishlistStore((s) => (product ? s.ids.includes(product.id) : false))

  const [activeImg, setActiveImg] = useState(0)
  const [color, setColor] = useState<string | undefined>(product?.variants.colors[0])
  const [size, setSize] = useState<string | undefined>(product?.variants.sizes[0])
  const [qty, setQty] = useState(1)
  const [tab, setTab] = useState<Tab>('description')

  if (!product) return <NotFoundPage />

  const images = productImages(product.slug, 4)
  const outOfStock = product.stock === 0
  const related = relatedProducts(product)

  const stockLabel = outOfStock
    ? { text: t('product.outOfStock'), cls: 'bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-500' }
    : product.stock <= 10
      ? { text: t('product.lowStock', { count: product.stock }), cls: 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-500' }
      : { text: t('product.inStock'), cls: 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-500' }

  function handleAdd() {
    if (outOfStock || !product) return
    addItem(
      {
        productId: product.id,
        slug: product.slug,
        name: product.name,
        price: product.price,
        image: images[0],
        color,
        size,
        stock: product.stock,
      },
      qty
    )
    notify(t('product.addedToCart'))
    openCart()
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <Breadcrumb
        items={[
          { label: t('breadcrumb.home'), to: '/' },
          { label: t('common.products'), to: '/produtos' },
          { label: t(`categories.${product.category}.name`), to: `/categoria/${product.category}` },
          { label: product.name },
        ]}
      />

      <div className="mt-6 grid gap-10 lg:grid-cols-2">
        {/* Gallery */}
        <div>
          <div className="overflow-hidden rounded-2xl border border-themed bg-gray-100 dark:bg-gray-800">
            <img src={images[activeImg]} alt={product.name} className="aspect-square w-full object-cover" />
          </div>
          <div className="mt-4 grid grid-cols-4 gap-3">
            {images.map((img, i) => (
              <button
                key={img}
                type="button"
                onClick={() => setActiveImg(i)}
                className={`overflow-hidden rounded-xl border-2 transition ${
                  i === activeImg ? 'border-brand-500' : 'border-themed hover:border-brand-300'
                }`}
              >
                <img src={img} alt={`${product.name} ${i + 1}`} className="aspect-square w-full object-cover" />
              </button>
            ))}
          </div>
        </div>

        {/* Info */}
        <div>
          <p className="text-sm font-medium uppercase tracking-wide text-muted">{product.brand}</p>
          <h1 className="mt-1 text-2xl font-bold text-heading sm:text-3xl">{product.name}</h1>
          <div className="mt-3 flex items-center gap-3">
            <RatingStars rating={product.rating} showValue />
            <span className="text-sm text-muted">{t('product.reviews', { count: product.reviewsCount })}</span>
          </div>

          <div className="mt-5">
            <PriceTag price={product.price} priceOld={product.priceOld} size="lg" showBadge />
          </div>

          <span className={`mt-4 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ${stockLabel.cls}`}>
            <IconBase name={outOfStock ? 'x' : 'check'} size={14} />
            {stockLabel.text}
          </span>

          {/* Color variants */}
          {product.variants.colors.length > 0 && (
            <div className="mt-6">
              <p className="mb-2 text-sm font-medium text-heading">
                {t('product.color')}
              </p>
              <div className="flex gap-2">
                {product.variants.colors.map((c) => (
                  <button
                    key={c}
                    type="button"
                    onClick={() => setColor(c)}
                    aria-label={c}
                    className={`h-9 w-9 rounded-full border-2 transition ${
                      color === c ? 'border-brand-500 ring-2 ring-brand-200' : 'border-themed'
                    }`}
                    style={{ backgroundColor: c }}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Size variants */}
          {product.variants.sizes.length > 0 && (
            <div className="mt-6">
              <p className="mb-2 text-sm font-medium text-heading">{t('product.size')}</p>
              <div className="flex flex-wrap gap-2">
                {product.variants.sizes.map((s) => (
                  <button
                    key={s}
                    type="button"
                    onClick={() => setSize(s)}
                    className={`flex h-10 min-w-12 items-center justify-center rounded-lg border px-3 text-sm font-medium transition ${
                      size === s
                        ? 'border-brand-500 bg-brand-50 text-brand-600 dark:bg-brand-500/15'
                        : 'border-themed text-body hover:border-brand-300'
                    }`}
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Quantity + actions */}
          <div className="mt-7 flex flex-wrap items-center gap-3">
            <QuantityStepper value={qty} onChange={setQty} max={Math.max(1, product.stock)} />
            <button
              type="button"
              onClick={handleAdd}
              disabled={outOfStock}
              className="flex h-12 flex-1 items-center justify-center gap-2 rounded-lg bg-brand-500 px-6 text-sm font-medium text-white shadow-theme-xs transition hover:bg-brand-600 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <IconBase name="cart" size={18} />
              {t('product.addToCart')}
            </button>
            <button
              type="button"
              onClick={() => toggleWish(product.id)}
              aria-label={isWished ? t('product.removeFromWishlist') : t('product.addToWishlist')}
              className="flex h-12 w-12 items-center justify-center rounded-lg border border-themed text-body transition hover:text-error-500"
            >
              <IconBase name={isWished ? 'heart-filled' : 'heart'} size={20} className={isWished ? 'text-error-500' : ''} />
            </button>
          </div>

          {/* Trust row */}
          <div className="mt-7 grid grid-cols-3 gap-3 border-t border-themed pt-6 text-center">
            {[
              { icon: 'truck', key: 'home.benefits.shippingTitle' },
              { icon: 'shield-check', key: 'home.benefits.secureTitle' },
              { icon: 'refresh', key: 'home.benefits.returnTitle' },
            ].map((b) => (
              <div key={b.icon} className="flex flex-col items-center gap-1.5">
                <IconBase name={b.icon} size={22} className="text-brand-500" />
                <span className="text-xs text-muted">{t(b.key)}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="mt-14">
        <div className="flex gap-1 border-b border-themed">
          {(['description', 'specs', 'reviews'] as Tab[]).map((tb) => (
            <button
              key={tb}
              type="button"
              onClick={() => setTab(tb)}
              className={`-mb-px border-b-2 px-4 py-3 text-sm font-medium transition ${
                tab === tb
                  ? 'border-brand-500 text-brand-600'
                  : 'border-transparent text-muted hover:text-heading'
              }`}
            >
              {t(`product.tabs.${tb}`)}
            </button>
          ))}
        </div>

        <div className="py-6">
          {tab === 'description' && <p className="max-w-3xl text-sm leading-relaxed text-body">{product.description}</p>}

          {tab === 'specs' && (
            <div className="max-w-2xl overflow-hidden rounded-xl border border-themed">
              <table className="w-full text-sm">
                <tbody>
                  {product.specs.map((spec, i) => (
                    <tr key={spec.label} className={i % 2 ? 'bg-gray-50 dark:bg-gray-900/40' : ''}>
                      <td className="w-1/3 px-4 py-3 font-medium text-heading">{spec.label}</td>
                      <td className="px-4 py-3 text-body">{spec.value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {tab === 'reviews' && (
            <div className="max-w-3xl space-y-5">
              {product.reviews.map((r, i) => (
                <div key={i} className="surface-card rounded-xl border p-5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-50 font-semibold text-brand-600 dark:bg-brand-500/15">
                        {r.author.charAt(0)}
                      </span>
                      <div>
                        <p className="text-sm font-medium text-heading">{r.author}</p>
                        <p className="text-xs text-muted">{r.date}</p>
                      </div>
                    </div>
                    <RatingStars rating={r.rating} size={14} />
                  </div>
                  <p className="mt-3 text-sm text-body">{r.comment}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Related */}
      {related.length > 0 && (
        <div className="mt-14">
          <SectionHeader title={t('product.related')} />
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
            {related.map((p) => (
              <ProductCard key={p.id} product={p} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
