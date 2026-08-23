import { ArrowRight, ArrowsLeftRight, CheckCircle, Flask, HardDrives, MagnifyingGlass, ShieldCheck } from '../components/icons'
import type { EngineStatus } from '../api/types'
import type { RecentAnalysis, View } from '../appTypes'
import { EngineCards } from '../components/EngineCards'

const workflows = [
  { view: 'inspect' as const, icon: MagnifyingGlass, title: 'Understand a state', copy: 'Browse modules, resources, providers, and dependencies without exposing attribute values.' },
  { view: 'compare' as const, icon: ArrowsLeftRight, title: 'Explain the delta', copy: 'Separate exact changes from probable identity moves, with confidence and supporting evidence.' },
  { view: 'rehearse' as const, icon: Flask, title: 'Rehearse the next move', copy: 'Compare plans, provider schemas, and engine behavior inside a copied restricted workspace.' },
]

export function Home({ engines, loadingEngines, recent, onNavigate, onOpenRecent }: {
  engines: EngineStatus[]
  loadingEngines: boolean
  recent: RecentAnalysis[]
  onNavigate: (view: View) => void
  onOpenRecent: (analysis: RecentAnalysis) => void
}) {
  return <main className="homePage">
    <section className="hero" aria-labelledby="hero-title">
      <img className="heroImage" src="/assets/planparity-hero.png" alt="Two infrastructure landscapes connected by a bridge beside a local analysis workstation" />
      <div className="heroScrim" />
      <div className="heroContent">
        <h1 id="hero-title">Know what your infrastructure will do.</h1>
        <p>Inspect state, compare snapshots, and rehearse Terraform or OpenTofu locally, without touching the original.</p>
        <div className="heroActions">
          <button className="button buttonLight" type="button" onClick={() => onNavigate('inspect')}>Inspect state <ArrowRight size={17} weight="bold" /></button>
          <button className="button buttonGlass" type="button" onClick={() => onNavigate('rehearse')}>Run rehearsal</button>
        </div>
      </div>
    </section>

    <section className="engineSection pageBand" aria-labelledby="engine-heading">
      <div className="sectionIntro">
        <h2 id="engine-heading">Your local toolchain, ready when you are.</h2>
        <p>PlanParity discovers installed engines and keeps every analysis on this machine.</p>
      </div>
      <EngineCards engines={engines} loading={loadingEngines} />
    </section>

    <section className="workflowSection pageBand" aria-labelledby="workflow-heading">
      <div className="sectionIntro compactIntro">
        <h2 id="workflow-heading">One workbench for every compatibility question.</h2>
      </div>
      <div className="workflowMosaic">
        {workflows.map((workflow, index) => {
          const Icon = workflow.icon
          return <button className={`workflowTile workflowTile${index + 1}`} type="button" key={workflow.view} onClick={() => onNavigate(workflow.view)}>
            <span className="workflowIcon"><Icon size={24} weight="duotone" /></span>
            <span className="workflowCopy"><strong>{workflow.title}</strong><span>{workflow.copy}</span></span>
            <ArrowRight className="workflowArrow" size={20} weight="bold" />
          </button>
        })}
      </div>
    </section>

    <section className="safetySection pageBand" aria-labelledby="safety-heading">
      <div className="safetyLead">
        <ShieldCheck size={34} weight="duotone" />
        <h2 id="safety-heading">Evidence without authority.</h2>
        <p>PlanParity can inspect and rehearse. It has no execution path for apply, destroy, or state push.</p>
      </div>
      <div className="safetyContract">
        <div><CheckCircle size={20} weight="fill" /><strong>Originals stay untouched</strong><span>Work happens against copied state and configuration.</span></div>
        <div><CheckCircle size={20} weight="fill" /><strong>Values stay behind</strong><span>The browser receives sanitized structure, not raw values.</span></div>
        <div><CheckCircle size={20} weight="fill" /><strong>Unknown stays unknown</strong><span>Unverified behavior is never promoted to safe.</span></div>
      </div>
    </section>

    <section className="historySection pageBand" aria-labelledby="history-heading">
      <div className="historyHeader">
        <div><HardDrives size={27} weight="duotone" /><h2 id="history-heading">Recent in this session</h2></div>
        <span>Memory only</span>
      </div>
      {recent.length ? <div className="historyList">{recent.map((item) => <button type="button" onClick={() => onOpenRecent(item)} key={item.id}>
        <span className="historyKind">{item.kind}</span><code>{item.id.slice(0, 12)}</code><time dateTime={item.at}>{new Date(item.at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</time><ArrowRight size={17} />
      </button>)}</div> : <div className="historyEmpty"><span>No analyses yet.</span><p>Your completed work appears here until this local process stops.</p></div>}
    </section>
  </main>
}
