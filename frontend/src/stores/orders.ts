import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { createOrder, fetchOrders, fetchOrder } from '@/api/orders'
import type { CartItem } from './cart'
import type { Order, OrdersResponse } from '@/api/orders'

export type OrderStatus = 'placed' | 'paid' | 'shipped' | 'delivered'
export type { Order } from '@/api/orders'

interface OrdersState {
  orders: Order[]
  lastOrderId: string | null
  isLoading: boolean
  placeOrder: (order: Parameters<typeof createOrder>[0]) => Promise<string | null>
  loadOrders: (page?: number, pageSize?: number) => Promise<void>
  getOrder: (id: string) => Order | undefined
  refreshOrder: (id: string) => Promise<Order | null>
}

function genOrderId(): string {
  const n = Math.floor(100000 + Math.random() * 900000)
  return `LJ-${n}`
}

export const useOrdersStore = create<OrdersState>()(
  persist(
    (set, get) => ({
      orders: [],
      lastOrderId: null,
      isLoading: false,
      async placeOrder(orderData) {
        set({ isLoading: true })
        try {
          const order = await createOrder(orderData)
          const newOrders = [order, ...get().orders]
          set({ orders: newOrders, lastOrderId: order.id, isLoading: false })
          return order.id
        } catch {
          set({ isLoading: false })
          return null
        }
      },
      async loadOrders(page = 1, pageSize = 10) {
        set({ isLoading: true })
        try {
          const response: OrdersResponse = await fetchOrders(page, pageSize)
          if (page === 1) {
            set({ orders: response.orders, isLoading: false })
          } else {
            set({ orders: [...get().orders, ...response.orders], isLoading: false })
          }
        } catch {
          set({ isLoading: false })
        }
      },
      getOrder(id) {
        return get().orders.find((o) => o.id === id)
      },
      async refreshOrder(id) {
        try {
          const order = await fetchOrder(id)
          set({ orders: get().orders.map((o) => (o.id === id ? order : o)) })
          return order
        } catch {
          return null
        }
      },
    }),
    { name: 'shop-orders', partialize: (s) => ({ orders: s.orders, lastOrderId: s.lastOrderId }) }
  )
)