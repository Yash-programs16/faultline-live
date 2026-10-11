import { useCallback, useEffect, useRef, useState } from 'react'
import { api, ApiError, normalizeApiResponse } from '@/api/client'
import type { ReadyResponse } from '@/api/ops'

export interface Resource<T> {
 data: T | null; loading: boolean; refreshing: boolean; error: string | null; lastSuccess: string | null; retry: () => Promise<void>
}
// Each source owns its failure, refresh, and cancellation state. Timers schedule
// only after the previous request settles; manual refresh cannot overlap it.
export function useOperations<T>(path: string, period = 10000): Resource<T> {
 const [state, setState] = useState<Omit<Resource<T>, 'retry'>>({ data: null, loading: true, refreshing: false, error: null, lastSuccess: null })
 const active = useRef<AbortController | null>(null)
 const mounted = useRef(false)
 const retry = useCallback(async () => {
  if (active.current) return
  const controller = new AbortController(); active.current = controller
  setState(s => ({ ...s, refreshing: true }))
  try {
   let data: T
   try { data = await api.get<T>(path, { signal: controller.signal }) }
   catch (error) {
    if (path === '/api/ready' && error instanceof ApiError && error.status === 503 && error.data && typeof error.data === 'object' && 'checks' in error.data) data = normalizeApiResponse(error.data) as T
    else throw error
   }
   if (mounted.current && !controller.signal.aborted) setState({ data, loading: false, refreshing: false, error: null, lastSuccess: new Date().toISOString() })
  } catch (error) {
   if (mounted.current && !controller.signal.aborted) setState(s => ({ ...s, loading: false, refreshing: false, error: `${path}: ${error instanceof Error ? error.message : String(error)}` }))
  } finally { if (active.current === controller) active.current = null }
 }, [path])
 useEffect(() => {
  mounted.current = true
  let timer: ReturnType<typeof setTimeout>
  let cancelled = false
  const poll = async () => { await retry(); if (!cancelled && period > 0) timer = setTimeout(poll, period) }
  void poll()
  return () => { cancelled = true; mounted.current = false; clearTimeout(timer); active.current?.abort(); active.current = null }
 }, [retry, period])
 return { ...state, retry }
}
export function dependencyStatus(ready: ReadyResponse | null, name: string) {
 const check = ready?.checks[name]
 return check ? check.status === 'ok' ? 'healthy' : 'down' : 'unknown'
}
