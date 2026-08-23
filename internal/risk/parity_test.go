package risk

import (
	"testing"

	"github.com/saadabdullaah/planparity/internal/compare"
	"github.com/saadabdullaah/planparity/internal/model"
)

func TestParityFindingsCoverMismatchDestructionAndSchema(t *testing.T) {
	baseline := model.PlanResult{Changes: []model.PlanChange{{Address: "terraform_data.delete", Actions: []model.PlanAction{model.PlanNoOp}}, {Address: "terraform_data.replace", Actions: []model.PlanAction{model.PlanNoOp}}}, Summary: model.PlanSummary{NoOps: 2}}
	candidate := model.PlanResult{Changes: []model.PlanChange{{Address: "terraform_data.delete", Actions: []model.PlanAction{model.PlanDelete}}, {Address: "terraform_data.replace", Actions: []model.PlanAction{model.PlanReplace}, Replacement: true}}, Summary: model.PlanSummary{Deletes: 1, Replacements: 1}}
	comparison := compare.Plans(baseline, candidate)
	schemas := []model.ProviderSchemaDifference{{Address: "registry.terraform.io/hashicorp/example", Change: "changed", ChangedResourceTypes: []string{"example_item"}, ChangedDataSourceTypes: []string{}}}
	findings := Parity(baseline, candidate, comparison, schemas)
	wanted := map[string]bool{"PP-ENGN-001": false, "PP-DEST-001": false, "PP-DEST-002": false, "PP-PROV-001": false}
	for _, finding := range findings {
		if _, exists := wanted[finding.ID]; exists {
			wanted[finding.ID] = true
		}
	}
	for id, found := range wanted {
		if !found {
			t.Errorf("missing finding %s: %#v", id, findings)
		}
	}
}

func TestPlanAlwaysStatesProviderIsolationIsUnverified(t *testing.T) {
	findings := Plan(model.PlanResult{})
	if len(findings) != 1 || findings[0].ID != "PP-SAFE-001" || findings[0].Verification != model.VerificationUnverified {
		t.Fatalf("safety limitation missing: %#v", findings)
	}
}
