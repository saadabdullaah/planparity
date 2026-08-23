package normalize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func TestProviderSchemasExposeOnlyFingerprintsAndTypeNames(t *testing.T) {
	raw := []byte(`{
		"format_version": "1.0",
		"provider_schemas": {
			"registry.terraform.io/hashicorp/example": {
				"provider": {"version": 0, "block": {"description": "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9"}},
				"resource_schemas": {
					"example_item": {
						"version": 2,
						"block": {"attributes": {"token": {"description": "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9", "type": "string"}}}
					}
				}
			}
		}
	}`)
	result, err := ProviderSchemas(raw, model.EngineMetadata{Kind: model.EngineTerraform, Name: "Terraform", Version: "1.12.1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Providers) != 1 || len(result.Providers[0].ResourceTypes) != 1 || result.Providers[0].ResourceTypes[0].Type != "example_item" {
		t.Fatalf("unexpected provider normalization: %#v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), secretCanary) || strings.Contains(string(encoded), "description") {
		t.Fatalf("provider schema detail escaped normalization: %s", encoded)
	}
}
