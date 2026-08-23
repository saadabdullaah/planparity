export type EngineKind = 'terraform' | 'opentofu'
export type JobStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED'
export type Severity = 'INFO' | 'WARNING' | 'HIGH' | 'CRITICAL' | 'UNKNOWN'

export interface PublicFailure { code: string; operation?: string; message: string; retryable?: boolean }
export interface EngineStatus {
  kind: EngineKind
  name: string
  available: boolean
  path?: string
  version?: string
  platform?: string
  capabilities: Record<string, boolean>
  checked_at: string
  error?: PublicFailure
}
export interface JobSnapshot {
  id: string
  kind: string
  status: JobStatus
  created_at: string
  started_at?: string
  completed_at?: string
  last_event?: JobEvent
  failure?: PublicFailure
}
export interface JobEvent { id: number; type: string; message?: string; at: string }
export interface JobAccepted { job_id: string; status_url: string; events_url: string; result_url: string }

export interface AttributeStructure {
  type_fingerprint: string
  keys: string[]
  sensitive_paths: string[]
  unknown_paths: string[]
}
export interface ResourceInstance {
  address: string
  module_address?: string
  mode: string
  resource_type: string
  resource_name: string
  instance_key?: string
  provider_address: string
  schema_version: number
  status: string
  dependencies: string[]
  structure: AttributeStructure
}
export interface ProviderReference { address: string; resource_count: number; schema_versions: number[] }
export interface ModuleReference { address: string; parent_address?: string; resource_count: number }
export interface OutputMetadata { name: string; sensitive: boolean; type_fingerprint: string }
export interface Finding {
  id: string
  category: string
  severity: Severity
  verification: string
  confidence: number
  address?: string
  title: string
  explanation: string
  evidence: Array<{ kind: string; message: string }>
  recommendation: string
}
export interface StateSnapshot {
  schema_version: string
  format_version: string
  engine: { kind: EngineKind; name: string; version: string; platform?: string }
  lineage?: string
  serial?: number
  resources: ResourceInstance[]
  providers: ProviderReference[]
  modules: ModuleReference[]
  outputs: OutputMetadata[]
  source_hash: string
}
export interface InspectionResult { schema_version: string; snapshot: StateSnapshot; findings: Finding[] }

export interface ComparisonSummary {
  added: number
  removed: number
  changed: number
  unchanged: number
  probable_moves: number
  provider_changes: number
}
export interface StateComparison {
  schema_version: string
  before_hash: string
  after_hash: string
  added: ResourceInstance[]
  removed: ResourceInstance[]
  changed: Array<{ address: string; changed_fields: string[] }>
  probable_moves: Array<{ from_address: string; to_address: string; confidence: number; ambiguous: boolean }>
  provider_changes: Array<{ address: string; change: string }>
  findings: Finding[]
  summary: ComparisonSummary
}

export interface PlanSummary { creates: number; updates: number; deletes: number; replacements: number; no_ops: number; reads: number; unknown: number }
export interface RehearsalResult {
  schema_version: string
  kind: string
  execution_mode: string
  baseline?: { summary: PlanSummary }
  candidate?: { summary: PlanSummary }
  comparison?: { equivalent: boolean; baseline_summary: PlanSummary; candidate_summary: PlanSummary }
  checks: Array<{ name: string; verification: string; message: string }>
  findings: Finding[]
  limitations: string[]
  schema_differences: Array<{ provider_address: string; change: string }>
}

export type AnalysisResult = InspectionResult | StateComparison | RehearsalResult
