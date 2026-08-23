import { ArrowLeft, Database, ShieldCheck } from './components/icons'
import { useCallback, useEffect, useState } from 'react'
import { api, bootstrapSession, downloadResult } from './api/client'
import type { AnalysisResult, EngineStatus, InspectionResult, StateComparison } from './api/types'
import type { RecentAnalysis, View } from './appTypes'
import { AppHeader } from './components/AppHeader'
import { JobProgress } from './components/JobProgress'
import { AnalysisForm } from './features/AnalysisForm'
import { ComparisonView, RehearsalView } from './features/ComparisonSummary'
import { Home } from './features/Home'
import { StateExplorer } from './features/StateExplorer'
import { useAnalysisJob } from './hooks/useAnalysisJob'

function isInspection(value: AnalysisResult): value is InspectionResult { return 'snapshot' in value }
function isComparison(value: AnalysisResult): value is StateComparison { return 'summary' in value && 'before_hash' in value }

function Results({ result, jobID }: { result: AnalysisResult; jobID: string }) {
  const exportResult = () => void downloadResult(jobID)
  if (isInspection(result)) return <StateExplorer result={result} onExport={exportResult} />
  if (isComparison(result)) return <ComparisonView result={result} onExport={exportResult} />
  return <RehearsalView result={result} onExport={exportResult} />
}

const waitingCopy: Record<Exclude<View, 'home'>, { title: string; copy: string }> = {
  inspect: { title: 'State evidence will appear here', copy: 'Start an inspection to reveal sanitized structure, topology, and findings.' },
  compare: { title: 'Comparison evidence will appear here', copy: 'Start a comparison to trace exact deltas and probable resource moves.' },
  rehearse: { title: 'Rehearsal evidence will appear here', copy: 'Start a rehearsal to compare plans, schemas, checks, and limitations.' },
}

export default function App() {
  const [view, setView] = useState<View>('home')
  const [engines, setEngines] = useState<EngineStatus[]>([])
  const [loadingEngines, setLoadingEngines] = useState(true)
  const [authenticated] = useState(() => bootstrapSession())
  const [recent, setRecent] = useState<RecentAnalysis[]>([])
  const [openedRecent, setOpenedRecent] = useState<RecentAnalysis | null>(null)
  const onCompleted = useCallback((result: AnalysisResult, id: string) => {
    const kind = isInspection(result) ? 'inspection' : isComparison(result) ? 'comparison' : result.kind.toLowerCase() + ' rehearsal'
    setRecent((items) => [{ id, kind, at: new Date().toISOString(), result }, ...items].slice(0, 5))
  }, [])
  const analysis = useAnalysisJob(onCompleted)

  useEffect(() => {
    api.engines().then(setEngines).catch(() => setEngines([])).finally(() => setLoadingEngines(false))
  }, [])

  function navigate(next: View) {
    analysis.reset()
    setOpenedRecent(null)
    setView(next)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  function openRecent(item: RecentAnalysis) {
    analysis.reset()
    setOpenedRecent(item)
    setView(isInspection(item.result) ? 'inspect' : isComparison(item.result) ? 'compare' : 'rehearse')
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const result = openedRecent?.result ?? analysis.result
  const resultJobID = openedRecent?.id ?? analysis.job?.id

  return <div className={'shell shell-' + view}>
    <AppHeader view={view} authenticated={authenticated} onNavigate={navigate} />
    {view === 'home' ? <Home engines={engines} loadingEngines={loadingEngines} recent={recent} onNavigate={navigate} onOpenRecent={openRecent} /> : <main className="analysisPage">
      <div className="analysisBreadcrumb"><button type="button" onClick={() => navigate('home')}><ArrowLeft size={16} />Workbench</button><span>/</span><strong>{view}</strong></div>
      <div className="analysisLayout">
        <aside className="formColumn"><AnalysisForm key={view} view={view} busy={analysis.busy} onSubmit={(path, payload) => { setOpenedRecent(null); void analysis.start(path, payload) }} /><div className="browserContract"><ShieldCheck size={19} weight="duotone" /><div><strong>Browser contract</strong><span>Sanitized results stay in memory. No local storage or indexed database.</span></div></div></aside>
        <div className="analysisOutput">
          <JobProgress analysis={analysis} />
          {result && resultJobID && <Results result={result} jobID={resultJobID} />}
          {!result && !analysis.job && <div className="waitingPanel"><span className="waitingIcon"><Database size={30} weight="duotone" /></span><strong>{waitingCopy[view].title}</strong><p>{waitingCopy[view].copy}</p><small>Raw state and plan values never enter the browser.</small></div>}
        </div>
      </div>
    </main>}
    <footer><span>PlanParity. Local-first compatibility evidence.</span><span>Restricted execution. API v1.</span></footer>
  </div>
}
