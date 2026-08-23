import { ArrowRight, ArrowsLeftRight, CheckCircle, DownloadSimple, Graph, ListChecks, Scales, WarningCircle, Wrench } from '../components/icons'
import { lazy, Suspense, useMemo, useState } from 'react'
import type { PlanChange, PlanResult, PlanSummary, RehearsalResult, ResourceInstance, StateComparison } from '../api/types'
import { Findings } from '../components/Findings'

const DependencyGraph = lazy(() => import('./DependencyGraph'))

function Metric({ label, value, tone = '', detail }: { label: string; value: number | string; tone?: string; detail: string }) {
  return <div className={'metric ' + tone}><span>{label}</span><strong>{value}</strong><small>{detail}</small></div>
}

function ResourceMini({ resource }: { resource: ResourceInstance }) {
  return <dl className="resourceMini"><div><dt>Provider</dt><dd><code>{resource.provider_address}</code></dd></div><div><dt>Schema</dt><dd>{resource.schema_version}</dd></div><div><dt>Dependencies</dt><dd>{resource.dependencies?.length ?? 0}</dd></div><div><dt>Structural keys</dt><dd>{resource.structure.keys?.length ?? 0}</dd></div></dl>
}

function PlanMetrics({ summary }: { summary: PlanSummary }) {
  return <div className="planMetrics"><div><span>Create</span><strong>{summary.creates}</strong></div><div><span>Update</span><strong>{summary.updates}</strong></div><div className="metricDanger"><span>Delete</span><strong>{summary.deletes}</strong></div><div className="metricDanger"><span>Replace</span><strong>{summary.replacements}</strong></div><div><span>No-op</span><strong>{summary.no_ops}</strong></div><div><span>Read</span><strong>{summary.reads}</strong></div></div>
}

function PlanChangeList({ changes, empty }: { changes: PlanChange[]; empty: string }) {
  if (!changes.length) return <div className="emptyState compactEmpty"><CheckCircle size={24} weight="duotone" /><strong>{empty}</strong></div>
  return <div className="planChangeList">{changes.map((change) => <details key={change.address}><summary><code>{change.address}</code><span>{change.actions.join(' + ')}</span></summary><div><p><strong>Provider</strong><code>{change.provider}</code></p><p><strong>Changed paths</strong><span>{change.changed_paths.join(', ') || 'None reported'}</span></p><p><strong>Unknown paths</strong><span>{change.unknown_paths.join(', ') || 'None reported'}</span></p></div></details>)}</div>
}

type ComparisonTab = 'overview' | 'changes' | 'moves' | 'graph' | 'findings'

export function ComparisonView({ result, onExport }: { result: StateComparison; onExport: () => void }) {
  const [activeTab, setActiveTab] = useState<ComparisonTab>('overview')
  const graphData = useMemo(() => {
    const resources = [...result.added, ...result.removed, ...result.changed.map((item) => item.after), ...result.probable_moves.flatMap((move) => [move.from, move.to])]
    const unique = [...new Map(resources.map((resource) => [resource.address, resource])).values()]
    const tones: Record<string, 'added' | 'removed' | 'changed' | 'moved'> = {}
    result.added.forEach((resource) => { tones[resource.address] = 'added' })
    result.removed.forEach((resource) => { tones[resource.address] = 'removed' })
    result.changed.forEach((change) => { tones[change.address] = 'changed' })
    result.probable_moves.forEach((move) => { tones[move.from_address] = 'moved'; tones[move.to_address] = 'moved' })
    const moduleNames = [...new Set(unique.map((resource) => resource.module_address || 'root'))]
    return { resources: unique, tones, modules: moduleNames.map((address) => ({ address: address === 'root' ? '' : address, resource_count: unique.filter((resource) => (resource.module_address || 'root') === address).length })) }
  }, [result])
  const tabs: Array<{ id: ComparisonTab; label: string; count?: number }> = [
    { id: 'overview', label: 'Overview' }, { id: 'changes', label: 'Changes', count: result.summary.added + result.summary.removed + result.summary.changed + result.summary.provider_changes }, { id: 'moves', label: 'Moves', count: result.summary.probable_moves }, { id: 'graph', label: 'Impact graph' }, { id: 'findings', label: 'Findings', count: result.findings.length },
  ]

  return <div className="resultStack">
    <div className="resultTitle"><div><span className="resultKicker"><ArrowsLeftRight size={18} weight="duotone" />Deterministic comparison</span><h2>Snapshot delta</h2><p>Exact structural differences and bounded identity inference.</p></div><button type="button" className="button buttonSecondary" onClick={onExport}><DownloadSimple size={17} weight="bold" />Export JSON</button></div>
    <div className="resultTabs" role="tablist" aria-label="Comparison views">{tabs.map((tab) => <button type="button" role="tab" aria-selected={activeTab === tab.id} className={activeTab === tab.id ? 'active' : ''} onClick={() => setActiveTab(tab.id)} key={tab.id}>{tab.label}{typeof tab.count === 'number' && tab.count > 0 && <span>{tab.count}</span>}</button>)}</div>

    {activeTab === 'overview' && <div className="tabPanel">
      <section className="metrics metricSix"><Metric label="Added" value={result.summary.added} tone="positive" detail="New identities" /><Metric label="Removed" value={result.summary.removed} tone="negative" detail="Missing identities" /><Metric label="Changed" value={result.summary.changed} tone="warning" detail="Matched identities" /><Metric label="Unchanged" value={result.summary.unchanged} detail="Stable identities" /><Metric label="Moves" value={result.summary.probable_moves} tone="move" detail="Inferred identities" /><Metric label="Providers" value={result.summary.provider_changes} detail="Usage changes" /></section>
      <section className="hashCompare"><div><span>Before fingerprint</span><code>{result.before_hash}</code></div><ArrowRight size={18} /><div><span>After fingerprint</span><code>{result.after_hash}</code></div></section>
      <div className="overviewSplit">
        <section className="resultPanel"><div className="panelTitle"><ListChecks size={20} /><div><h3>Change inventory</h3><p>Verified snapshot observations</p></div></div><div className="changeInventory"><button type="button" onClick={() => setActiveTab('changes')}><span className="changeSwatch swatchAdded" /><strong>Added</strong><span>{result.summary.added}</span></button><button type="button" onClick={() => setActiveTab('changes')}><span className="changeSwatch swatchRemoved" /><strong>Removed</strong><span>{result.summary.removed}</span></button><button type="button" onClick={() => setActiveTab('changes')}><span className="changeSwatch swatchChanged" /><strong>Changed</strong><span>{result.summary.changed}</span></button><button type="button" onClick={() => setActiveTab('moves')}><span className="changeSwatch swatchMoved" /><strong>Probable moves</strong><span>{result.summary.probable_moves}</span></button></div></section>
        <section className="resultPanel"><div className="panelTitle"><Scales size={20} /><div><h3>Review posture</h3><p>What deserves attention first</p></div></div>{result.findings.length ? <button type="button" className="reviewSummary" onClick={() => setActiveTab('findings')}><WarningCircle size={27} weight="duotone" /><strong>{result.findings.length} evidence-backed {result.findings.length === 1 ? 'finding' : 'findings'}</strong><span>Open the review queue</span></button> : <div className="emptyState compactEmpty successEmpty"><CheckCircle size={26} weight="duotone" /><strong>No review items</strong></div>}</section>
      </div>
    </div>}

    {activeTab === 'changes' && <div className="tabPanel changeSections">
      <section className="resultPanel"><div className="panelTitle"><ListChecks size={20} /><div><h3>Matched resources</h3><p>Side-by-side sanitized metadata</p></div></div>{result.changed.length ? <div className="changedList">{result.changed.map((change) => <article key={change.address}><div className="changedHead"><code>{change.address}</code><span>{change.changed_fields.join(', ')}</span></div><div className="beforeAfter"><div><span>Before</span><ResourceMini resource={change.before} /></div><ArrowRight size={20} /><div><span>After</span><ResourceMini resource={change.after} /></div></div></article>)}</div> : <div className="emptyState compactEmpty"><CheckCircle size={24} /><strong>No matched resources changed</strong></div>}</section>
      <div className="overviewSplit"><section className="resultPanel"><div className="panelTitle"><span className="changeSwatch swatchAdded" /><div><h3>Added</h3><p>{result.added.length} resources</p></div></div><div className="resourceAddressList">{result.added.map((resource) => <code key={resource.address}>{resource.address}</code>)}</div></section><section className="resultPanel"><div className="panelTitle"><span className="changeSwatch swatchRemoved" /><div><h3>Removed</h3><p>{result.removed.length} resources</p></div></div><div className="resourceAddressList">{result.removed.map((resource) => <code key={resource.address}>{resource.address}</code>)}</div></section></div>
      <section className="resultPanel"><div className="panelTitle"><Wrench size={20} /><div><h3>Provider changes</h3><p>Schema version usage between snapshots</p></div></div>{result.provider_changes.length ? <div className="providerChanges">{result.provider_changes.map((change) => <div key={change.address}><code>{change.address}</code><span>{change.before_schema_versions?.join(', ') || 'absent'}</span><ArrowRight size={15} /><span>{change.after_schema_versions?.join(', ') || 'absent'}</span><strong>{change.change}</strong></div>)}</div> : <div className="emptyState compactEmpty">Provider usage is unchanged.</div>}</section>
    </div>}

    {activeTab === 'moves' && <section className="tabPanel">{result.probable_moves.length ? <div className="moveList">{result.probable_moves.map((move) => <article key={move.from_address + '-' + move.to_address}><div className="moveRoute"><code>{move.from_address}</code><span><ArrowRight size={18} /></span><code>{move.to_address}</code><strong>{Math.round(move.confidence * 100)}%</strong></div><div className="moveSignals">{move.signals.map((signal) => <span className={signal.match ? 'matched' : ''} key={signal.name}>{signal.name.replaceAll('_', ' ')} <strong>{Math.round(signal.weight * 100)}</strong></span>)}</div>{move.ambiguous && <div className="ambiguity"><WarningCircle size={17} /><span>Ambiguous identity. Alternatives: {move.alternatives?.join(', ') || 'competing match'}</span></div>}</article>)}</div> : <div className="emptyState"><CheckCircle size={28} weight="duotone" /><strong>No probable moves</strong><p>No disappeared and appeared resources crossed the confidence threshold.</p></div>}</section>}
    {activeTab === 'graph' && <section className="tabPanel resultPanel graphPanel"><div className="panelTitle"><Graph size={20} /><div><h3>Dependency impact</h3><p>Change colors and probable-move edges share one topology.</p></div></div><Suspense fallback={<div className="graphLoading">Preparing impact graph</div>}><DependencyGraph resources={graphData.resources} modules={graphData.modules} moves={result.probable_moves} tones={graphData.tones} /></Suspense></section>}
    {activeTab === 'findings' && <section className="tabPanel"><Findings findings={result.findings} /></section>}
  </div>
}

type RehearsalTab = 'overview' | 'plans' | 'schemas' | 'findings'

function EngineColumn({ label, result }: { label: string; result?: PlanResult }) {
  if (!result) return <div className="engineColumn emptyEngine"><span>{label}</span><strong>Not produced</strong></div>
  return <div className="engineColumn"><span>{label}</span><h3>{result.engine.name}</h3><p>{result.engine.version}{result.engine.platform ? ' on ' + result.engine.platform : ''}</p><PlanMetrics summary={result.summary} /><code>{result.source_hash}</code></div>
}

export function RehearsalView({ result, onExport }: { result: RehearsalResult; onExport: () => void }) {
  const [activeTab, setActiveTab] = useState<RehearsalTab>('overview')
  const isParity = result.kind === 'MIGRATE'
  const isUpgrade = result.kind === 'UPGRADE'
  const title = isParity ? 'Terraform and OpenTofu parity' : isUpgrade ? 'Upgrade rehearsal' : 'Plan rehearsal'
  const tabs: Array<{ id: RehearsalTab; label: string; count?: number }> = [{ id: 'overview', label: 'Overview' }, { id: 'plans', label: 'Plan actions', count: (result.baseline?.changes.length ?? 0) + (result.candidate?.changes.length ?? 0) }, { id: 'schemas', label: 'Provider schemas', count: result.schema_differences.length }, { id: 'findings', label: 'Findings', count: result.findings.length }]
  const equivalent = result.comparison?.equivalent

  return <div className="resultStack">
    <div className="resultTitle"><div><span className="resultKicker"><FlaskResultIcon kind={result.kind} />{result.execution_mode.toLowerCase().replaceAll('-', ' ')}</span><h2>{title}</h2><p>{isParity ? 'The same copied inputs, interpreted by both engines.' : isUpgrade ? 'Baseline and candidate behavior compared without applying changes.' : 'A refresh-disabled plan from a copied local workspace.'}</p></div><button type="button" className="button buttonSecondary" onClick={onExport}><DownloadSimple size={17} weight="bold" />Export JSON</button></div>
    <div className="resultTabs" role="tablist" aria-label="Rehearsal views">{tabs.map((tab) => <button type="button" role="tab" aria-selected={activeTab === tab.id} className={activeTab === tab.id ? 'active' : ''} onClick={() => setActiveTab(tab.id)} key={tab.id}>{tab.label}{typeof tab.count === 'number' && tab.count > 0 && <span>{tab.count}</span>}</button>)}</div>

    {activeTab === 'overview' && <div className="tabPanel">
      {result.comparison && <div className={'equivalenceBanner ' + (equivalent ? 'equivalent' : 'different')}>{equivalent ? <CheckCircle size={24} weight="fill" /> : <WarningCircle size={24} weight="fill" />}<div><strong>{equivalent ? 'Plan actions are equivalent' : 'Plan actions differ'}</strong><span>{equivalent ? 'Both runs produced the same normalized action summary.' : 'Review candidate-only, baseline-only, or changed actions before proceeding.'}</span></div></div>}
      <section className={'engineComparison ' + (!result.candidate ? 'singleEngine' : '')}><EngineColumn label={result.candidate ? 'Baseline' : 'Rehearsed engine'} result={result.baseline} />{result.candidate && <><span className="comparisonBridge"><ArrowsLeftRight size={22} weight="duotone" /></span><EngineColumn label="Candidate" result={result.candidate} /></>}</section>
      <section className="resultPanel"><div className="panelTitle"><ListChecks size={20} /><div><h3>Compatibility checks</h3><p>Verified, inferred, and bounded assertions</p></div></div><div className="checkList">{result.checks.map((check) => <div key={check.name}><span className={'verification verification-' + check.verification.toLowerCase()}>{check.verification}</span><div><strong>{check.name}</strong><p>{check.message}</p></div></div>)}</div></section>
      {!!result.limitations.length && <section className="limitationsPanel"><WarningCircle size={20} /><div><h3>Known boundaries</h3><ul>{result.limitations.map((item) => <li key={item}>{item}</li>)}</ul></div></section>}
    </div>}

    {activeTab === 'plans' && <div className="tabPanel planColumns"><section className="resultPanel"><div className="panelTitle"><span className="columnLabel">Baseline</span><div><h3>{result.baseline?.engine.name ?? 'Rehearsed plan'}</h3><p>{result.baseline?.changes.length ?? 0} normalized actions</p></div></div><PlanChangeList changes={result.baseline?.changes ?? []} empty="No baseline changes" /></section>{result.candidate && <section className="resultPanel"><div className="panelTitle"><span className="columnLabel">Candidate</span><div><h3>{result.candidate.engine.name}</h3><p>{result.candidate.changes.length} normalized actions</p></div></div><PlanChangeList changes={result.candidate.changes} empty="No candidate changes" /></section>}</div>}
    {activeTab === 'schemas' && <section className="tabPanel resultPanel"><div className="panelTitle"><Wrench size={20} /><div><h3>Provider schema differences</h3><p>Value-free fingerprints and changed type names</p></div></div>{result.schema_differences.length ? <div className="schemaList">{result.schema_differences.map((difference) => <details key={difference.address}><summary><code>{difference.address}</code><span>{difference.change}</span></summary><div><p><strong>Resource types</strong><span>{difference.changed_resource_types.join(', ') || 'None'}</span></p><p><strong>Data sources</strong><span>{difference.changed_data_source_types.join(', ') || 'None'}</span></p></div></details>)}</div> : <div className="emptyState"><CheckCircle size={27} weight="duotone" /><strong>No provider schema differences</strong><p>The compared fingerprints are equivalent for this rehearsal.</p></div>}</section>}
    {activeTab === 'findings' && <section className="tabPanel"><Findings findings={result.findings} /></section>}
  </div>
}

function FlaskResultIcon({ kind }: { kind: RehearsalResult['kind'] }) {
  return kind === 'MIGRATE' ? <Scales size={18} weight="duotone" /> : kind === 'UPGRADE' ? <Wrench size={18} weight="duotone" /> : <ListChecks size={18} weight="duotone" />
}
