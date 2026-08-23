import { CheckCircle, Cpu, WarningCircle } from './icons'
import type { EngineStatus } from '../api/types'

export function EngineCards({ engines, loading }: { engines: EngineStatus[]; loading: boolean }) {
  if (loading) return <div className="engineGrid" aria-busy="true"><div className="engineCard engineSkeleton"><span /><span /><span /></div><div className="engineCard engineSkeleton"><span /><span /><span /></div></div>
  return <div className="engineGrid">
    {engines.map((engine) => <article className={'engineCard ' + (engine.available ? 'engineReady' : 'engineMissing')} key={engine.kind}>
      <div className="engineIdentity">
        <span className="engineIcon"><Cpu size={23} weight="duotone" /></span>
        <div><strong>{engine.name || (engine.kind === 'opentofu' ? 'OpenTofu' : 'Terraform')}</strong><span>{engine.available ? engine.version : 'Not detected'}</span></div>
      </div>
      <span className="engineBadge">{engine.available ? <CheckCircle size={16} weight="fill" /> : <WarningCircle size={16} weight="fill" />}{engine.available ? 'Ready' : 'Needs setup'}</span>
      <p>{engine.available ? engine.platform : engine.error?.message ?? 'Configure an explicit binary path when you start an analysis.'}</p>
      {engine.path && <code title={engine.path}>{engine.path}</code>}
    </article>)}
  </div>
}
