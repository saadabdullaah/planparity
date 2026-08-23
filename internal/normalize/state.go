package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/saadabdullaah/planparity/internal/model"
)

type rawStateShow struct {
	FormatVersion    string     `json:"format_version"`
	TerraformVersion string     `json:"terraform_version"`
	Lineage          string     `json:"lineage"`
	Serial           uint64     `json:"serial"`
	Values           *rawValues `json:"values"`
}

type rawValues struct {
	Outputs    map[string]rawOutput `json:"outputs"`
	RootModule *rawModule           `json:"root_module"`
}

type rawOutput struct {
	Sensitive bool            `json:"sensitive"`
	Type      json.RawMessage `json:"type"`
}

type rawModule struct {
	Address      string        `json:"address"`
	Resources    []rawResource `json:"resources"`
	ChildModules []rawModule   `json:"child_modules"`
}

type rawResource struct {
	Address         string          `json:"address"`
	Mode            string          `json:"mode"`
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	ProviderName    string          `json:"provider_name"`
	SchemaVersion   uint64          `json:"schema_version"`
	Values          json.RawMessage `json:"values"`
	SensitiveValues json.RawMessage `json:"sensitive_values"`
	DependsOn       []string        `json:"depends_on"`
	Tainted         bool            `json:"tainted"`
	DeposedKey      string          `json:"deposed_key"`
	Status          string          `json:"status"`
}

// State converts Terraform/OpenTofu show JSON to a value-free snapshot.
func State(raw []byte, kind model.EngineKind) (model.StateSnapshot, error) {
	var document rawStateShow
	if err := json.Unmarshal(raw, &document); err != nil {
		return model.StateSnapshot{}, &Error{Code: ErrorInvalidJSON, Message: "engine state output is not valid JSON", Cause: err}
	}
	if err := validateFormatVersion(document.FormatVersion); err != nil {
		return model.StateSnapshot{}, err
	}
	if document.Values != nil && document.Values.RootModule == nil {
		return model.StateSnapshot{}, &Error{Code: ErrorMissingValues, Message: "engine state output does not contain a root module"}
	}

	engineName := "Terraform"
	if kind == model.EngineOpenTofu {
		engineName = "OpenTofu"
	}
	sourceHash := sha256.Sum256(raw)
	snapshot := model.StateSnapshot{
		SchemaVersion: model.SchemaV1,
		FormatVersion: document.FormatVersion,
		Engine: model.EngineMetadata{
			Kind:    kind,
			Name:    engineName,
			Version: document.TerraformVersion,
		},
		Lineage:    document.Lineage,
		Serial:     document.Serial,
		Resources:  []model.ResourceInstance{},
		Providers:  []model.ProviderReference{},
		Modules:    []model.ModuleReference{},
		Outputs:    []model.OutputMetadata{},
		SourceHash: hex.EncodeToString(sourceHash[:]),
	}

	if document.Values != nil {
		snapshot.Outputs = normalizeOutputs(document.Values.Outputs)
		if err := appendModule(&snapshot, *document.Values.RootModule, "", true); err != nil {
			return model.StateSnapshot{}, err
		}
	}
	finalizeState(&snapshot)
	return snapshot, nil
}

func validateFormatVersion(version string) error {
	if version == "" {
		return &Error{Code: ErrorUnsupportedFormat, Message: "engine output is missing format_version"}
	}
	majorText := strings.SplitN(version, ".", 2)[0]
	major, err := strconv.Atoi(majorText)
	if err != nil || major != 1 {
		return &Error{Code: ErrorUnsupportedFormat, Message: fmt.Sprintf("state format version %q is unsupported", version), Cause: err}
	}
	return nil
}

func appendModule(snapshot *model.StateSnapshot, module rawModule, parent string, root bool) error {
	address := module.Address
	if root {
		address = "root"
	}
	snapshot.Modules = append(snapshot.Modules, model.ModuleReference{
		Address:       address,
		ParentAddress: parent,
		ResourceCount: len(module.Resources),
	})

	for _, resource := range module.Resources {
		structureHash, keys, err := fingerprintRaw(resource.Values)
		if err != nil {
			return &Error{Code: ErrorInvalidJSON, Message: fmt.Sprintf("resource %q has invalid value structure", resource.Address), Cause: err}
		}
		dependencies := append([]string{}, resource.DependsOn...)
		sort.Strings(dependencies)
		mode := model.ResourceManaged
		if resource.Mode == string(model.ResourceData) {
			mode = model.ResourceData
		}
		moduleAddress := module.Address
		status := resourceStatus(resource)
		snapshot.Resources = append(snapshot.Resources, model.ResourceInstance{
			Address:         resource.Address,
			ModuleAddress:   moduleAddress,
			Mode:            mode,
			ResourceType:    resource.Type,
			ResourceName:    resource.Name,
			InstanceKey:     instanceKey(resource.Address, resource.Name),
			ProviderAddress: resource.ProviderName,
			SchemaVersion:   resource.SchemaVersion,
			Status:          status,
			Dependencies:    dependencies,
			Structure: model.AttributeStructure{
				TypeFingerprint: structureHash,
				Keys:            keys,
				SensitivePaths:  markerPaths(resource.SensitiveValues),
				UnknownPaths:    []string{},
			},
		})
	}

	for _, child := range module.ChildModules {
		if err := appendModule(snapshot, child, address, false); err != nil {
			return err
		}
	}
	return nil
}

func normalizeOutputs(outputs map[string]rawOutput) []model.OutputMetadata {
	if outputs == nil {
		return []model.OutputMetadata{}
	}
	result := make([]model.OutputMetadata, 0, len(outputs))
	for name, output := range outputs {
		result = append(result, model.OutputMetadata{
			Name:            name,
			Sensitive:       output.Sensitive,
			TypeFingerprint: fingerprintType(output.Type),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func finalizeState(snapshot *model.StateSnapshot) {
	sort.Slice(snapshot.Resources, func(i, j int) bool { return snapshot.Resources[i].Address < snapshot.Resources[j].Address })
	sort.Slice(snapshot.Modules, func(i, j int) bool { return snapshot.Modules[i].Address < snapshot.Modules[j].Address })

	type providerAggregate struct {
		count   int
		schemas map[uint64]struct{}
	}
	aggregates := make(map[string]*providerAggregate)
	for _, resource := range snapshot.Resources {
		aggregate := aggregates[resource.ProviderAddress]
		if aggregate == nil {
			aggregate = &providerAggregate{schemas: make(map[uint64]struct{})}
			aggregates[resource.ProviderAddress] = aggregate
		}
		aggregate.count++
		aggregate.schemas[resource.SchemaVersion] = struct{}{}
	}
	addresses := make([]string, 0, len(aggregates))
	for address := range aggregates {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	snapshot.Providers = make([]model.ProviderReference, 0, len(addresses))
	for _, address := range addresses {
		aggregate := aggregates[address]
		schemas := make([]uint64, 0, len(aggregate.schemas))
		for version := range aggregate.schemas {
			schemas = append(schemas, version)
		}
		sort.Slice(schemas, func(i, j int) bool { return schemas[i] < schemas[j] })
		snapshot.Providers = append(snapshot.Providers, model.ProviderReference{Address: address, ResourceCount: aggregate.count, SchemaVersions: schemas})
	}
}

func resourceStatus(resource rawResource) model.ResourceStatus {
	if resource.DeposedKey != "" || strings.EqualFold(resource.Status, "deposed") {
		return model.ResourceDeposed
	}
	if resource.Tainted || strings.EqualFold(resource.Status, "tainted") {
		return model.ResourceTainted
	}
	if resource.Status != "" && !strings.EqualFold(resource.Status, "normal") {
		return model.ResourceUnknown
	}
	return model.ResourceNormal
}

func instanceKey(address, name string) string {
	needle := name + "["
	index := strings.LastIndex(address, needle)
	if index < 0 {
		return ""
	}
	return address[index+len(name):]
}
