import { CheckCircle, CircleNotch, StopCircle, WarningCircle } from './icons'
import type { AnalysisJob } from '../hooks/useAnalysisJob'

export function JobProgress({ analysis }: { analysis: AnalysisJob }) {
  if (!analysis.job && !analysis.error) return null
  const status = analysis.job?.status ?? 'FAILED'
  const complete = status === 'COMPLETED'
  const failed = status === 'FAILED' || status === 'CANCELLED' || !!analysis.error
  return <section className={'progressPanel ' + (complete ? 'progressComplete ' : '') + (failed ? 'progressFailed' : '')} aria-live="polite">
    <div className="progressHead">
      <span className="progressIcon">{complete ? <CheckCircle size={22} weight="fill" /> : failed ? <WarningCircle size={22} weight="fill" /> : <CircleNotch className="spin" size={22} weight="bold" />}</span>
      <div><span className="jobStatus">{status}</span><strong>{analysis.events.at(-1)?.message ?? (analysis.error || 'Preparing analysis')}</strong></div>
      {analysis.busy && <button className="button buttonQuiet buttonSmall" type="button" onClick={() => void analysis.cancel()}><StopCircle size={16} />Cancel</button>}
    </div>
    {analysis.busy && <div className="progressTrack" aria-hidden="true"><span /></div>}
    {analysis.error && <p className="errorText">{analysis.error}</p>}
    {!!analysis.events.length && <ol className="eventList">{analysis.events.slice(-5).map((event) => <li key={event.id}><span>{event.id}</span>{event.message || event.type}</li>)}</ol>}
  </section>
}
