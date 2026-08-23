import { fireEvent, render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import App from './App'

beforeEach(() => {
  window.location.hash = '#token=test-session-token'
  vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response(JSON.stringify({ engines: [] }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }))))
  vi.stubGlobal('scrollTo', vi.fn())
})

afterEach(() => vi.unstubAllGlobals())

describe('App', () => {
  it('states the non-mutation safety contract and removes the token fragment', async () => {
    render(<App />)
    expect(screen.getByText('No apply. No destroy. No state push.')).toBeInTheDocument()
    expect(window.location.hash).toBe('')
    expect(await screen.findByText('Engine readiness')).toBeInTheDocument()
  })

  it('opens the inspection workflow without persisting browser data', () => {
    const storageSpy = vi.spyOn(Storage.prototype, 'setItem')
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Inspect state' }))
    expect(screen.getByRole('heading', { name: 'Inspect a state snapshot' })).toBeInTheDocument()
    expect(screen.getByText('Refresh disabled · backend disabled · no apply path')).toBeInTheDocument()
    expect(storageSpy).not.toHaveBeenCalled()
  })
})
