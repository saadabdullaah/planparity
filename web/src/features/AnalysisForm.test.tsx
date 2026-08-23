import { fireEvent, render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import { AnalysisForm } from './AnalysisForm'

it('clears stale rehearsal evidence when the workflow mode changes', () => {
  const onModeChange = vi.fn()
  render(<AnalysisForm view="rehearse" busy={false} onModeChange={onModeChange} onSubmit={() => undefined} />)

  fireEvent.click(screen.getByRole('button', { name: 'Upgrade' }))
  expect(onModeChange).toHaveBeenCalledTimes(1)

  fireEvent.click(screen.getByRole('button', { name: 'Upgrade' }))
  expect(onModeChange).toHaveBeenCalledTimes(1)

  fireEvent.click(screen.getByRole('button', { name: 'Parity' }))
  expect(onModeChange).toHaveBeenCalledTimes(2)
})
