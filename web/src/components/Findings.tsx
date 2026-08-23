import { CaretDown, CheckCircle, FunnelSimple, MagnifyingGlass, ShieldWarning, WarningCircle } from './icons'
import { useMemo, useState } from 'react'
import type { Finding, Severity } from '../api/types'

const severityOrder: Severity[] = ['CRITICAL', 'HIGH', 'WARNING', 'INFO', 'UNKNOWN']

function FindingIcon({ severity }: { severity: Severity }) {
  if (severity === 'CRITICAL' || severity === 'HIGH') return <ShieldWarning size={21} weight="duotone" />
  if (severity === 'WARNING' || severity === 'UNKNOWN') return <WarningCircle size={21} weight="duotone" />
  return <CheckCircle size={21} weight="duotone" />
}

export function Findings({ findings }: { findings: Finding[] }) {
  const [query, setQuery] = useState('')
  const [severity, setSeverity] = useState('ALL')
  const [category, setCategory] = useState('ALL')
  const [verification, setVerification] = useState('ALL')
  const categories = useMemo(() => [...new Set(findings.map((finding) => finding.category))].sort(), [findings])
  const verifications = useMemo(() => [...new Set(findings.map((finding) => finding.verification))].sort(), [findings])
  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return findings.filter((finding) => {
      if (severity !== 'ALL' && finding.severity !== severity) return false
      if (category !== 'ALL' && finding.category !== category) return false
      if (verification !== 'ALL' && finding.verification !== verification) return false
      if (!needle) return true
      return [finding.id, finding.title, finding.address, finding.explanation, finding.recommendation].join(' ').toLowerCase().includes(needle)
    })
  }, [category, findings, query, severity, verification])

  if (!findings.length) return <div className="emptyState successEmpty"><CheckCircle size={28} weight="duotone" /><strong>No findings</strong><p>This analysis did not produce a review item.</p></div>

  return <div className="findingsExplorer">
    <div className="findingTools">
      <label className="searchField"><MagnifyingGlass size={17} /><span className="srOnly">Search findings</span><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search ID, resource, or message" /></label>
      <label><FunnelSimple size={16} /><span className="srOnly">Severity</span><select aria-label="Filter findings by severity" value={severity} onChange={(event) => setSeverity(event.target.value)}><option value="ALL">All severities</option>{severityOrder.map((item) => <option value={item} key={item}>{item}</option>)}</select></label>
      <label><span className="srOnly">Category</span><select aria-label="Filter findings by category" value={category} onChange={(event) => setCategory(event.target.value)}><option value="ALL">All categories</option>{categories.map((item) => <option value={item} key={item}>{item.replaceAll('_', ' ')}</option>)}</select></label>
      <label><span className="srOnly">Verification</span><select aria-label="Filter findings by verification" value={verification} onChange={(event) => setVerification(event.target.value)}><option value="ALL">All evidence</option>{verifications.map((item) => <option value={item} key={item}>{item}</option>)}</select></label>
    </div>
    <div className="findingCount">{filtered.length} of {findings.length} findings</div>
    {filtered.length ? <div className="findingList">{filtered.map((finding) => <details className={'finding severityBorder-' + finding.severity.toLowerCase()} key={finding.id + '-' + (finding.address ?? '')}>
      <summary>
        <span className={'findingIcon severityText-' + finding.severity.toLowerCase()}><FindingIcon severity={finding.severity} /></span>
        <span className="findingSummary"><span><code>{finding.id}</code><span className={'severity severity-' + finding.severity.toLowerCase()}>{finding.severity}</span><span className="verification">{finding.verification}</span></span><strong>{finding.title}</strong>{finding.address && <code className="findingAddress">{finding.address}</code>}</span>
        <span className="confidence">{Math.round(finding.confidence * 100)}% confidence</span><CaretDown className="findingCaret" size={17} />
      </summary>
      <div className="findingBody"><p>{finding.explanation}</p><div className="evidenceBlock"><strong>Evidence</strong>{finding.evidence?.length ? <ul>{finding.evidence.map((item, index) => <li key={item.kind + '-' + index}><span>{item.kind.replaceAll('_', ' ')}</span><p>{item.message}</p></li>)}</ul> : <p>No additional evidence was reported.</p>}</div><div className="recommendation"><strong>Recommended review</strong><p>{finding.recommendation}</p></div></div>
    </details>)}</div> : <div className="emptyState"><MagnifyingGlass size={26} /><strong>No matches</strong><p>Adjust the finding filters to widen the review queue.</p></div>}
  </div>
}
