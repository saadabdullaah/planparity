package model

// ComparisonSummary provides stable top-level counts.
type ComparisonSummary struct {
	Added           int `json:"added"`
	Removed         int `json:"removed"`
	Changed         int `json:"changed"`
	Unchanged       int `json:"unchanged"`
	ProbableMoves   int `json:"probable_moves"`
	ProviderChanges int `json:"provider_changes"`
}

// ResourceChange captures a matched resource's structural differences.
type ResourceChange struct {
	Address             string   `json:"address"`
	ProviderChanged     bool     `json:"provider_changed"`
	SchemaChanged       bool     `json:"schema_changed"`
	StructureChanged    bool     `json:"structure_changed"`
	DependenciesChanged bool     `json:"dependencies_changed"`
	ChangedFields       []string `json:"changed_fields"`
}

// MoveSignal is one scored feature in probable-move inference.
type MoveSignal struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Match  bool    `json:"match"`
}

// ProbableMove links disappeared and appeared resource identities.
type ProbableMove struct {
	FromAddress  string       `json:"from_address"`
	ToAddress    string       `json:"to_address"`
	Confidence   float64      `json:"confidence"`
	Signals      []MoveSignal `json:"signals"`
	Ambiguous    bool         `json:"ambiguous"`
	Alternatives []string     `json:"alternatives,omitempty"`
}

// ProviderChange describes provider presence or schema-version changes.
type ProviderChange struct {
	Address              string   `json:"address"`
	BeforeSchemaVersions []uint64 `json:"before_schema_versions"`
	AfterSchemaVersions  []uint64 `json:"after_schema_versions"`
	Change               string   `json:"change"`
}

// StateComparison is a deterministic state-to-state comparison.
type StateComparison struct {
	SchemaVersion   string             `json:"schema_version"`
	BeforeHash      string             `json:"before_hash"`
	AfterHash       string             `json:"after_hash"`
	Added           []ResourceInstance `json:"added"`
	Removed         []ResourceInstance `json:"removed"`
	Changed         []ResourceChange   `json:"changed"`
	ProbableMoves   []ProbableMove     `json:"probable_moves"`
	ProviderChanges []ProviderChange   `json:"provider_changes"`
	Findings        []Finding          `json:"findings"`
	Summary         ComparisonSummary  `json:"summary"`
}
