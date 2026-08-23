import type { Finding } from '../api/types'

export function Findings({ findings }: { findings: Finding[] }) {
  if (!findings.length) return <div className="emptyState">No findings were produced.</div>
  return (
    <div className="findingList">
      {findings.map((finding) => (
        <article className="finding" key={finding.id}>
          <div><span className={`severity severity-${finding.severity.toLowerCase()}`}>{finding.severity}</span><span className="verification">{finding.verification}</span></div>
          <h3>{finding.title}</h3>
          {finding.address && <code>{finding.address}</code>}
          <p>{finding.explanation}</p>
          <p className="recommendation">Next: {finding.recommendation}</p>
        </article>
      ))}
    </div>
  )
}
