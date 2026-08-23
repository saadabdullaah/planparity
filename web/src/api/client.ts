import type { AnalysisResult, EngineStatus, JobAccepted, JobEvent, JobSnapshot } from './types'

let sessionToken = ''

export function bootstrapSession(): boolean {
  const params = new URLSearchParams(window.location.hash.replace(/^#/, ''))
  const token = params.get('token')
  if (token) sessionToken = token
  if (window.location.hash) window.history.replaceState(null, '', `${window.location.pathname}${window.location.search}`)
  return sessionToken.length > 0
}

export function hasSession(): boolean { return sessionToken.length > 0 }
export function setSessionTokenForTest(token: string): void { sessionToken = token }

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (sessionToken) headers.set('X-PlanParity-Token', sessionToken)
  if (init.body) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...init, headers, cache: 'no-store' })
  if (!response.ok) {
    const body: unknown = await response.json().catch(() => undefined)
    const message = typeof body === 'object' && body !== null && 'error' in body &&
      typeof body.error === 'object' && body.error !== null && 'message' in body.error &&
      typeof body.error.message === 'string' ? body.error.message : `Request failed (${response.status})`
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export const api = {
  engines: async (): Promise<EngineStatus[]> => (await request<{ engines: EngineStatus[] }>('/api/v1/engines')).engines,
  start: (path: string, payload: unknown): Promise<JobAccepted> => request(path, { method: 'POST', body: JSON.stringify(payload) }),
  job: (id: string): Promise<JobSnapshot> => request(`/api/v1/jobs/${encodeURIComponent(id)}`),
  result: (id: string): Promise<AnalysisResult> => request(`/api/v1/results/${encodeURIComponent(id)}`),
  cancel: (id: string): Promise<void> => request(`/api/v1/jobs/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  exportURL: (id: string): string => `/api/v1/results/${encodeURIComponent(id)}/export`,
}

export async function downloadResult(id: string): Promise<void> {
  const response = await fetch(api.exportURL(id), { headers: { 'X-PlanParity-Token': sessionToken }, cache: 'no-store' })
  if (!response.ok) throw new Error('Export failed.')
  const url = URL.createObjectURL(await response.blob())
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `planparity-${id}.json`
  anchor.click()
  URL.revokeObjectURL(url)
}

export async function streamEvents(
  id: string,
  after: number,
  onEvent: (event: JobEvent) => void,
  signal: AbortSignal,
): Promise<void> {
  const response = await fetch(`/api/v1/jobs/${encodeURIComponent(id)}/events`, {
    headers: { Accept: 'text/event-stream', 'Last-Event-ID': String(after), 'X-PlanParity-Token': sessionToken },
    cache: 'no-store',
    signal,
  })
  if (!response.ok || !response.body) throw new Error('Progress stream unavailable.')
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const blocks = buffer.split('\n\n')
    buffer = blocks.pop() ?? ''
    for (const block of blocks) {
      const data = block.split('\n').find((line) => line.startsWith('data:'))?.slice(5).trim()
      if (data) onEvent(JSON.parse(data) as JobEvent)
    }
  }
}
