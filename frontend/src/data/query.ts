import type { Product, SortKey } from './types'

export interface ProductFilters {
  query?: string
  category?: string
  brands?: string[]
  minPrice?: number
  maxPrice?: number
  minRating?: number
}

export function searchProducts(list: Product[], query: string): Product[] {
  const q = query.trim().toLowerCase()
  if (!q) return list
  return list.filter(
    (p) =>
      p.name.toLowerCase().includes(q) ||
      p.description.toLowerCase().includes(q) ||
      p.brand.toLowerCase().includes(q) ||
      p.category.toLowerCase().includes(q)
  )
}

export function filterProducts(list: Product[], f: ProductFilters): Product[] {
  let out = list
  if (f.query) out = searchProducts(out, f.query)
  if (f.category) out = out.filter((p) => p.category === f.category)
  if (f.brands && f.brands.length) out = out.filter((p) => f.brands!.includes(p.brand))
  if (f.minPrice != null) out = out.filter((p) => p.price >= f.minPrice!)
  if (f.maxPrice != null) out = out.filter((p) => p.price <= f.maxPrice!)
  if (f.minRating != null) out = out.filter((p) => p.rating >= f.minRating!)
  return out
}

export function sortProducts(list: Product[], sort: SortKey): Product[] {
  const out = [...list]
  switch (sort) {
    case 'price-asc':
      return out.sort((a, b) => a.price - b.price)
    case 'price-desc':
      return out.sort((a, b) => b.price - a.price)
    case 'newest':
      return out.sort((a, b) => +new Date(b.createdAt) - +new Date(a.createdAt))
    case 'rating':
      return out.sort((a, b) => b.rating - a.rating)
    case 'relevance':
    default:
      return out.sort((a, b) => Number(b.bestseller) - Number(a.bestseller) || b.rating - a.rating)
  }
}

export function paginate<T>(list: T[], page: number, perPage: number): T[] {
  const start = (page - 1) * perPage
  return list.slice(start, start + perPage)
}
