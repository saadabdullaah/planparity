package risk

import (
	"fmt"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Plan classifies destructive actions and bounded safety limitations.
func Plan(result model.PlanResult) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, change := range result.Changes {
		if change.Replacement {
			findings = append(findings, model.Finding{
				ID: "PP-DEST-002", Category: model.CategoryDestructiveChange, Severity: model.SeverityHigh,
				Verification: model.VerificationVerified, Confidence: 1, Address: change.Address,
				Title: "Resource replacement detected", Explanation: fmt.Sprintf("%s is planned for replacement.", change.Address),
				Evidence:       []model.Evidence{{Kind: "plan_action", Message: "action=REPLACE"}},
				Recommendation: "Review replacement paths and operational impact before proceeding.",
			})
			continue
		}
		for _, action := range change.Actions {
			if action == model.PlanDelete {
				findings = append(findings, model.Finding{
					ID: "PP-DEST-001", Category: model.CategoryDestructiveChange, Severity: model.SeverityHigh,
					Verification: model.VerificationVerified, Confidence: 1, Address: change.Address,
					Title: "Resource deletion detected", Explanation: fmt.Sprintf("%s is planned for deletion.", change.Address),
					Evidence:       []model.Evidence{{Kind: "plan_action", Message: "action=DELETE"}},
					Recommendation: "Confirm the deletion is intended and recoverable.",
				})
			}
		}
	}
	findings = append(findings, model.Finding{
		ID: "PP-SAFE-001", Category: model.CategorySafetyLimitation, Severity: model.SeverityWarning,
		Verification: model.VerificationUnverified, Confidence: 0,
		Title: "Provider behavior is not fully isolated", Explanation: "Refresh was disabled and known credentials were removed, but providers may still attempt network access or execute local code.",
		Evidence:       []model.Evidence{{Kind: "execution_mode", Message: model.LocalRestricted}},
		Recommendation: "Treat real cloud API behavior as unverified and review the documented security boundary.",
	})
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Address == findings[j].Address {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].Address < findings[j].Address
	})
	return findings
}
