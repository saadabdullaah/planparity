package model

// PlanAction is an engine-neutral action classification.
type PlanAction string

const (
	PlanNoOp    PlanAction = "NO_OP"
	PlanCreate  PlanAction = "CREATE"
	PlanRead    PlanAction = "READ"
	PlanUpdate  PlanAction = "UPDATE"
	PlanDelete  PlanAction = "DELETE"
	PlanReplace PlanAction = "REPLACE"
	PlanUnknown PlanAction = "UNKNOWN"
)

// PlanResult is a sanitized plan interpretation.
type PlanResult struct {
	SchemaVersion string         `json:"schema_version"`
	Engine        EngineMetadata `json:"engine"`
	Changes       []PlanChange   `json:"changes"`
	Diagnostics   []Diagnostic   `json:"diagnostics"`
	Summary       PlanSummary    `json:"summary"`
	SourceHash    string         `json:"source_hash"`
}

// PlanChange describes actions and structural paths, not before/after values.
type PlanChange struct {
	Address      string       `json:"address"`
	ResourceType string       `json:"resource_type"`
	Provider     string       `json:"provider"`
	Actions      []PlanAction `json:"actions"`
	Replacement  bool         `json:"replacement"`
	ChangedPaths []string     `json:"changed_paths"`
	UnknownPaths []string     `json:"unknown_paths"`
}

// PlanSummary counts normalized actions.
type PlanSummary struct {
	Creates      int `json:"creates"`
	Reads        int `json:"reads"`
	Updates      int `json:"updates"`
	Deletes      int `json:"deletes"`
	Replacements int `json:"replacements"`
	NoOps        int `json:"no_ops"`
	Unknown      int `json:"unknown"`
}

// Diagnostic is sanitized engine feedback.
type Diagnostic struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail,omitempty"`
}
