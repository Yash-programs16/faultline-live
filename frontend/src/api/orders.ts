import { api, ApiError } from './client'

export interface CartItem {
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

export interface Address {
  id: number
  label: string
  recipient: string
  street: string
  number: string
  complement?: string
  district: string
  city: string
  state: string
  zip: string
}

export interface CreateOrderRequest {
  items: CartItem[]
  addressId: number
  paymentMethod: 'card' | 'pix' | 'boleto'
  couponCode?: string
}

export interface OrderItem {
  id: number
  productId: number
  slug: string
  name: string
  price: number
  image: string
  color?: string
  size?: string
  qty: number
}

export interface Order {
  id: string
  createdAt: string
  items: OrderItem[]
  subtotal: number
  discount: number
  shipping: number
  total: number
  payment: 'card' | 'pix' | 'boleto'
  status: 'placed' | 'paid' | 'shipped' | 'delivered'
  shippingAddress: Address
}

export interface OrdersResponse {
  orders: Order[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export async function createOrder(data: CreateOrderRequest): Promise<Order> {
  return api.post<Order>('/api/orders', data)
}

export async function fetchOrders(page = 1, pageSize = 10): Promise<OrdersResponse> {
  return api.get<OrdersResponse>(`/api/orders?page=${page}&pageSize=${pageSize}`)
}

export async function fetchOrder(id: string): Promise<Order> {
  return api.get<Order>(`/api/orders/${encodeURIComponent(id)}`)
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}