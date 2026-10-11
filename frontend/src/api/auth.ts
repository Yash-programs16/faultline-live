import { api, ApiError } from './client'

export interface User {
  id: number
  name: string
  email: string
  /** Address records are loaded separately from /api/account/addresses. */
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

/**
 * The Go User model doesn't include an addresses field. Keep the frontend's
 * User shape stable so account components can safely render before addresses
 * have been fetched.
 */
function withAddressDefaults(user: Omit<User, 'addresses'> & Partial<Pick<User, 'addresses'>>): User {
  return {
    ...user,
    addresses: Array.isArray(user.addresses) ? user.addresses : [],
  }
}

function normalizeAuthResponse(response: AuthResponse): AuthResponse {
  if (!response || !response.user || typeof response.token !== 'string') {
    throw new ApiError(502, 'The server returned an invalid authentication response')
  }

  return {
    ...response,
    user: withAddressDefaults(response.user),
  }
}

/** Go request structs use json:"is_default", not json:"isDefault". */
function addressPayload(data: Partial<AddressRequest>): Record<string, unknown> {
  const { isDefault, ...rest } = data
  return {
    ...rest,
    ...(isDefault !== undefined ? { is_default: isDefault } : {}),
  }
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const response = await api.post<AuthResponse>('/api/auth/register', data)
  return normalizeAuthResponse(response)
}

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const response = await api.post<AuthResponse>('/api/auth/login', data)
  return normalizeAuthResponse(response)
}

export async function logout(): Promise<void> {
  await api.post<void>('/api/auth/logout', {})
}

export async function fetchMe(): Promise<User> {
  const user = await api.get<User>('/api/auth/me')
  return withAddressDefaults(user)
}

export async function fetchAddresses(): Promise<Address[]> {
  return api.get<Address[]>('/api/account/addresses')
}

export async function createAddress(data: AddressRequest): Promise<Address> {
  return api.post<Address>('/api/account/addresses', addressPayload(data))
}

export async function updateAddress(
  id: number,
  data: Partial<AddressRequest>
): Promise<Address> {
  return api.patch<Address>(`/api/account/addresses/${id}`, addressPayload(data))
}

export async function deleteAddress(id: number): Promise<void> {
  await api.delete<void>(`/api/account/addresses/${id}`)
}

export async function fetchWishlist(): Promise<number[]> {
  return api.get<number[]>('/api/account/wishlist')
}

export async function addToWishlist(productId: number): Promise<void> {
  await api.post<void>(`/api/account/wishlist/${productId}`, {})
}

export async function removeFromWishlist(productId: number): Promise<void> {
  await api.delete<void>(`/api/account/wishlist/${productId}`)
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function setAuthToken(token: string | null): void {
  if (typeof window === 'undefined') return

  if (token) {
    window.localStorage.setItem('auth_token', token)
  } else {
    window.localStorage.removeItem('auth_token')
  }
}

export function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null
  return window.localStorage.getItem('auth_token')
}
