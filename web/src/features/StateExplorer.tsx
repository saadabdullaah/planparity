import { useMemo, useState } from 'react'
import type { InspectionResult, ResourceInstance } from '../api/types'
import { Findings } from '../components/Findings'

function Metric({ label, value }: { label: string; value: number | string }) {
  return <div className="metric"><span>{label}</span><strong>{value}</strong></div>
}

function ResourceRow({ resource }: { resource: ResourceInstance }) {
  return (
    <tr>
      <td><code>{resource.address}</code></td>
      <td>{resource.mode}</td>
      <td>{resource.provider_address}</td>
      <td>{resource.status}</td>
      <td>{resource.structure.keys.length}</td>
      <td>{resource.dependencies.length}</td>
    </tr>
  )
}

export function StateExplorer({ result, onExport }: { result: InspectionResult; onExport: () => void }) {
  const [query, setQuery] = useState('')
  const snapshot = result.snapshot
  const resources = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return needle ? snapshot.resources.filter((resource) => `${resource.address} ${resource.provider_address} ${resource.resource_type}`.toLowerCase().includes(needle)) : snapshot.resources
  }, [query, snapshot.resources])

  return (
    <div className="resultStack">
      <div className="resultTitle">
        <div><p className="eyebrow">Sanitized inspection</p><h2>State explorer</h2></div>
        <button type="button" className="secondary" onClick={onExport}>Export JSON</button>
      </div>
      <section className="metrics">
        <Metric label="Resources" value={snapshot.resources.length} />
        <Metric label="Providers" value={snapshot.providers.length} />
        <Metric label="Modules" value={snapshot.modules.length} />
        <Metric label="Outputs" value={snapshot.outputs.length} />
      </section>
      <section className="panel integrity">
        <div><span>Engine</span><strong>{snapshot.engine.name} {snapshot.engine.version}</strong></div>
        <div><span>Serial</span><strong>{snapshot.serial ?? 'Not reported'}</strong></div>
        <div><span>Source fingerprint</span><code>{snapshot.source_hash}</code></div>
      </section>
      <section className="panel">
        <div className="panelHead"><div><p className="eyebrow">Value-free inventory</p><h2>Resources</h2></div><input aria-label="Filter resources" placeholder="Filter address, type, provider…" value={query} onChange={(event) => setQuery(event.target.value)} /></div>
        <div className="tableWrap"><table><thead><tr><th>Address</th><th>Mode</th><th>Provider</th><th>Status</th><th>Keys</th><th>Dependencies</th></tr></thead><tbody>{resources.map((resource) => <ResourceRow key={resource.address} resource={resource} />)}</tbody></table></div>
        {!resources.length && <div className="emptyState">No resources match that filter.</div>}
      </section>
      <div className="splitPanels">
        <section className="panel"><p className="eyebrow">Provider inventory</p><h2>Providers</h2>{snapshot.providers.map((provider) => <div className="listRow" key={provider.address}><code>{provider.address}</code><span>{provider.resource_count} resources</span></div>)}</section>
        <section className="panel"><p className="eyebrow">Module topology</p><h2>Modules</h2>{snapshot.modules.map((module) => <div className="listRow" key={module.address || 'root'}><code>{module.address || 'root'}</code><span>{module.resource_count} resources</span></div>)}</section>
      </div>
      <section><p className="eyebrow">Review queue</p><h2>Findings</h2><Findings findings={result.findings} /></section>
    </div>
  )
}

