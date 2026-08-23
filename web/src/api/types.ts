export type EngineKind = 'terraform' | 'opentofu'
export type JobStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED'
export type Severity = 'INFO' | 'WARNING' | 'HIGH' | 'CRITICAL' | 'UNKNOWN'
export type VerificationState = 'VERIFIED' | 'INFERRED' | 'UNVERIFIED' | 'UNSUPPORTED' | 'FAILED'
export type PlanAction = 'NO_OP' | 'CREATE' | 'READ' | 'UPDATE' | 'DELETE' | 'REPLACE' | 'UNKNOWN'

export interface PublicFailure { code: string; operation?: string; message: string; retryable?: boolean }
export interface EngineMetadata { kind: EngineKind; name: string; version: string; platform?: string }
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
export interface Evidence { kind: string; message: string }
export interface Finding {
  id: string
  category: string
  severity: Severity
  verification: VerificationState
  confidence: number
  address?: string
  title: string
  explanation: string
  evidence: Evidence[]
  recommendation: string
}
export interface StateSnapshot {
  schema_version: string
  format_version: string
  engine: EngineMetadata
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
export interface ResourceChange {
  address: string
  before: ResourceInstance
  after: ResourceInstance
  provider_changed: boolean
  schema_changed: boolean
  structure_changed: boolean
  dependencies_changed: boolean
  changed_fields: string[]
}
export interface MoveSignal { name: string; weight: number; match: boolean }
export interface ProbableMove {
  from_address: string
  to_address: string
  from: ResourceInstance
  to: ResourceInstance
  confidence: number
  signals: MoveSignal[]
  ambiguous: boolean
  alternatives?: string[]
}
export interface ProviderChange {
  address: string
  before_schema_versions: number[]
  after_schema_versions: number[]
  change: string
}
export interface StateComparison {
  schema_version: string
  before_hash: string
  after_hash: string
  added: ResourceInstance[]
  removed: ResourceInstance[]
  changed: ResourceChange[]
  probable_moves: ProbableMove[]
  provider_changes: ProviderChange[]
  findings: Finding[]
  summary: ComparisonSummary
}

export interface PlanSummary {
  creates: number
  updates: number
  deletes: number
  replacements: number
  no_ops: number
  reads: number
  unknown: number
}
export interface Diagnostic { severity: string; summary: string; detail?: string }
export interface PlanChange {
  address: string
  resource_type: string
  provider: string
  actions: PlanAction[]
  replacement: boolean
  changed_paths: string[]
  unknown_paths: string[]
}
export interface PlanResult {
  schema_version: string
  engine: EngineMetadata
  changes: PlanChange[]
  diagnostics: Diagnostic[]
  summary: PlanSummary
  source_hash: string
}
export interface PlanComparison {
  equivalent: boolean
  baseline_summary: PlanSummary
  candidate_summary: PlanSummary
  only_baseline: PlanChange[]
  only_candidate: PlanChange[]
  changed: PlanChange[]
}
export interface TypeSchemaFingerprint { type: string; version: number; fingerprint: string }
export interface ProviderSchemaFingerprint {
  address: string
  fingerprint: string
  resource_types: TypeSchemaFingerprint[]
  data_source_types: TypeSchemaFingerprint[]
}
export interface ProviderSchemaSet {
  schema_version: string
  engine: EngineMetadata
  providers: ProviderSchemaFingerprint[]
  source_hash: string
}
export interface ProviderSchemaDifference {
  address: string
  change: string
  changed_resource_types: string[]
  changed_data_source_types: string[]
}
export interface CompatibilityCheck { name: string; verification: VerificationState; message: string }
export interface RehearsalResult {
  schema_version: string
  kind: 'PLAN' | 'UPGRADE' | 'MIGRATE'
  execution_mode: 'LOCAL-RESTRICTED'
  baseline?: PlanResult
  candidate?: PlanResult
  comparison?: PlanComparison
  baseline_schemas?: ProviderSchemaSet
  candidate_schemas?: ProviderSchemaSet
  checks: CompatibilityCheck[]
  findings: Finding[]
  limitations: string[]
  schema_differences: ProviderSchemaDifference[]
}

export type AnalysisResult = InspectionResult | StateComparison | RehearsalResult
