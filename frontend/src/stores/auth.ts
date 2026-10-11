import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { register, login, logout, fetchMe, fetchAddresses, createAddress, updateAddress, deleteAddress, setAuthToken, getAuthToken } from '@/api/auth'
import type { User, Address, AddressRequest } from '@/api/auth'

export { type User, type Address } from '@/api/auth'

export const DEMO_CREDENTIALS = { email: 'admin@loja.com', password: 'demo123' }

interface AuthState {
  user: User | null
  isAuthenticated: () => boolean
  login: (email: string, password: string) => Promise<boolean>
  register: (name: string, email: string, password: string) => Promise<boolean>
  logout: () => Promise<void>
  hydrate: () => Promise<void>
  // Address methods
  fetchAddresses: () => Promise<void>
  addAddress: (address: AddressRequest) => Promise<void>
  updateAddress: (id: number, patch: Partial<AddressRequest>) => Promise<void>
  removeAddress: (id: number) => Promise<void>
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isAuthenticated() {
        return get().user !== null
      },
      async login(email, password) {
        try {
          const response = await login({ email, password })
          setAuthToken(response.token)
          set({ user: response.user })
          return true
        } catch {
          return false
        }
      },
      async register(name, email, password) {
        try {
          const response = await register({ name, email, password })
          setAuthToken(response.token)
          set({ user: response.user })
          return true
        } catch {
          return false
        }
      },
      async logout() {
        try {
          await logout()
        } catch {
          // ignore
        }
        setAuthToken(null)
        set({ user: null })
      },
      async hydrate() {
        const token = getAuthToken()
        if (!token) return
        try {
          const user = await fetchMe()
          set({ user })
        } catch {
          setAuthToken(null)
          set({ user: null })
        }
      },
      async fetchAddresses() {
        try {
          const addresses = await fetchAddresses()
          set({ user: { ...get().user!, addresses } })
        } catch {
          // ignore
        }
      },
      async addAddress(address) {
        try {
          const newAddress = await createAddress(address)
          set({ user: { ...get().user!, addresses: [...get().user!.addresses, newAddress] } })
        } catch {
          // ignore
        }
      },
      async updateAddress(id, patch) {
        try {
          const updated = await updateAddress(id, patch)
          set({ user: { ...get().user!, addresses: get().user!.addresses.map(a => a.id === id ? updated : a) } })
        } catch {
          // ignore
        }
      },
      async removeAddress(id) {
        try {
          await deleteAddress(id)
          set({ user: { ...get().user!, addresses: get().user!.addresses.filter(a => a.id !== id) } })
        } catch {
          // ignore
        }
      },
    }),
    { name: 'shop-auth', partialize: (s) => ({ user: s.user }) }
  )
)