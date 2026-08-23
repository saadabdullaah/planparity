package model

// RehearsalKind identifies a plan, upgrade, or cross-engine workflow.
type RehearsalKind string

const (
	RehearsalPlan    RehearsalKind = "PLAN"
	RehearsalUpgrade RehearsalKind = "UPGRADE"
	RehearsalMigrate RehearsalKind = "MIGRATE"
)

// CompatibilityCheck records one verified or bounded assertion.
type CompatibilityCheck struct {
	Name         string            `json:"name"`
	Verification VerificationState `json:"verification"`
	Message      string            `json:"message"`
}

// PlanComparison describes engine-to-engine action differences.
type PlanComparison struct {
	Equivalent       bool         `json:"equivalent"`
	BaselineSummary  PlanSummary  `json:"baseline_summary"`
	CandidateSummary PlanSummary  `json:"candidate_summary"`
	OnlyBaseline     []PlanChange `json:"only_baseline"`
	OnlyCandidate    []PlanChange `json:"only_candidate"`
	Changed          []PlanChange `json:"changed"`
}

// RehearsalResult is the public, sanitized result of restricted execution.
type RehearsalResult struct {
	SchemaVersion     string                     `json:"schema_version"`
	Kind              RehearsalKind              `json:"kind"`
	ExecutionMode     string                     `json:"execution_mode"`
	Baseline          *PlanResult                `json:"baseline,omitempty"`
	Candidate         *PlanResult                `json:"candidate,omitempty"`
	Comparison        *PlanComparison            `json:"comparison,omitempty"`
	BaselineSchemas   *ProviderSchemaSet         `json:"baseline_schemas,omitempty"`
	CandidateSchemas  *ProviderSchemaSet         `json:"candidate_schemas,omitempty"`
	SchemaDifferences []ProviderSchemaDifference `json:"schema_differences"`
	Checks            []CompatibilityCheck       `json:"checks"`
	Findings          []Finding                  `json:"findings"`
	Limitations       []string                   `json:"limitations"`
}
