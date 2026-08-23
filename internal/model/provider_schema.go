package model

// ProviderSchemaSet is a value-free fingerprint set from one initialized engine workspace.
type ProviderSchemaSet struct {
	SchemaVersion string                      `json:"schema_version"`
	Engine        EngineMetadata              `json:"engine"`
	Providers     []ProviderSchemaFingerprint `json:"providers"`
	SourceHash    string                      `json:"source_hash"`
}

// ProviderSchemaFingerprint summarizes a provider and its exposed type schemas.
type ProviderSchemaFingerprint struct {
	Address         string                  `json:"address"`
	Fingerprint     string                  `json:"fingerprint"`
	ResourceTypes   []TypeSchemaFingerprint `json:"resource_types"`
	DataSourceTypes []TypeSchemaFingerprint `json:"data_source_types"`
}

// TypeSchemaFingerprint records schema version and a hash, never descriptions or defaults.
type TypeSchemaFingerprint struct {
	Type        string `json:"type"`
	Version     uint64 `json:"version"`
	Fingerprint string `json:"fingerprint"`
}

// ProviderSchemaDifference describes one provider-level fingerprint difference.
type ProviderSchemaDifference struct {
	Address                string   `json:"address"`
	Change                 string   `json:"change"`
	ChangedResourceTypes   []string `json:"changed_resource_types"`
	ChangedDataSourceTypes []string `json:"changed_data_source_types"`
}
