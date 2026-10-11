const API_BASE = (import.meta.env?.VITE_API_BASE_URL || '').trim()

function joinUrl(base: string, path: string): string {
  if (!base) return path

  const normalizedBase = base.endsWith('/') ? base.slice(0, -1) : base
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${normalizedBase}${normalizedPath}`
}

type JsonRecord = Record<string, unknown>

function isRecord(value: unknown): value is JsonRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

/** Convert API keys such as `request_count` and `TotalPages` to `requestCount` and `totalPages`. */
function normalizeKeys(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(normalizeKeys)
  }

  if (!isRecord(value)) {
    return value
  }

  const normalized: JsonRecord = {}

  for (const [key, child] of Object.entries(value)) {
    const camelKey = key
      .replace(/[_-]+([a-zA-Z0-9])/g, (_match, character: string) => character.toUpperCase())
      .replace(/^[A-Z]/, (character) => character.toLowerCase())

    normalized[camelKey] = normalizeKeys(child)
  }

  return normalized
}

/** Unwrap the standard `{ data: ... }` API envelope before normalizing response keys. */
export function normalizeApiResponse(payload: unknown): unknown {
  const body = isRecord(payload) && Object.prototype.hasOwnProperty.call(payload, 'data')
    ? payload.data
    : payload

  return normalizeKeys(body)
}

async function readError(response: Response): Promise<{ error?: string; message?: string; [key: string]: unknown }> {
  try {
    const payload: unknown = await response.json()
    if (!isRecord(payload)) return {}

    // Support both `{ error: "..." }` and `{ data: ..., error: "..." }`-style errors.
    const nested = isRecord(payload.error) ? payload.error : undefined
    return {
      ...payload,
      error: typeof payload.error === 'string'
        ? payload.error
        : nested && typeof nested.message === 'string'
          ? nested.message
          : undefined,
      message: typeof payload.message === 'string' ? payload.message : undefined,
    }
  } catch {
    return {}
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = joinUrl(API_BASE, path)
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')

  if (options.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  // This client runs in the browser; guard localStorage for tests or non-browser callers.
  const token = typeof window !== 'undefined'
    ? window.localStorage.getItem('auth_token')
    : null

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), path.endsWith('/investigate') ? 190000 : 15000)
  let response: Response
  try {
    response = await fetch(url, { ...options, headers, signal: options.signal ? AbortSignal.any([options.signal, controller.signal]) : controller.signal })
  } catch (error) {
    throw new ApiError(0, `${options.method || 'GET'} ${path}: ${controller.signal.aborted ? 'request timed out' : error instanceof Error ? error.message : 'network unavailable'}`)
  } finally {
    clearTimeout(timeout)
  }

  if (!response.ok) {
    const errorData = await readError(response)
    const message = errorData.error || errorData.message || response.statusText || 'Request failed'
    throw new ApiError(response.status, `${options.method || 'GET'} ${path} → HTTP ${response.status}: ${message}${response.status === 401 ? '. For local access, start the backend with FAULTLINE_DEMO_MODE=true and SERVER_HOST=127.0.0.1.' : ''}`, errorData)
  }

  if (response.status === 204) {
    return undefined as T
  }

  // Handle empty successful responses without trying to parse invalid JSON.
  if (response.status === 205 || response.headers.get('Content-Length') === '0') {
    return undefined as T
  }

  let payload: unknown
  try {
    payload = await response.json()
  } catch {
    throw new ApiError(response.status, 'The server returned an invalid JSON response')
  }

  return normalizeApiResponse(payload) as T
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export const api = {
  get<T>(path: string, options: RequestInit = {}): Promise<T> {
    return request<T>(path, { ...options, method: 'GET' })
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
