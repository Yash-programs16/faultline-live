export interface Category {
  id: number
  slug: string
  name: string
  icon: string
  description: string
}

export interface ProductVariant {
  colors: string[]
  sizes: string[]
}

export interface ProductSpec {
  label: string
  value: string
}

export interface Review {
  author: string
  rating: number
  date: string
  comment: string
}

export interface Product {
  id: number
  slug: string
  name: string
  price: number
  priceOld?: number
  category: string
  brand: string
  rating: number
  reviewsCount: number
  stock: number
  description: string
  variants: ProductVariant
  specs: ProductSpec[]
  reviews: Review[]
  featured?: boolean
  bestseller?: boolean
  isNew?: boolean
  createdAt: string
}

export type SortKey = 'relevance' | 'price-asc' | 'price-desc' | 'newest' | 'rating'
