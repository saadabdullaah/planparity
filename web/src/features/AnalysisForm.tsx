import { CaretDown, Flask, FolderOpen, Gauge, LockKey, Path, ShieldCheck } from '../components/icons'
import { useState, type FormEvent, type ReactNode } from 'react'
import type { View } from '../appTypes'

type AnalysisView = Exclude<View, 'home'>
type RehearsalKind = 'plan' | 'upgrade' | 'migrate'

function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return <label className="field"><span>{label}</span>{children}{hint && <small>{hint}</small>}</label>
}

function EngineSelect({ value, onChange, label = 'Engine' }: { value: string; onChange: (value: string) => void; label?: string }) {
  return <label className="field"><span>{label}</span><select value={value} onChange={(event) => onChange(event.target.value)}><option value="terraform">Terraform</option><option value="opentofu">OpenTofu</option></select></label>
}

const descriptions: Record<AnalysisView, { title: string; copy: string }> = {
  inspect: { title: 'Inspect a state snapshot', copy: 'Build a sanitized structural inventory from one local state file.' },
  compare: { title: 'Compare state snapshots', copy: 'Trace resource identity and provider changes between two points in time.' },
  rehearse: { title: 'Run a restricted rehearsal', copy: 'Generate plans inside a temporary workspace with refresh disabled.' },
}

export function AnalysisForm({ view, onSubmit, busy }: { view: AnalysisView; onSubmit: (path: string, payload: unknown) => void; busy: boolean }) {
  const [engine, setEngine] = useState('terraform')
  const [enginePath, setEnginePath] = useState('')
  const [state, setState] = useState('')
  const [before, setBefore] = useState('')
  const [after, setAfter] = useState('')
  const [config, setConfig] = useState('')
  const [kind, setKind] = useState<RehearsalKind>('plan')
  const [candidateEngine, setCandidateEngine] = useState('terraform')
  const [baselinePath, setBaselinePath] = useState('terraform')
  const [candidatePath, setCandidatePath] = useState('tofu')
  const [threshold, setThreshold] = useState(0.85)
  const [allowRegistry, setAllowRegistry] = useState(false)
  const [inheritProxy, setInheritProxy] = useState(false)
  const [advanced, setAdvanced] = useState(false)

  function submit(event: FormEvent) {
    event.preventDefault()
    if (view === 'inspect') onSubmit('/api/v1/inspect', { engine, engine_path: enginePath || undefined, state })
    if (view === 'compare') onSubmit('/api/v1/compare', { engine, engine_path: enginePath || undefined, before, after, probable_move_threshold: threshold })
    if (view === 'rehearse' && kind === 'plan') onSubmit('/api/v1/rehearsals/plan', { engine, engine_path: enginePath || undefined, config, state, allow_registry_access: allowRegistry, inherit_proxy: inheritProxy })
    if (view === 'rehearse' && kind === 'migrate') onSubmit('/api/v1/rehearsals/migrate', { terraform_path: baselinePath, opentofu_path: candidatePath, config, state, allow_registry_access: allowRegistry, inherit_proxy: inheritProxy })
    if (view === 'rehearse' && kind === 'upgrade') onSubmit('/api/v1/rehearsals/upgrade', { from_engine: engine, from_engine_path: baselinePath, to_engine: candidateEngine, to_engine_path: candidatePath, config, state, allow_registry_access: allowRegistry, inherit_proxy: inheritProxy })
  }

  const icon = view === 'inspect' ? <FolderOpen size={22} weight="duotone" /> : view === 'compare' ? <Path size={22} weight="duotone" /> : <Flask size={22} weight="duotone" />

  return <form className="analysisForm" onSubmit={submit}>
    <div className="formIntro"><span className="formIcon">{icon}</span><h1>{descriptions[view].title}</h1><p>{descriptions[view].copy}</p></div>

    {view === 'rehearse' && <fieldset className="modePicker"><legend>Rehearsal type</legend>{(['plan', 'upgrade', 'migrate'] as const).map((option) => <button className={kind === option ? 'active' : ''} type="button" onClick={() => setKind(option)} key={option}>{option === 'migrate' ? 'Parity' : option[0].toUpperCase() + option.slice(1)}</button>)}</fieldset>}

    {view !== 'rehearse' || kind === 'plan' ? <EngineSelect value={engine} onChange={setEngine} /> : null}
    {view === 'inspect' && <Field label="State file" hint="Absolute path to a .tfstate file"><input required value={state} onChange={(event) => setState(event.target.value)} placeholder="D:\\infra\\terraform.tfstate" /></Field>}
    {view === 'compare' && <><Field label="Before state"><input required value={before} onChange={(event) => setBefore(event.target.value)} placeholder="D:\\infra\\state-before.tfstate" /></Field><Field label="After state"><input required value={after} onChange={(event) => setAfter(event.target.value)} placeholder="D:\\infra\\state-after.tfstate" /></Field></>}
    {view === 'rehearse' && <><Field label="Configuration directory"><input required value={config} onChange={(event) => setConfig(event.target.value)} placeholder="D:\\infra" /></Field><Field label="State file"><input required value={state} onChange={(event) => setState(event.target.value)} placeholder="D:\\infra\\terraform.tfstate" /></Field></>}

    {view === 'rehearse' && kind === 'migrate' && <div className="fieldGrid"><Field label="Terraform binary"><input required value={baselinePath} onChange={(event) => setBaselinePath(event.target.value)} /></Field><Field label="OpenTofu binary"><input required value={candidatePath} onChange={(event) => setCandidatePath(event.target.value)} /></Field></div>}
    {view === 'rehearse' && kind === 'upgrade' && <><div className="fieldGrid"><EngineSelect label="Baseline engine" value={engine} onChange={setEngine} /><EngineSelect label="Candidate engine" value={candidateEngine} onChange={setCandidateEngine} /></div><div className="fieldGrid"><Field label="Baseline binary"><input required value={baselinePath} onChange={(event) => setBaselinePath(event.target.value)} /></Field><Field label="Candidate binary"><input required value={candidatePath} onChange={(event) => setCandidatePath(event.target.value)} /></Field></div></>}

    <button className="advancedToggle" type="button" onClick={() => setAdvanced((shown) => !shown)} aria-expanded={advanced}><Gauge size={17} />Analysis controls<CaretDown className={advanced ? 'rotated' : ''} size={15} /></button>
    {advanced && <div className="advancedPanel">
      {(view === 'inspect' || view === 'compare' || (view === 'rehearse' && kind === 'plan')) && <Field label="Engine binary" hint="Optional explicit path"><input value={enginePath} onChange={(event) => setEnginePath(event.target.value)} placeholder={engine === 'terraform' ? 'terraform' : 'tofu'} /></Field>}
      {view === 'compare' && <Field label={`Move confidence: ${Math.round(threshold * 100)}%`}><input className="rangeInput" type="range" min="0.5" max="1" step="0.01" value={threshold} onChange={(event) => setThreshold(Number(event.target.value))} /></Field>}
      {view === 'rehearse' && <div className="toggleList"><label><input type="checkbox" checked={allowRegistry} onChange={(event) => setAllowRegistry(event.target.checked)} /><span><strong>Allow registry access</strong><small>Only for provider initialization when the cache is incomplete.</small></span></label><label><input type="checkbox" checked={inheritProxy} onChange={(event) => setInheritProxy(event.target.checked)} /><span><strong>Inherit proxy settings</strong><small>Pass only approved proxy variables into the restricted runner.</small></span></label></div>}
    </div>}

    <div className="safetyNotice"><ShieldCheck size={20} weight="duotone" /><div><strong>Restricted execution</strong><span>Refresh disabled. Backend disabled. No apply path.</span></div></div>
    <button className="button buttonPrimary submitButton" type="submit" disabled={busy}>{busy ? 'Analysis running' : 'Start analysis'}<LockKey size={17} weight="bold" /></button>
  </form>
}
