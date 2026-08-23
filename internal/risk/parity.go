package risk

import (
	"fmt"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Parity classifies plan and provider-schema differences between baseline and candidate.
func Parity(baseline, candidate model.PlanResult, comparison model.PlanComparison, schemas []model.ProviderSchemaDifference) []model.Finding {
	findings := make([]model.Finding, 0)
	if !comparison.Equivalent {
		findings = append(findings, model.Finding{
			ID: "PP-ENGN-001", Category: model.CategoryEngineDifference, Severity: model.SeverityWarning,
			Verification: model.VerificationVerified, Confidence: 1,
			Title: "Plan behavior differs", Explanation: "The candidate engine or version produced different normalized plan behavior from the baseline.",
			Evidence:       []model.Evidence{{Kind: "plan_difference", Message: fmt.Sprintf("only_baseline=%d only_candidate=%d changed=%d", len(comparison.OnlyBaseline), len(comparison.OnlyCandidate), len(comparison.Changed))}},
			Recommendation: "Review every candidate-only and changed action before migration or upgrade.",
		})
	}
	baselineMap := make(map[string]model.PlanChange, len(baseline.Changes))
	for _, change := range baseline.Changes {
		baselineMap[change.Address] = change
	}
	for _, change := range candidate.Changes {
		previous, existed := baselineMap[change.Address]
		if change.Replacement && (!existed || !previous.Replacement) {
			findings = append(findings, model.Finding{
				ID: "PP-DEST-002", Category: model.CategoryDestructiveChange, Severity: model.SeverityHigh,
				Verification: model.VerificationVerified, Confidence: 1, Address: change.Address,
				Title: "New candidate replacement", Explanation: fmt.Sprintf("%s is replaced only by the candidate behavior.", change.Address),
				Evidence:       []model.Evidence{{Kind: "candidate_plan", Message: "new action=REPLACE"}},
				Recommendation: "Do not proceed until the replacement is understood and accepted.",
			})
		}
		if hasAction(change.Actions, model.PlanDelete) && (!existed || !hasAction(previous.Actions, model.PlanDelete)) {
			findings = append(findings, model.Finding{
				ID: "PP-DEST-001", Category: model.CategoryDestructiveChange, Severity: model.SeverityHigh,
				Verification: model.VerificationVerified, Confidence: 1, Address: change.Address,
				Title: "New candidate deletion", Explanation: fmt.Sprintf("%s is deleted only by the candidate behavior.", change.Address),
				Evidence:       []model.Evidence{{Kind: "candidate_plan", Message: "new action=DELETE"}},
				Recommendation: "Do not proceed until the deletion is understood and accepted.",
			})
		}
	}
	for _, schema := range schemas {
		findings = append(findings, model.Finding{
			ID: "PP-PROV-001", Category: model.CategoryProviderSchema, Severity: model.SeverityWarning,
			Verification: model.VerificationVerified, Confidence: 1, Address: schema.Address,
			Title: "Provider schema differs", Explanation: fmt.Sprintf("Provider %s schema was %s for the candidate.", schema.Address, schema.Change),
			Evidence:       []model.Evidence{{Kind: "provider_schema", Message: fmt.Sprintf("resource_types=%d data_sources=%d", len(schema.ChangedResourceTypes), len(schema.ChangedDataSourceTypes))}},
			Recommendation: "Review provider upgrade and migration notes for the changed types.",
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Address == findings[j].Address {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].Address < findings[j].Address
	})
	return findings
}

func hasAction(actions []model.PlanAction, target model.PlanAction) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}
