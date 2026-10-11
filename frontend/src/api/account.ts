import { api, ApiError } from './client'
import type { User, Address, AddressRequest } from './auth'
import type { Order } from './orders'

export interface AccountOverview {
  user: User
  recentOrders: Order[]
  wishlistCount: number
  addressesCount: number
}

export async function fetchAccountOverview(): Promise<AccountOverview> {
  return api.get<AccountOverview>('/api/account')
}

export async function fetchAddresses(): Promise<Address[]> {
  return api.get<Address[]>('/api/account/addresses')
}

export async function createAddress(data: AddressRequest): Promise<Address> {
  return api.post<Address>('/api/account/addresses', data)
}

export async function updateAddress(id: number, data: Partial<AddressRequest>): Promise<Address> {
  return api.patch<Address>(`/api/account/addresses/${id}`, data)
}

export async function deleteAddress(id: number): Promise<void> {
  return api.delete<void>(`/api/account/addresses/${id}`)
}

export async function fetchWishlist(): Promise<number[]> {
  return api.get<number[]>('/api/account/wishlist')
}

export async function addToWishlist(productId: number): Promise<void> {
  return api.post<void>(`/api/account/wishlist/${productId}`, {})
}

export async function removeFromWishlist(productId: number): Promise<void> {
  return api.delete<void>(`/api/account/wishlist/${productId}`)
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}