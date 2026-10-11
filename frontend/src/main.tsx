import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'

import './i18n'
import './assets/styles/app.css'
import { registerSW } from 'virtual:pwa-register'
import { useAuthStore } from './stores/auth'

registerSW({ immediate: true })

// Hydrate auth from token on app start
useAuthStore.getState().hydrate()

createRoot(document.getElementById('app')!).render(
  <StrictMode>
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <App />
    </BrowserRouter>
  </StrictMode>
)