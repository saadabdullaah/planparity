package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

type rawProviderSchemas struct {
	FormatVersion   string                       `json:"format_version"`
	ProviderSchemas map[string]rawProviderSchema `json:"provider_schemas"`
}

type rawProviderSchema struct {
	Provider          json.RawMessage          `json:"provider"`
	ResourceSchemas   map[string]rawTypeSchema `json:"resource_schemas"`
	DataSourceSchemas map[string]rawTypeSchema `json:"data_source_schemas"`
}

type rawTypeSchema struct {
	Version uint64          `json:"version"`
	Block   json.RawMessage `json:"block"`
}

// ProviderSchemas converts raw schema JSON to hashes and type names only.
func ProviderSchemas(raw []byte, metadata model.EngineMetadata) (model.ProviderSchemaSet, error) {
	var document rawProviderSchemas
	if err := json.Unmarshal(raw, &document); err != nil {
		return model.ProviderSchemaSet{}, &Error{Code: ErrorInvalidJSON, Message: "engine provider schema output is not valid JSON", Cause: err}
	}
	if document.ProviderSchemas == nil {
		return model.ProviderSchemaSet{}, &Error{Code: ErrorMissingValues, Message: "engine output does not contain provider schemas"}
	}
	sourceHash := sha256.Sum256(raw)
	result := model.ProviderSchemaSet{
		SchemaVersion: model.SchemaV1, Engine: metadata, Providers: []model.ProviderSchemaFingerprint{}, SourceHash: hex.EncodeToString(sourceHash[:]),
	}
	addresses := make([]string, 0, len(document.ProviderSchemas))
	for address := range document.ProviderSchemas {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	for _, address := range addresses {
		provider := document.ProviderSchemas[address]
		result.Providers = append(result.Providers, model.ProviderSchemaFingerprint{
			Address: address, Fingerprint: canonicalFingerprint(provider.Provider),
			ResourceTypes: normalizeTypeSchemas(provider.ResourceSchemas), DataSourceTypes: normalizeTypeSchemas(provider.DataSourceSchemas),
		})
	}
	return result, nil
}

func normalizeTypeSchemas(schemas map[string]rawTypeSchema) []model.TypeSchemaFingerprint {
	if schemas == nil {
		return []model.TypeSchemaFingerprint{}
	}
	types := make([]string, 0, len(schemas))
	for resourceType := range schemas {
		types = append(types, resourceType)
	}
	sort.Strings(types)
	result := make([]model.TypeSchemaFingerprint, 0, len(types))
	for _, resourceType := range types {
		schema := schemas[resourceType]
		result = append(result, model.TypeSchemaFingerprint{Type: resourceType, Version: schema.Version, Fingerprint: canonicalFingerprint(schema.Block)})
	}
	return result
}

func canonicalFingerprint(raw json.RawMessage) string {
	if len(raw) == 0 {
		return hashString("null")
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return hashString("invalid")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return hashString("invalid")
	}
	return hashString(string(canonical))
}
