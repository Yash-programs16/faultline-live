import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  message: string
  type: ToastType
  duration: number
}

let toastId = 0

interface UiState {
  darkMode: boolean
  cartDrawerOpen: boolean
  toasts: Toast[]

  toggleDarkMode: () => void
  setDarkMode: (val: boolean) => void
  applyTheme: () => void
  openCartDrawer: () => void
  closeCartDrawer: () => void
  notify: (message: string, type?: ToastType, duration?: number) => void
  notifySuccess: (message: string) => void
  notifyError: (message: string) => void
  notifyWarning: (message: string) => void
  dismissToast: (id: number) => void
}

export const useUiStore = create<UiState>()(
  persist(
    (set, get) => ({
      darkMode: false,
      cartDrawerOpen: false,
      toasts: [],

      toggleDarkMode() {
        set({ darkMode: !get().darkMode })
        get().applyTheme()
      },
      setDarkMode(val) {
        set({ darkMode: val })
        get().applyTheme()
      },
      applyTheme() {
        const root = document.documentElement
        if (get().darkMode) root.classList.add('dark')
        else root.classList.remove('dark')
      },
      openCartDrawer() {
        set({ cartDrawerOpen: true })
      },
      closeCartDrawer() {
        set({ cartDrawerOpen: false })
      },
      notify(message, type = 'info', duration = 3500) {
        const id = ++toastId
        set({ toasts: [...get().toasts, { id, message, type, duration }] })
        setTimeout(() => get().dismissToast(id), duration)
      },
      notifySuccess(message) {
        get().notify(message, 'success')
      },
      notifyError(message) {
        get().notify(message, 'error')
      },
      notifyWarning(message) {
        get().notify(message, 'warning')
      },
      dismissToast(id) {
        set({ toasts: get().toasts.filter((t) => t.id !== id) })
      },
    }),
    {
      name: 'shop-ui',
      partialize: (s) => ({ darkMode: s.darkMode }),
    }
  )
)
