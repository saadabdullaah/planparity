import type { EngineStatus } from '../api/types'

export function EngineCards({ engines, loading }: { engines: EngineStatus[]; loading: boolean }) {
  return (
    <div className="engineGrid" aria-busy={loading}>
      {loading && <div className="engineCard skeleton">Checking local engines…</div>}
      {!loading && engines.map((engine) => (
        <article className="engineCard" key={engine.kind}>
          <div className="engineHead">
            <span className={`availability ${engine.available ? 'online' : ''}`} />
            <strong>{engine.name || (engine.kind === 'opentofu' ? 'OpenTofu' : 'Terraform')}</strong>
            <span className="pill">{engine.available ? 'Ready' : 'Not found'}</span>
          </div>
          <p>{engine.available ? `${engine.version} · ${engine.platform}` : engine.error?.message ?? 'Configure an explicit binary path.'}</p>
          {engine.path && <code title={engine.path}>{engine.path}</code>}
        </article>
      ))}
    </div>
  )
}

