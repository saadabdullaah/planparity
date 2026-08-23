import type { RehearsalResult, StateComparison } from '../api/types'
import { Findings } from '../components/Findings'

function SummaryCard({ label, value, tone = '' }: { label: string; value: number | string; tone?: string }) {
  return <div className={`metric ${tone}`}><span>{label}</span><strong>{value}</strong></div>
}

export function ComparisonView({ result, onExport }: { result: StateComparison; onExport: () => void }) {
  return (
    <div className="resultStack">
      <div className="resultTitle"><div><p className="eyebrow">Deterministic comparison</p><h2>Snapshot delta</h2></div><button type="button" className="secondary" onClick={onExport}>Export JSON</button></div>
      <section className="metrics">
        <SummaryCard label="Added" value={result.summary.added} tone="positive" />
        <SummaryCard label="Removed" value={result.summary.removed} tone="negative" />
        <SummaryCard label="Changed" value={result.summary.changed} tone="warning" />
        <SummaryCard label="Probable moves" value={result.summary.probable_moves} />
      </section>
      <section className="panel"><p className="eyebrow">Identity changes</p><h2>Probable moves</h2>{result.probable_moves.length ? result.probable_moves.map((move) => <div className="moveRow" key={`${move.from_address}-${move.to_address}`}><code>{move.from_address}</code><span>→</span><code>{move.to_address}</code><strong>{Math.round(move.confidence * 100)}%</strong></div>) : <div className="emptyState">No probable moves detected.</div>}</section>
      <section><p className="eyebrow">Review queue</p><h2>Findings</h2><Findings findings={result.findings} /></section>
    </div>
  )
}

export function RehearsalView({ result, onExport }: { result: RehearsalResult; onExport: () => void }) {
  const summary = result.candidate?.summary ?? result.baseline?.summary
  return (
    <div className="resultStack">
      <div className="resultTitle"><div><p className="eyebrow">{result.execution_mode}</p><h2>{result.kind.toLowerCase()} rehearsal</h2></div><button type="button" className="secondary" onClick={onExport}>Export JSON</button></div>
      {summary && <section className="metrics"><SummaryCard label="Creates" value={summary.creates} tone="positive" /><SummaryCard label="Updates" value={summary.updates} tone="warning" /><SummaryCard label="Deletes" value={summary.deletes} tone="negative" /><SummaryCard label="Replacements" value={summary.replacements} tone="negative" /></section>}
      <section className="panel"><p className="eyebrow">Compatibility evidence</p><h2>Checks</h2>{result.checks.map((check) => <div className="checkRow" key={check.name}><span className={`verification verification-${check.verification.toLowerCase()}`}>{check.verification}</span><div><strong>{check.name}</strong><p>{check.message}</p></div></div>)}</section>
      <section><p className="eyebrow">Review queue</p><h2>Findings</h2><Findings findings={result.findings} /></section>
      {!!result.limitations.length && <section className="panel limitations"><p className="eyebrow">Known boundaries</p><h2>Limitations</h2><ul>{result.limitations.map((item) => <li key={item}>{item}</li>)}</ul></section>}
    </div>
  )
}
