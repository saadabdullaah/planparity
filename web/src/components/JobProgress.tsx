import type { AnalysisJob } from '../hooks/useAnalysisJob'

export function JobProgress({ analysis }: { analysis: AnalysisJob }) {
  if (!analysis.job && !analysis.error) return null
  return (
    <section className="progressPanel" aria-live="polite">
      <div className="progressHead">
        <div>
          <span className={`jobStatus status-${analysis.job?.status.toLowerCase() ?? 'failed'}`}>{analysis.job?.status ?? 'FAILED'}</span>
          <strong>{analysis.events.at(-1)?.message ?? (analysis.error || 'Preparing analysis')}</strong>
        </div>
        {analysis.busy && <button className="ghost small" type="button" onClick={() => void analysis.cancel()}>Cancel</button>}
      </div>
      {analysis.busy && <div className="progressTrack"><span /></div>}
      {analysis.error && <p className="errorText">{analysis.error}</p>}
      <ol className="eventList">
        {analysis.events.slice(-6).map((event) => <li key={event.id}>{event.message || event.type}</li>)}
      </ol>
    </section>
  )
}

