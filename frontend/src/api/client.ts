const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

function joinUrl(base: string, path: string): string {
  if (!base) return path
  const normalizedBase = base.endsWith('/') ? base.slice(0, -1) : base
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = joinUrl(API_BASE, path)
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')

  const token = localStorage.getItem('auth_token')
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  if (!response.ok) {
    let errorData: { error?: string; message?: string } = {}
    try {
      errorData = await response.json()
    } catch {
      // ignore
    }
    const message = errorData.error || errorData.message || response.statusText
    throw new ApiError(response.status, message, errorData)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export const api = {
  get<T>(path: string): Promise<T> {
    return request<T>(path, { method: 'GET' })
  },
  post<T>(path: string, body: unknown): Promise<T> {
    return request<T>(path, { method: 'POST', body: JSON.stringify(body) })
  },
  patch<T>(path: string, body: unknown): Promise<T> {
    return request<T>(path, { method: 'PATCH', body: JSON.stringify(body) })
  },
  delete<T>(path: string): Promise<T> {
    return request<T>(path, { method: 'DELETE' })
  },
}