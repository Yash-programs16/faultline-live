import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { fetchWishlist, addToWishlist, removeFromWishlist } from '@/api/auth'

interface WishlistState {
  ids: number[]
  isLoading: boolean
  load: () => Promise<void>
  toggle: (id: number) => Promise<void>
  has: (id: number) => boolean
  remove: (id: number) => Promise<void>
  clear: () => void
  count: () => number
}

export const useWishlistStore = create<WishlistState>()(
  persist(
    (set, get) => ({
      ids: [],
      isLoading: false,
      async load() {
        set({ isLoading: true })
        try {
          const ids = await fetchWishlist()
          set({ ids, isLoading: false })
        } catch {
          set({ isLoading: false })
        }
      },
      async toggle(id) {
        const ids = get().ids
        const has = ids.includes(id)
        const newIds = has ? ids.filter((x) => x !== id) : [...ids, id]
        set({ ids: newIds })
        try {
          if (has) {
            await removeFromWishlist(id)
          } else {
            await addToWishlist(id)
          }
        } catch {
          set({ ids }) // revert on error
        }
      },
      has(id) {
        return get().ids.includes(id)
      },
      async remove(id) {
        const ids = get().ids.filter((x) => x !== id)
        set({ ids })
        try {
          await removeFromWishlist(id)
        } catch {
          set({ ids: get().ids }) // revert on error
        }
      },
      clear() {
        set({ ids: [] })
      },
      count() {
        return get().ids.length
      },
    }),
    { name: 'shop-wishlist', partialize: (s) => ({ ids: s.ids }) }
  )
)