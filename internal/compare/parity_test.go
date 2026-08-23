package compare

import (
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func TestPlansDetectsCandidateBehaviorDifference(t *testing.T) {
	baseline := model.PlanResult{Changes: []model.PlanChange{{Address: "terraform_data.example", ResourceType: "terraform_data", Provider: "terraform.io/builtin/terraform", Actions: []model.PlanAction{model.PlanNoOp}, ChangedPaths: []string{}, UnknownPaths: []string{}}}, Summary: model.PlanSummary{NoOps: 1}}
	candidate := model.PlanResult{Changes: []model.PlanChange{{Address: "terraform_data.example", ResourceType: "terraform_data", Provider: "terraform.io/builtin/terraform", Actions: []model.PlanAction{model.PlanDelete}, ChangedPaths: []string{}, UnknownPaths: []string{}}}, Summary: model.PlanSummary{Deletes: 1}}
	result := Plans(baseline, candidate)
	if result.Equivalent || len(result.Changed) != 1 || result.Changed[0].Actions[0] != model.PlanDelete {
		t.Fatalf("candidate plan difference was not preserved: %#v", result)
	}
}

func TestProviderSchemasDetectsTypeAndProviderChangesDeterministically(t *testing.T) {
	baseline := model.ProviderSchemaSet{Providers: []model.ProviderSchemaFingerprint{{Address: "registry.terraform.io/hashicorp/example", Fingerprint: "before", ResourceTypes: []model.TypeSchemaFingerprint{{Type: "example_item", Version: 0, Fingerprint: "v0"}}, DataSourceTypes: []model.TypeSchemaFingerprint{}}}}
	candidate := model.ProviderSchemaSet{Providers: []model.ProviderSchemaFingerprint{{Address: "registry.terraform.io/hashicorp/example", Fingerprint: "after", ResourceTypes: []model.TypeSchemaFingerprint{{Type: "example_item", Version: 1, Fingerprint: "v1"}}, DataSourceTypes: []model.TypeSchemaFingerprint{}}}}
	result := ProviderSchemas(baseline, candidate)
	if len(result) != 1 || result[0].Change != "changed" || len(result[0].ChangedResourceTypes) != 1 || result[0].ChangedResourceTypes[0] != "example_item" {
		t.Fatalf("provider schema change was not detected: %#v", result)
	}
}
