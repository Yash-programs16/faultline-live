import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface CartItem {
  id: string // unique per product + variant combo
  productId: number
  slug: string
  name: string
  price: number
  image: string
  color?: string
  size?: string
  qty: number
  stock: number
}

export interface Coupon {
  code: string
  /** discount as a fraction, e.g. 0.1 = 10% */
  percent: number
}

/** Demo coupons. A converter should keep these visible in the cart help text. */
export const COUPONS: Record<string, Coupon> = {
  DEMO10: { code: 'DEMO10', percent: 0.1 },
  LOJA20: { code: 'LOJA20', percent: 0.2 },
}

const FREE_SHIPPING_THRESHOLD = 299
const FLAT_SHIPPING = 24.9

function itemKey(productId: number, color?: string, size?: string): string {
  return [productId, color ?? '', size ?? ''].join('|')
}

interface CartState {
  items: CartItem[]
  coupon: Coupon | null
  addItem: (item: Omit<CartItem, 'id' | 'qty'>, qty?: number) => void
  removeItem: (id: string) => void
  setQty: (id: string, qty: number) => void
  increment: (id: string) => void
  decrement: (id: string) => void
  clear: () => void
  applyCoupon: (code: string) => boolean
  removeCoupon: () => void
  count: () => number
  subtotal: () => number
  discount: () => number
  shipping: () => number
  total: () => number
}

export const useCartStore = create<CartState>()(
  persist(
    (set, get) => ({
      items: [],
      coupon: null,

      addItem(item, qty = 1) {
        const id = itemKey(item.productId, item.color, item.size)
        const existing = get().items.find((i) => i.id === id)
        if (existing) {
          const next = Math.min(existing.stock, existing.qty + qty)
          set({ items: get().items.map((i) => (i.id === id ? { ...i, qty: next } : i)) })
        } else {
          set({ items: [...get().items, { ...item, id, qty: Math.min(item.stock, qty) }] })
        }
      },
      removeItem(id) {
        set({ items: get().items.filter((i) => i.id !== id) })
      },
      setQty(id, qty) {
        set({
          items: get().items.map((i) =>
            i.id === id ? { ...i, qty: Math.max(1, Math.min(i.stock, qty)) } : i
          ),
        })
      },
      increment(id) {
        const item = get().items.find((i) => i.id === id)
        if (item) get().setQty(id, item.qty + 1)
      },
      decrement(id) {
        const item = get().items.find((i) => i.id === id)
        if (item) get().setQty(id, item.qty - 1)
      },
      clear() {
        set({ items: [], coupon: null })
      },
      applyCoupon(code) {
        const found = COUPONS[code.trim().toUpperCase()]
        if (!found) return false
        set({ coupon: found })
        return true
      },
      removeCoupon() {
        set({ coupon: null })
      },
      count() {
        return get().items.reduce((sum, i) => sum + i.qty, 0)
      },
      subtotal() {
        return get().items.reduce((sum, i) => sum + i.price * i.qty, 0)
      },
      discount() {
        const c = get().coupon
        return c ? get().subtotal() * c.percent : 0
      },
      shipping() {
        const sub = get().subtotal()
        if (sub === 0 || sub >= FREE_SHIPPING_THRESHOLD) return 0
        return FLAT_SHIPPING
      },
      total() {
        return Math.max(0, get().subtotal() - get().discount() + get().shipping())
      },
    }),
    {
      name: 'shop-cart',
      partialize: (s) => ({ items: s.items, coupon: s.coupon }),
    }
  )
)

export const FREE_SHIPPING = FREE_SHIPPING_THRESHOLD
