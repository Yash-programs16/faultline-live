import { api, ApiError } from './client'
import type { Product, Category } from '@/data/types'

export interface ProductsResponse {
  products: Product[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface ProductFilters {
  query?: string
  category?: string
  brands?: string[]
  minPrice?: number
  maxPrice?: number
  minRating?: number
  sort?: string
  page?: number
  pageSize?: number
}

function buildQuery(filters: ProductFilters): string {
  const params = new URLSearchParams()
  if (filters.query) params.set('q', filters.query)
  if (filters.category) params.set('category', filters.category)
  if (filters.brands?.length) params.set('brands', filters.brands.join(','))
  if (filters.minPrice != null) params.set('minPrice', String(filters.minPrice))
  if (filters.maxPrice != null) params.set('maxPrice', String(filters.maxPrice))
  if (filters.minRating != null) params.set('minRating', String(filters.minRating))
  if (filters.sort) params.set('sort', filters.sort)
  if (filters.page) params.set('page', String(filters.page))
  if (filters.pageSize) params.set('pageSize', String(filters.pageSize))
  const qs = params.toString()
  return qs ? `?${qs}` : ''
}

export async function fetchProducts(filters: ProductFilters = {}): Promise<ProductsResponse> {
  return api.get<ProductsResponse>(`/api/products${buildQuery(filters)}`)
}

export async function fetchProduct(slug: string): Promise<Product> {
  return api.get<Product>(`/api/products/${encodeURIComponent(slug)}`)
}

export async function fetchCategories(): Promise<Category[]> {
  return api.get<Category[]>('/api/categories')
}

export async function fetchBrands(): Promise<string[]> {
  return api.get<string[]>('/api/brands')
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}