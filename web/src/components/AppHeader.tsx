import { ArrowsLeftRight, Flask, House, List, LockSimple, MagnifyingGlass, X } from './icons'
import { useState } from 'react'
import type { View } from '../appTypes'

const navItems: Array<{ view: Exclude<View, 'home'>; label: string; icon: typeof MagnifyingGlass }> = [
  { view: 'inspect', label: 'Inspect', icon: MagnifyingGlass },
  { view: 'compare', label: 'Compare', icon: ArrowsLeftRight },
  { view: 'rehearse', label: 'Rehearse', icon: Flask },
]

export function AppHeader({ view, authenticated, onNavigate }: { view: View; authenticated: boolean; onNavigate: (view: View) => void }) {
  const [menuOpen, setMenuOpen] = useState(false)
  const navigate = (next: View) => { setMenuOpen(false); onNavigate(next) }

  return (
    <header className={`topbar ${view === 'home' ? 'topbarHero' : ''}`}>
      <button className="brand" type="button" onClick={() => navigate('home')} aria-label="PlanParity home">
        <span className="brandGlyph" aria-hidden="true"><span>P</span><span>P</span></span>
        <span className="brandName">PlanParity</span>
      </button>
      <nav className="desktopNav" aria-label="Primary navigation">
        <button className={view === 'home' ? 'active' : ''} type="button" onClick={() => navigate('home')}><House size={16} weight="bold" />Home</button>
        {navItems.map((item) => {
          const Icon = item.icon
          return <button className={view === item.view ? 'active' : ''} type="button" onClick={() => navigate(item.view)} key={item.view}><Icon size={16} weight="bold" />{item.label}</button>
        })}
      </nav>
      <div className="headerActions">
        <span className={`sessionState ${authenticated ? 'isReady' : 'isWarning'}`}><LockSimple size={15} weight="bold" />{authenticated ? 'Local only' : 'Session missing'}</span>
        <button className="menuButton" type="button" onClick={() => setMenuOpen((open) => !open)} aria-label="Toggle navigation" aria-expanded={menuOpen}>
          {menuOpen ? <X size={21} /> : <List size={22} />}
        </button>
      </div>
      {menuOpen && <nav className="mobileNav" aria-label="Mobile navigation">
        <button type="button" onClick={() => navigate('home')}><House size={18} />Home</button>
        {navItems.map((item) => { const Icon = item.icon; return <button type="button" onClick={() => navigate(item.view)} key={item.view}><Icon size={18} />{item.label}</button> })}
      </nav>}
    </header>
  )
}
