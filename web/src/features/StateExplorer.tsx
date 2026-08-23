import { BracketsCurly, Database, DownloadSimple, Graph, MagnifyingGlass, Package, TreeStructure, WarningCircle } from '../components/icons'
import { lazy, Suspense, useMemo, useState } from 'react'
import type { InspectionResult, ResourceInstance } from '../api/types'
import { Findings } from '../components/Findings'

const DependencyGraph = lazy(() => import('./DependencyGraph'))

type StateTab = 'overview' | 'resources' | 'graph' | 'findings'

function Metric({ label, value, detail }: { label: string; value: number | string; detail: string }) {
  return <div className="metric"><span>{label}</span><strong>{value}</strong><small>{detail}</small></div>
}

function ResourceRow({ resource }: { resource: ResourceInstance }) {
  return <tr><td><code>{resource.address}</code><span className="mobileCellLabel">{resource.resource_type}</span></td><td>{resource.mode}</td><td><code>{resource.provider_address}</code></td><td><span className={'resourceStatus status-' + resource.status.toLowerCase()}>{resource.status}</span></td><td>{resource.structure.keys?.length ?? 0}</td><td>{resource.dependencies?.length ?? 0}</td></tr>
}

function ModuleTree({ modules }: { modules: InspectionResult['snapshot']['modules'] }) {
  if (!modules.length) return <div className="emptyState compactEmpty"><TreeStructure size={24} /><strong>Root module only</strong></div>
  return <div className="moduleTree">{modules.map((module) => {
    const label = module.address || 'root'
    const depth = label === 'root' ? 0 : label.split('.').filter((part) => part === 'module').length
    return <div style={{ '--tree-depth': depth } as React.CSSProperties} key={label}><span className="treeRail" /><Package size={17} weight={depth ? 'regular' : 'fill'} /><code>{label}</code><span>{module.resource_count}</span></div>
  })}</div>
}

const tabs: Array<{ id: StateTab; label: string }> = [
  { id: 'overview', label: 'Overview' },
  { id: 'resources', label: 'Resources' },
  { id: 'graph', label: 'Graph' },
  { id: 'findings', label: 'Findings' },
]

export function StateExplorer({ result, onExport }: { result: InspectionResult; onExport: () => void }) {
  const [activeTab, setActiveTab] = useState<StateTab>('overview')
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState('ALL')
  const [provider, setProvider] = useState('ALL')
  const snapshot = result.snapshot
  const providers = useMemo(() => snapshot.providers.map((item) => item.address), [snapshot.providers])
  const resources = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return snapshot.resources.filter((resource) => {
      if (status !== 'ALL' && resource.status !== status) return false
      if (provider !== 'ALL' && resource.provider_address !== provider) return false
      return !needle || [resource.address, resource.provider_address, resource.resource_type].join(' ').toLowerCase().includes(needle)
    })
  }, [provider, query, snapshot.resources, status])

  return <div className="resultStack">
    <div className="resultTitle"><div><span className="resultKicker"><Database size={18} weight="duotone" />Sanitized inspection</span><h2>State explorer</h2><p>{snapshot.engine.name} {snapshot.engine.version} interpreted this snapshot.</p></div><button type="button" className="button buttonSecondary" onClick={onExport}><DownloadSimple size={17} weight="bold" />Export JSON</button></div>
    <div className="resultTabs" role="tablist" aria-label="State explorer views">{tabs.map((tab) => <button type="button" role="tab" aria-selected={activeTab === tab.id} className={activeTab === tab.id ? 'active' : ''} onClick={() => setActiveTab(tab.id)} key={tab.id}>{tab.label}{tab.id === 'findings' && result.findings.length > 0 && <span>{result.findings.length}</span>}</button>)}</div>

    {activeTab === 'overview' && <div className="tabPanel">
      <section className="metrics metricFour"><Metric label="Resources" value={snapshot.resources.length} detail="Value-free instances" /><Metric label="Providers" value={snapshot.providers.length} detail="Provider addresses" /><Metric label="Modules" value={snapshot.modules.length || 1} detail="Root included" /><Metric label="Outputs" value={snapshot.outputs.length} detail="Metadata only" /></section>
      <section className="integrityStrip"><div><BracketsCurly size={19} /><span>Format</span><strong>{snapshot.format_version}</strong></div><div><span>Serial</span><strong>{snapshot.serial ?? 'Not reported'}</strong></div><div className="hashValue"><span>Source fingerprint</span><code>{snapshot.source_hash}</code></div></section>
      <div className="overviewSplit">
        <section className="resultPanel"><div className="panelTitle"><TreeStructure size={20} /><div><h3>Module tree</h3><p>Containment without values</p></div></div><ModuleTree modules={snapshot.modules} /></section>
        <section className="resultPanel"><div className="panelTitle"><Package size={20} /><div><h3>Providers</h3><p>Addresses and schema usage</p></div></div>{snapshot.providers.length ? <div className="dataRows">{snapshot.providers.map((item) => <div key={item.address}><code>{item.address}</code><span>{item.resource_count} resources</span><small>schema {item.schema_versions?.join(', ') || 'unknown'}</small></div>)}</div> : <div className="emptyState compactEmpty">No providers reported.</div>}</section>
      </div>
      {!!result.findings.length && <button type="button" className="reviewBanner" onClick={() => setActiveTab('findings')}><WarningCircle size={22} weight="duotone" /><span><strong>{result.findings.length} review {result.findings.length === 1 ? 'item' : 'items'}</strong><small>Open the evidence-backed findings queue.</small></span><span>Review findings</span></button>}
    </div>}

    {activeTab === 'resources' && <section className="tabPanel resultPanel">
      <div className="resourceToolbar"><div><h3>Resource inventory</h3><p>No attribute values are included.</p></div><div className="resourceFilters"><label className="searchField"><MagnifyingGlass size={17} /><span className="srOnly">Search resources</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search address, type, or provider" /></label><select aria-label="Filter resources by status" value={status} onChange={(event) => setStatus(event.target.value)}><option value="ALL">All statuses</option>{[...new Set(snapshot.resources.map((item) => item.status))].sort().map((item) => <option value={item} key={item}>{item}</option>)}</select><select aria-label="Filter resources by provider" value={provider} onChange={(event) => setProvider(event.target.value)}><option value="ALL">All providers</option>{providers.map((item) => <option value={item} key={item}>{item}</option>)}</select></div></div>
      <div className="tableSummary">{resources.length} of {snapshot.resources.length} resources</div>
      <div className="tableWrap"><table><thead><tr><th>Address</th><th>Mode</th><th>Provider</th><th>Status</th><th>Keys</th><th>Dependencies</th></tr></thead><tbody>{resources.map((resource) => <ResourceRow key={resource.address} resource={resource} />)}</tbody></table></div>
      {!resources.length && <div className="emptyState"><MagnifyingGlass size={26} /><strong>No resources match</strong><p>Try a broader query or clear one of the filters.</p></div>}
    </section>}

    {activeTab === 'graph' && <section className="tabPanel resultPanel graphPanel"><div className="panelTitle"><Graph size={20} /><div><h3>Dependency graph</h3><p>Modules contain resources. Arrows point from dependencies to dependents.</p></div></div><Suspense fallback={<div className="graphLoading">Preparing topology</div>}><DependencyGraph resources={snapshot.resources} modules={snapshot.modules} /></Suspense></section>}
    {activeTab === 'findings' && <section className="tabPanel"><Findings findings={result.findings} /></section>}
  </div>
}
