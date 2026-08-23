import { useCallback, useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { api, bootstrapSession, downloadResult } from './api/client'
import type { AnalysisResult, EngineStatus, InspectionResult, StateComparison } from './api/types'
import { EngineCards } from './components/EngineCards'
import { JobProgress } from './components/JobProgress'
import { ComparisonView, RehearsalView } from './features/ComparisonSummary'
import { StateExplorer } from './features/StateExplorer'
import { useAnalysisJob } from './hooks/useAnalysisJob'

type View = 'home' | 'inspect' | 'compare' | 'rehearse'
interface RecentAnalysis { id: string; kind: string; at: string; result: AnalysisResult }

const principles = [
  'Original configuration and state remain untouched',
  'Raw values never cross the backend boundary',
  'Terraform and OpenTofu are evaluated as equal engines',
]

function isInspection(value: AnalysisResult): value is InspectionResult { return 'snapshot' in value }
function isComparison(value: AnalysisResult): value is StateComparison { return 'summary' in value && 'before_hash' in value }

function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return <label className="field"><span>{label}</span>{children}{hint && <small>{hint}</small>}</label>
}

function EngineSelect({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return <select value={value} onChange={(event) => onChange(event.target.value)}><option value="terraform">Terraform</option><option value="opentofu">OpenTofu</option></select>
}

function AnalyzeForm({ view, onSubmit, busy }: { view: Exclude<View, 'home'>; onSubmit: (path: string, payload: unknown) => void; busy: boolean }) {
  const [engine, setEngine] = useState('terraform')
  const [state, setState] = useState('')
  const [before, setBefore] = useState('')
  const [after, setAfter] = useState('')
  const [config, setConfig] = useState('')
  const [kind, setKind] = useState('plan')
  const [candidateEngine, setCandidateEngine] = useState('terraform')
  const [terraformPath, setTerraformPath] = useState('terraform')
  const [opentofuPath, setOpentofuPath] = useState('tofu')

  function submit(event: FormEvent) {
    event.preventDefault()
    if (view === 'inspect') onSubmit('/api/v1/inspect', { engine, state })
    if (view === 'compare') onSubmit('/api/v1/compare', { engine, before, after, probable_move_threshold: 0.85 })
    if (view === 'rehearse' && kind === 'plan') onSubmit('/api/v1/rehearsals/plan', { engine, config, state, allow_registry_access: false })
    if (view === 'rehearse' && kind === 'migrate') onSubmit('/api/v1/rehearsals/migrate', { terraform_path: terraformPath, opentofu_path: opentofuPath, config, state, allow_registry_access: false })
    if (view === 'rehearse' && kind === 'upgrade') onSubmit('/api/v1/rehearsals/upgrade', { from_engine: engine, from_engine_path: terraformPath, to_engine: candidateEngine, to_engine_path: opentofuPath, config, state, allow_registry_access: false })
  }

  return (
    <form className="analysisForm panel" onSubmit={submit}>
      <div className="formIntro"><p className="eyebrow">Local input</p><h1>{view === 'inspect' ? 'Inspect a state snapshot' : view === 'compare' ? 'Compare state snapshots' : 'Run a restricted rehearsal'}</h1><p>Enter absolute local paths. Files are read by the local backend; content is never uploaded or retained.</p></div>
      {view === 'rehearse' && <Field label="Rehearsal"><select value={kind} onChange={(event) => setKind(event.target.value)}><option value="plan">Plan</option><option value="upgrade">Upgrade</option><option value="migrate">Terraform → OpenTofu migration</option></select></Field>}
      {view !== 'rehearse' || kind === 'plan' ? <Field label="Engine"><EngineSelect value={engine} onChange={setEngine} /></Field> : null}
      {view === 'inspect' && <Field label="State file" hint="Example: D:\\infra\\terraform.tfstate"><input required value={state} onChange={(event) => setState(event.target.value)} placeholder="Absolute path to state" /></Field>}
      {view === 'compare' && <><Field label="Before state"><input required value={before} onChange={(event) => setBefore(event.target.value)} placeholder="Absolute path to baseline state" /></Field><Field label="After state"><input required value={after} onChange={(event) => setAfter(event.target.value)} placeholder="Absolute path to candidate state" /></Field></>}
      {view === 'rehearse' && <><Field label="Configuration directory"><input required value={config} onChange={(event) => setConfig(event.target.value)} placeholder="Absolute path to configuration" /></Field><Field label="State file"><input required value={state} onChange={(event) => setState(event.target.value)} placeholder="Absolute path to state" /></Field></>}
      {view === 'rehearse' && kind === 'migrate' && <div className="fieldGrid"><Field label="Terraform binary"><input required value={terraformPath} onChange={(event) => setTerraformPath(event.target.value)} /></Field><Field label="OpenTofu binary"><input required value={opentofuPath} onChange={(event) => setOpentofuPath(event.target.value)} /></Field></div>}
      {view === 'rehearse' && kind === 'upgrade' && <><div className="fieldGrid"><Field label="Baseline engine"><EngineSelect value={engine} onChange={setEngine} /></Field><Field label="Candidate engine"><EngineSelect value={candidateEngine} onChange={setCandidateEngine} /></Field></div><div className="fieldGrid"><Field label="Baseline binary"><input required value={terraformPath} onChange={(event) => setTerraformPath(event.target.value)} /></Field><Field label="Candidate binary"><input required value={opentofuPath} onChange={(event) => setOpentofuPath(event.target.value)} /></Field></div></>}
      <div className="safetyNotice"><strong>Restricted execution</strong><span>Refresh disabled · backend disabled · no apply path</span></div>
      <button className="primary wide" type="submit" disabled={busy}>{busy ? 'Analysis running…' : 'Start analysis'}</button>
    </form>
  )
}

function Results({ result, jobID }: { result: AnalysisResult; jobID: string }) {
  const exportResult = () => void downloadResult(jobID)
  if (isInspection(result)) return <StateExplorer result={result} onExport={exportResult} />
  if (isComparison(result)) return <ComparisonView result={result} onExport={exportResult} />
  return <RehearsalView result={result} onExport={exportResult} />
}

export default function App() {
  const [view, setView] = useState<View>('home')
  const [engines, setEngines] = useState<EngineStatus[]>([])
  const [loadingEngines, setLoadingEngines] = useState(true)
  const [authenticated] = useState(() => bootstrapSession())
  const [recent, setRecent] = useState<RecentAnalysis[]>([])
  const onCompleted = useCallback((result: AnalysisResult, id: string) => setRecent((items) => [{ id, kind: 'analysis', at: new Date().toISOString(), result }, ...items].slice(0, 5)), [])
  const analysis = useAnalysisJob(onCompleted)

  useEffect(() => {
    api.engines().then(setEngines).catch(() => setEngines([])).finally(() => setLoadingEngines(false))
  }, [])

  function navigate(next: View) { analysis.reset(); setView(next); window.scrollTo({ top: 0, behavior: 'smooth' }) }

  return (
    <div className="shell">
      <header className="topbar">
        <button className="brand" type="button" onClick={() => navigate('home')} aria-label="PlanParity home"><span className="brandMark" aria-hidden="true">P</span><span>PlanParity</span></button>
        <nav aria-label="Primary"><button className={view === 'inspect' ? 'active' : ''} onClick={() => navigate('inspect')}>Inspect</button><button className={view === 'compare' ? 'active' : ''} onClick={() => navigate('compare')}>Compare</button><button className={view === 'rehearse' ? 'active' : ''} onClick={() => navigate('rehearse')}>Rehearse</button></nav>
        <span className={`status ${authenticated ? '' : 'statusWarning'}`}><i /> {authenticated ? 'Local-restricted' : 'Session missing'}</span>
      </header>

      {view === 'home' ? <main>
        <section className="hero">
          <div><p className="eyebrow">Infrastructure compatibility laboratory</p><h1>See the state transition <em>before</em> it sees production.</h1><p className="lede">Inspect, compare, and rehearse Terraform and OpenTofu behavior in copied local workspaces—with evidence that separates what is verified from what remains unknown.</p><div className="actions"><button className="primary" onClick={() => navigate('inspect')}>Inspect state</button><button className="secondary" onClick={() => navigate('compare')}>Compare snapshots</button></div></div>
          <aside className="proof" aria-label="Safety principles"><p className="proofLabel">Safety contract</p><strong>No apply. No destroy. No state push.</strong><ul>{principles.map((principle) => <li key={principle}>{principle}</li>)}</ul></aside>
        </section>
        <section className="workspaceSection"><div className="sectionHead"><div><p className="eyebrow">Local toolchain</p><h2>Engine readiness</h2></div><span className="muted">Detected on this machine</span></div><EngineCards engines={engines} loading={loadingEngines} /></section>
        <section className="workflowGrid"><button onClick={() => navigate('inspect')}><span>01</span><strong>Inspect</strong><p>Inventory resources, providers, modules, and value-free structure.</p></button><button onClick={() => navigate('compare')}><span>02</span><strong>Compare</strong><p>Detect exact changes and explain probable identity moves.</p></button><button onClick={() => navigate('rehearse')}><span>03</span><strong>Rehearse</strong><p>Run refresh-disabled plans in disposable copied workspaces.</p></button></section>
        {!!recent.length && <section className="workspaceSection"><p className="eyebrow">This process only</p><h2>Recent analyses</h2>{recent.map((item) => <div className="listRow" key={item.id}><code>{item.id.slice(0, 12)}</code><span>{new Date(item.at).toLocaleTimeString()}</span></div>)}</section>}
      </main> : <main className="analysisLayout"><AnalyzeForm view={view} busy={analysis.busy} onSubmit={(path, payload) => void analysis.start(path, payload)} /><div className="analysisOutput"><JobProgress analysis={analysis} />{analysis.result && analysis.job && <Results result={analysis.result} jobID={analysis.job.id} />}{!analysis.job && <div className="waitingPanel"><span>Analysis evidence will appear here.</span><p>Only sanitized, value-free contracts are sent to this browser.</p></div>}</div></main>}
      <footer><span>PlanParity · local-first compatibility evidence</span><span>No browser persistence · API v1</span></footer>
    </div>
  )
}
