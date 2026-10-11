import { api, ApiError } from './client'

export interface User {
  id: number
  name: string
  email: string
  addresses: Address[]
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
  isDefault?: boolean
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  user: User
  token: string
}

export interface AddressRequest {
  label: string
  recipient: string
  street: string
  number: string
  complement?: string
  district: string
  city: string
  state: string
  zip: string
  isDefault?: boolean
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  return api.post<AuthResponse>('/api/auth/register', data)
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  return api.post<AuthResponse>('/api/auth/login', data)
}

export async function logout(): Promise<void> {
  return api.post<void>('/api/auth/logout', {})
}

export async function fetchMe(): Promise<User> {
  return api.get<User>('/api/auth/me')
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

export function setAuthToken(token: string | null) {
  if (token) {
    localStorage.setItem('auth_token', token)
  } else {
    localStorage.removeItem('auth_token')
  }
}

export function getAuthToken(): string | null {
  return localStorage.getItem('auth_token')
}