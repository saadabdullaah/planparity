import { useCallback, useEffect, useRef, useState } from 'react'
import { api, streamEvents } from '../api/client'
import type { AnalysisResult, JobEvent, JobSnapshot } from '../api/types'

export interface AnalysisJob {
  job?: JobSnapshot
  events: JobEvent[]
  result?: AnalysisResult
  error?: string
  busy: boolean
  start: (path: string, payload: unknown) => Promise<void>
  cancel: () => Promise<void>
  reset: () => void
}

export function useAnalysisJob(onCompleted?: (result: AnalysisResult, id: string) => void): AnalysisJob {
  const [job, setJob] = useState<JobSnapshot>()
  const [events, setEvents] = useState<JobEvent[]>([])
  const [result, setResult] = useState<AnalysisResult>()
  const [error, setError] = useState<string>()
  const controller = useRef<AbortController | undefined>(undefined)
  const completed = useRef(false)

  const watch = useCallback(async (id: string) => {
    controller.current?.abort()
    controller.current = new AbortController()
    let lastEvent = 0
    void streamEvents(id, 0, (event) => {
      lastEvent = Math.max(lastEvent, event.id)
      setEvents((current) => current.some((item) => item.id === event.id) ? current : [...current, event])
    }, controller.current.signal).catch(() => undefined)

    while (!controller.current.signal.aborted) {
      const snapshot = await api.job(id)
      setJob(snapshot)
      if (snapshot.status === 'COMPLETED') {
        const output = await api.result(id)
        setResult(output)
        if (!completed.current) {
          completed.current = true
          onCompleted?.(output, id)
        }
        controller.current.abort()
        return
      }
      if (snapshot.status === 'FAILED' || snapshot.status === 'CANCELLED') {
        setError(snapshot.failure?.message ?? `Job ${snapshot.status.toLowerCase()}.`)
        controller.current.abort()
        return
      }
      await new Promise((resolve) => window.setTimeout(resolve, lastEvent ? 450 : 250))
    }
  }, [onCompleted])

  const start = useCallback(async (path: string, payload: unknown) => {
    setError(undefined)
    setResult(undefined)
    setEvents([])
    completed.current = false
    try {
      const accepted = await api.start(path, payload)
      setJob({ id: accepted.job_id, kind: '', status: 'QUEUED', created_at: new Date().toISOString() })
      await watch(accepted.job_id)
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Analysis could not be started.')
    }
  }, [watch])

  const cancel = useCallback(async () => {
    if (job) await api.cancel(job.id)
  }, [job])

  const reset = useCallback(() => {
    controller.current?.abort()
    setJob(undefined)
    setEvents([])
    setResult(undefined)
    setError(undefined)
  }, [])

  useEffect(() => () => controller.current?.abort(), [])
  return { job, events, result, error, busy: job?.status === 'QUEUED' || job?.status === 'RUNNING', start, cancel, reset }
}
