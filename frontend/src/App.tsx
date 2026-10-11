import { useEffect, lazy, Suspense } from 'react'
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import StoreLayout from '@/components/layout/StoreLayout'
import { useUiStore } from '@/stores/ui'
import { useAuthStore } from '@/stores/auth'

// Route-level code splitting — each page is its own chunk.
const HomePage = lazy(() => import('@/pages/HomePage'))
const CatalogPage = lazy(() => import('@/pages/CatalogPage'))
const CategoryPage = lazy(() => import('@/pages/CategoryPage'))
const SearchPage = lazy(() => import('@/pages/SearchPage'))
const ProductPage = lazy(() => import('@/pages/ProductPage'))
const CartPage = lazy(() => import('@/pages/CartPage'))
const CheckoutPage = lazy(() => import('@/pages/CheckoutPage'))
const OrderConfirmedPage = lazy(() => import('@/pages/OrderConfirmedPage'))
const AboutPage = lazy(() => import('@/pages/AboutPage'))
const ContactPage = lazy(() => import('@/pages/ContactPage'))
const LoginPage = lazy(() => import('@/pages/LoginPage'))
const RegisterPage = lazy(() => import('@/pages/RegisterPage'))
const NotFoundPage = lazy(() => import('@/pages/NotFoundPage'))

const AccountLayout = lazy(() => import('@/pages/account/AccountLayout'))
const AccountOverviewPage = lazy(() => import('@/pages/account/AccountOverviewPage'))
const AccountOrdersPage = lazy(() => import('@/pages/account/AccountOrdersPage'))
const AccountAddressesPage = lazy(() => import('@/pages/account/AccountAddressesPage'))
const AccountWishlistPage = lazy(() => import('@/pages/account/AccountWishlistPage'))

// Ops Dashboard
const OpsDashboardPage = lazy(() => import('@/pages/ops/OpsDashboardPage'))
const OpsIncidentPage = lazy(() => import('@/pages/ops/OpsIncidentPage'))

function RequireAuth({ children }: { children: React.ReactNode }) {
  const isAuth = useAuthStore((s) => s.isAuthenticated())
  const location = useLocation()
  if (!isAuth) {
    return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />
  }
  return <>{children}</>
}

function PageLoader() {
  return (
    <div className="flex h-96 w-full items-center justify-center">
      <span className="h-8 w-8 animate-spin rounded-full border-2 border-brand-500 border-t-transparent" />
    </div>
  )
}

export default function App() {
  const applyTheme = useUiStore((s) => s.applyTheme)

  useEffect(() => {
    applyTheme()
    document.title = 'Faultline Live'
  }, [applyTheme])

  return (
    <Suspense fallback={<PageLoader />}>
      <Routes>
        <Route element={<StoreLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/produtos" element={<CatalogPage />} />
          <Route path="/categoria/:slug" element={<CategoryPage />} />
          <Route path="/busca" element={<SearchPage />} />
          <Route path="/produto/:slug" element={<ProductPage />} />
          <Route path="/carrinho" element={<CartPage />} />
          <Route path="/pedido/confirmado" element={<OrderConfirmedPage />} />
          <Route path="/sobre" element={<AboutPage />} />
          <Route path="/contato" element={<ContactPage />} />

          {/* Auth-guarded */}
          <Route
            path="/checkout"
            element={
              <RequireAuth>
                <CheckoutPage />
              </RequireAuth>
            }
          />
          <Route
            path="/conta"
            element={
              <RequireAuth>
                <AccountLayout />
              </RequireAuth>
            }
          >
            <Route index element={<AccountOverviewPage />} />
            <Route path="pedidos" element={<AccountOrdersPage />} />
            <Route path="enderecos" element={<AccountAddressesPage />} />
            <Route path="favoritos" element={<AccountWishlistPage />} />
          </Route>

          {/* Ops Dashboard - Auth guarded for safety */}
          <Route
            path="/ops"
            element={
              <RequireAuth>
                <OpsDashboardPage />
              </RequireAuth>
            }
          />
          <Route
            path="/ops/incidents/:id"
            element={
              <RequireAuth>
                <OpsIncidentPage />
              </RequireAuth>
            }
          />

          <Route path="*" element={<NotFoundPage />} />
        </Route>

        {/* Standalone (no store chrome) */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/registro" element={<RegisterPage />} />
      </Routes>
    </Suspense>
  )
}