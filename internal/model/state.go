package model

// StateSnapshot is the sanitized, deterministic representation of engine state.
type StateSnapshot struct {
	SchemaVersion string              `json:"schema_version"`
	FormatVersion string              `json:"format_version"`
	Engine        EngineMetadata      `json:"engine"`
	Lineage       string              `json:"lineage,omitempty"`
	Serial        uint64              `json:"serial,omitempty"`
	Resources     []ResourceInstance  `json:"resources"`
	Providers     []ProviderReference `json:"providers"`
	Modules       []ModuleReference   `json:"modules"`
	Outputs       []OutputMetadata    `json:"outputs"`
	SourceHash    string              `json:"source_hash"`
}

// InspectionResult combines a sanitized snapshot with integrity findings.
type InspectionResult struct {
	SchemaVersion string        `json:"schema_version"`
	Snapshot      StateSnapshot `json:"snapshot"`
	Findings      []Finding     `json:"findings"`
}

// ResourceStatus describes lifecycle metadata exposed by an engine.
type ResourceStatus string

const (
	ResourceNormal  ResourceStatus = "NORMAL"
	ResourceTainted ResourceStatus = "TAINTED"
	ResourceDeposed ResourceStatus = "DEPOSED"
	ResourceUnknown ResourceStatus = "UNKNOWN"
)

// ResourceMode distinguishes managed resources from data sources.
type ResourceMode string

const (
	ResourceManaged ResourceMode = "managed"
	ResourceData    ResourceMode = "data"
)

// ResourceInstance contains resource identity and structure, never attribute values.
type ResourceInstance struct {
	Address         string             `json:"address"`
	ModuleAddress   string             `json:"module_address,omitempty"`
	Mode            ResourceMode       `json:"mode"`
	ResourceType    string             `json:"resource_type"`
	ResourceName    string             `json:"resource_name"`
	InstanceKey     string             `json:"instance_key,omitempty"`
	ProviderAddress string             `json:"provider_address"`
	SchemaVersion   uint64             `json:"schema_version"`
	Status          ResourceStatus     `json:"status"`
	Dependencies    []string           `json:"dependencies"`
	Structure       AttributeStructure `json:"structure"`
}

// AttributeStructure is a value-free fingerprint of an attribute object.
type AttributeStructure struct {
	TypeFingerprint string   `json:"type_fingerprint"`
	Keys            []string `json:"keys"`
	SensitivePaths  []string `json:"sensitive_paths"`
	UnknownPaths    []string `json:"unknown_paths"`
}

// ProviderReference summarizes provider usage without configuration values.
type ProviderReference struct {
	Address        string   `json:"address"`
	ResourceCount  int      `json:"resource_count"`
	SchemaVersions []uint64 `json:"schema_versions"`
}

// ModuleReference describes module containment.
type ModuleReference struct {
	Address       string `json:"address"`
	ParentAddress string `json:"parent_address,omitempty"`
	ResourceCount int    `json:"resource_count"`
}

// OutputMetadata contains output structure and sensitivity only.
type OutputMetadata struct {
	Name            string `json:"name"`
	Sensitive       bool   `json:"sensitive"`
	TypeFingerprint string `json:"type_fingerprint"`
}
