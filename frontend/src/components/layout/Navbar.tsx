import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { categories } from '@/data/categories'
import { useCartStore } from '@/stores/cart'
import { useWishlistStore } from '@/stores/wishlist'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import IconBase from '@/components/ui/IconBase'
import LanguageSwitcher from '@/components/ui/LanguageSwitcher'

export default function Navbar() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [mobileOpen, setMobileOpen] = useState(false)
  const [catOpen, setCatOpen] = useState(false)
  const catRef = useRef<HTMLDivElement | null>(null)

  const cartCount = useCartStore((s) => s.count())
  const wishCount = useWishlistStore((s) => s.ids.length)
  const openCart = useUiStore((s) => s.openCartDrawer)
  const darkMode = useUiStore((s) => s.darkMode)
  const toggleDark = useUiStore((s) => s.toggleDarkMode)
  const isAuth = useAuthStore((s) => s.user !== null)

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (catRef.current && !catRef.current.contains(e.target as Node)) setCatOpen(false)
    }
    document.addEventListener('click', onClick)
    return () => document.removeEventListener('click', onClick)
  }, [])

  function submitSearch(e: FormEvent) {
    e.preventDefault()
    const q = query.trim()
    if (q) navigate(`/busca?q=${encodeURIComponent(q)}`)
    setMobileOpen(false)
  }

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `text-sm font-medium transition ${isActive ? 'text-brand-600' : 'text-body hover:text-heading'}`

  return (
    <header className="surface-header sticky top-0 z-40 border-b border-themed">
      {/* top utility bar */}
      <div className="border-b border-themed bg-brand-500 text-white">
        <div className="mx-auto flex h-9 max-w-7xl items-center justify-center px-4 text-xs font-medium sm:px-6 lg:px-8">
          <IconBase name="truck" size={14} className="mr-1.5" />
          {t('home.benefits.shippingDesc')}
        </div>
      </div>

      <div className="mx-auto flex h-16 max-w-7xl items-center gap-3 px-4 sm:px-6 lg:gap-6 lg:px-8">
        <button
          type="button"
          onClick={() => setMobileOpen((s) => !s)}
          className="flex h-10 w-10 items-center justify-center rounded-lg text-body lg:hidden"
          aria-label="Menu"
        >
          <IconBase name={mobileOpen ? 'x' : 'menu'} size={22} />
        </button>

        <Link to="/store" className="flex items-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-brand-500 text-lg font-bold text-white shadow-theme-md">
            L
          </span>
          <span className="hidden text-lg font-bold text-heading sm:block">{t('brand')}</span>
        </Link>

        {/* search */}
        <form onSubmit={submitSearch} className="hidden flex-1 lg:block">
          <div className="relative mx-auto max-w-xl">
            <span className="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400">
              <IconBase name="search" size={18} />
            </span>
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('common.searchPlaceholder')}
              className="h-11 w-full rounded-full border border-themed bg-gray-50 pl-11 pr-4 text-sm text-body placeholder:text-muted outline-none focus:border-brand-300 focus-ring dark:bg-gray-900"
            />
          </div>
        </form>

        <div className="ml-auto flex items-center gap-1 sm:gap-2">
          <button
            type="button"
            onClick={toggleDark}
            aria-label={t('common.theme')}
            className="flex h-10 w-10 items-center justify-center rounded-full border border-themed text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
          >
            <IconBase name={darkMode ? 'sun' : 'moon'} size={18} />
          </button>

          <LanguageSwitcher />

          <Link
            to="/conta/favoritos"
            aria-label={t('common.wishlist')}
            className="relative hidden h-10 w-10 items-center justify-center rounded-full border border-themed text-body transition hover:bg-gray-50 dark:hover:bg-gray-800 sm:flex"
          >
            <IconBase name="heart" size={18} />
            {wishCount > 0 && (
              <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-error-500 px-1 text-[10px] font-bold text-white">
                {wishCount}
              </span>
            )}
          </Link>

          <Link
            to={isAuth ? '/conta' : '/login'}
            aria-label={t('common.account')}
            className="flex h-10 w-10 items-center justify-center rounded-full border border-themed text-body transition hover:bg-gray-50 dark:hover:bg-gray-800"
          >
            <IconBase name="user" size={18} />
          </Link>

          <button
            type="button"
            onClick={openCart}
            aria-label={t('common.cart')}
            className="relative flex h-10 w-10 items-center justify-center rounded-full bg-brand-500 text-white transition hover:bg-brand-600"
          >
            <IconBase name="cart" size={18} />
            {cartCount > 0 && (
              <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-error-500 px-1 text-[10px] font-bold text-white">
                {cartCount}
              </span>
            )}
          </button>
        </div>
      </div>

      {/* category nav (desktop) */}
      <div className="border-t border-themed">
        <div className="mx-auto flex h-12 max-w-7xl items-center gap-1 px-4 sm:px-6 lg:px-8">
          <div className="relative" ref={catRef}>
            <button
              type="button"
              onClick={() => setCatOpen((s) => !s)}
              className="flex h-9 items-center gap-2 rounded-lg bg-brand-50 px-3 text-sm font-medium text-brand-700 transition hover:bg-brand-100 dark:bg-brand-500/15 dark:text-brand-400"
            >
              <IconBase name="menu" size={16} />
              {t('nav.allCategories')}
              <IconBase name="chevron-down" size={14} />
            </button>
            {catOpen && (
              <div className="surface-card absolute left-0 z-30 mt-2 w-64 rounded-xl border p-2 shadow-theme-lg">
                {categories.map((c) => (
                  <Link
                    key={c.slug}
                    to={`/categoria/${c.slug}`}
                    onClick={() => setCatOpen(false)}
                    className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-body transition hover:bg-gray-50 hover:text-brand-600 dark:hover:bg-gray-800"
                  >
                    <IconBase name={c.icon} size={18} className="text-brand-500" />
                    {t(`categories.${c.slug}.name`)}
                  </Link>
                ))}
              </div>
            )}
          </div>

          <nav className="hidden items-center gap-1 lg:flex">
            <NavLink to="/produtos" className={navLinkClass} end>
              {({ isActive }) => (
                <span className={`rounded-lg px-3 py-2 ${isActive ? 'text-brand-600' : 'text-body hover:text-heading'}`}>
                  {t('common.products')}
                </span>
              )}
            </NavLink>
            {categories.slice(0, 4).map((c) => (
              <Link
                key={c.slug}
                to={`/categoria/${c.slug}`}
                className="rounded-lg px-3 py-2 text-sm font-medium text-body transition hover:text-heading"
              >
                {t(`categories.${c.slug}.name`)}
              </Link>
            ))}
            <Link to="/sobre" className="rounded-lg px-3 py-2 text-sm font-medium text-body transition hover:text-heading">
              {t('common.about')}
            </Link>
            <Link to="/contato" className="rounded-lg px-3 py-2 text-sm font-medium text-body transition hover:text-heading">
              {t('common.contact')}
            </Link>
          </nav>
        </div>
      </div>

      {/* mobile menu */}
      {mobileOpen && (
        <div className="surface-card border-t border-themed lg:hidden">
          <div className="space-y-4 px-4 py-4 sm:px-6">
            <form onSubmit={submitSearch}>
              <div className="relative">
                <span className="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400">
                  <IconBase name="search" size={18} />
                </span>
                <input
                  type="search"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder={t('common.searchPlaceholder')}
                  className="h-11 w-full rounded-full border border-themed bg-gray-50 pl-11 pr-4 text-sm text-body outline-none dark:bg-gray-900"
                />
              </div>
            </form>
            <nav className="flex flex-col gap-1">
              <Link to="/produtos" onClick={() => setMobileOpen(false)} className="rounded-lg px-3 py-2.5 text-sm font-medium text-body hover:bg-gray-50 dark:hover:bg-gray-800">
                {t('common.products')}
              </Link>
              {categories.map((c) => (
                <Link
                  key={c.slug}
                  to={`/categoria/${c.slug}`}
                  onClick={() => setMobileOpen(false)}
                  className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-body hover:bg-gray-50 dark:hover:bg-gray-800"
                >
                  <IconBase name={c.icon} size={18} className="text-brand-500" />
                  {t(`categories.${c.slug}.name`)}
                </Link>
              ))}
              <Link to="/sobre" onClick={() => setMobileOpen(false)} className="rounded-lg px-3 py-2.5 text-sm font-medium text-body hover:bg-gray-50 dark:hover:bg-gray-800">
                {t('common.about')}
              </Link>
              <Link to="/contato" onClick={() => setMobileOpen(false)} className="rounded-lg px-3 py-2.5 text-sm font-medium text-body hover:bg-gray-50 dark:hover:bg-gray-800">
                {t('common.contact')}
              </Link>
            </nav>
          </div>
        </div>
      )}
    </header>
  )
}
