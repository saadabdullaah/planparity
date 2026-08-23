import { fireEvent, render, screen } from '@testing-library/react'
import type { InspectionResult, RehearsalResult, StateComparison } from '../api/types'
import { ComparisonView, RehearsalView } from './ComparisonSummary'
import { StateExplorer } from './StateExplorer'

const resource = {
  address: 'terraform_data.application',
  mode: 'managed',
  resource_type: 'terraform_data',
  resource_name: 'application',
  provider_address: 'terraform.io/builtin/terraform',
  schema_version: 0,
  status: 'NORMAL',
  dependencies: [],
  structure: { type_fingerprint: 'shape', keys: ['input'], sensitive_paths: [], unknown_paths: [] },
}

const finding = {
  id: 'PP-DEST-001',
  category: 'DESTRUCTIVE_CHANGE',
  severity: 'HIGH' as const,
  verification: 'VERIFIED' as const,
  confidence: 1,
  address: resource.address,
  title: 'Resource removed from state',
  explanation: 'The resource exists only in the earlier snapshot.',
  evidence: [{ kind: 'state_presence', message: 'present before and absent after' }],
  recommendation: 'Confirm whether removal was intentional.',
}

it('filters inspected resources and exposes the findings explorer', () => {
  const result: InspectionResult = {
    schema_version: 'planparity/v1',
    snapshot: {
      schema_version: 'planparity/v1',
      format_version: '1.0',
      engine: { kind: 'terraform', name: 'Terraform', version: '1.12.0' },
      serial: 7,
      resources: [resource],
      providers: [{ address: resource.provider_address, resource_count: 1, schema_versions: [0] }],
      modules: [{ address: '', resource_count: 1 }],
      outputs: [],
      source_hash: 'abc',
    },
    findings: [finding],
  }
  render(<StateExplorer result={result} onExport={() => undefined} />)
  expect(screen.getByText('State explorer')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('tab', { name: 'Resources' }))
  expect(screen.getByText(resource.address)).toBeInTheDocument()
  fireEvent.change(screen.getByPlaceholderText('Search address, type, or provider'), { target: { value: 'missing' } })
  expect(screen.getByText('No resources match')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('tab', { name: /Findings/ }))
  expect(screen.getByText(finding.title)).toBeInTheDocument()
})

it('renders large resource inventories progressively', () => {
  const resources = Array.from({ length: 10_000 }, (_, index) => ({
    ...resource,
    address: `terraform_data.item_${String(index).padStart(5, '0')}`,
    resource_name: `item_${String(index).padStart(5, '0')}`,
  }))
  const result: InspectionResult = {
    schema_version: 'planparity/v1',
    snapshot: {
      schema_version: 'planparity/v1', format_version: '1.0',
      engine: { kind: 'terraform', name: 'Terraform', version: '1.12.0' },
      resources, providers: [], modules: [], outputs: [], source_hash: 'large',
    },
    findings: [],
  }
  render(<StateExplorer result={result} onExport={() => undefined} />)

  fireEvent.click(screen.getByRole('tab', { name: 'Resources' }))
  expect(screen.getByText('terraform_data.item_00249')).toBeInTheDocument()
  expect(screen.queryByText('terraform_data.item_00250')).not.toBeInTheDocument()
  expect(screen.getByText('Showing the first 250 of 10,000 matching resources.')).toBeInTheDocument()

  fireEvent.click(screen.getByRole('button', { name: 'Load 250 more' }))
  expect(screen.getByText('terraform_data.item_00499')).toBeInTheDocument()

  fireEvent.change(screen.getByPlaceholderText('Search address, type, or provider'), { target: { value: 'item_09999' } })
  expect(screen.getByText('terraform_data.item_09999')).toBeInTheDocument()
  expect(screen.getByText('1 of 10000 resources')).toBeInTheDocument()
})

it('shows both sides of a changed resource and move evidence', () => {
  const after = { ...resource, schema_version: 1 }
  const result: StateComparison = {
    schema_version: 'planparity/v1',
    before_hash: 'before',
    after_hash: 'after',
    added: [],
    removed: [],
    changed: [{ address: resource.address, before: resource, after, provider_changed: false, schema_changed: true, structure_changed: false, dependencies_changed: false, changed_fields: ['schema_version'] }],
    probable_moves: [{ from_address: 'terraform_data.old', to_address: 'module.app.terraform_data.old', from: { ...resource, address: 'terraform_data.old' }, to: { ...resource, address: 'module.app.terraform_data.old', module_address: 'module.app' }, confidence: .95, ambiguous: false, alternatives: [], signals: [{ name: 'resource_type', weight: .25, match: true }] }],
    provider_changes: [],
    findings: [finding],
    summary: { added: 0, removed: 0, changed: 1, unchanged: 0, probable_moves: 1, provider_changes: 0 },
  }
  render(<ComparisonView result={result} onExport={() => undefined} />)
  fireEvent.click(screen.getByRole('tab', { name: /Changes/ }))
  expect(screen.getByText('Before')).toBeInTheDocument()
  expect(screen.getByText('After')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('tab', { name: /Moves/ }))
  expect(screen.getByText('95%')).toBeInTheDocument()
  expect(screen.getByText(/resource type/)).toBeInTheDocument()
})

it('renders Terraform and OpenTofu side by side for parity rehearsals', () => {
  const summary = { creates: 0, updates: 1, deletes: 0, replacements: 0, no_ops: 1, reads: 0, unknown: 0 }
  const result: RehearsalResult = {
    schema_version: 'planparity/v1',
    kind: 'MIGRATE',
    execution_mode: 'LOCAL-RESTRICTED',
    baseline: { schema_version: 'planparity/v1', engine: { kind: 'terraform', name: 'Terraform', version: '1.12.0' }, changes: [], diagnostics: [], summary, source_hash: 'baseline' },
    candidate: { schema_version: 'planparity/v1', engine: { kind: 'opentofu', name: 'OpenTofu', version: '1.10.0' }, changes: [], diagnostics: [], summary, source_hash: 'candidate' },
    comparison: { equivalent: true, baseline_summary: summary, candidate_summary: summary, only_baseline: [], only_candidate: [], changed: [] },
    schema_differences: [],
    checks: [{ name: 'Plan actions', verification: 'VERIFIED', message: 'Normalized actions are equivalent.' }],
    findings: [],
    limitations: ['Cloud API behavior is not exercised.'],
  }
  render(<RehearsalView result={result} onExport={() => undefined} />)
  expect(screen.getByRole('heading', { name: 'Terraform and OpenTofu parity' })).toBeInTheDocument()
  expect(screen.getByText('Plan actions are equivalent')).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'Terraform' })).toBeInTheDocument()
  expect(screen.getByRole('heading', { name: 'OpenTofu' })).toBeInTheDocument()
})

it('tolerates legacy null lists and renders hostile text without HTML execution', () => {
  const hostile = '<img src=x onerror="globalThis.__planparityInjected=true">'
  const legacyResource = { ...resource, dependencies: null, structure: { ...resource.structure, keys: null } } as unknown as typeof resource
  const result: InspectionResult = {
    schema_version: 'planparity/v1',
    snapshot: {
      schema_version: 'planparity/v1', format_version: '1.0',
      engine: { kind: 'terraform', name: 'Terraform', version: '1.12.0' },
      resources: [legacyResource], providers: [], modules: [], outputs: [], source_hash: 'legacy',
    },
    findings: [{ ...finding, title: hostile, explanation: hostile, evidence: null as unknown as typeof finding.evidence }],
  }
  const { container } = render(<StateExplorer result={result} onExport={() => undefined} />)
  fireEvent.click(screen.getByRole('tab', { name: 'Resources' }))
  expect(screen.getAllByText('0')).toHaveLength(2)
  fireEvent.click(screen.getByRole('tab', { name: /Findings/ }))
  expect(screen.getAllByText(hostile)).toHaveLength(2)
  expect(container.querySelector('img')).toBeNull()
  expect((globalThis as typeof globalThis & { __planparityInjected?: boolean }).__planparityInjected).not.toBe(true)
})
