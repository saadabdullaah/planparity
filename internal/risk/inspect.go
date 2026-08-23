// Package risk converts normalized observations into review findings.
package risk

import (
	"fmt"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Inspect classifies unusual resource lifecycle metadata.
func Inspect(snapshot model.StateSnapshot) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, resource := range snapshot.Resources {
		switch resource.Status {
		case model.ResourceTainted:
			findings = append(findings, model.Finding{
				ID: "PP-COMP-003", Category: model.CategoryStateIntegrity, Severity: model.SeverityWarning,
				Verification: model.VerificationVerified, Confidence: 1, Address: resource.Address,
				Title: "Tainted resource instance", Explanation: "The engine exposed this resource instance as tainted.",
				Evidence:       []model.Evidence{{Kind: "state_status", Message: "status=TAINTED"}},
				Recommendation: "Review why the instance is tainted before changing engine or provider versions.",
			})
		case model.ResourceDeposed:
			findings = append(findings, model.Finding{
				ID: "PP-COMP-004", Category: model.CategoryStateIntegrity, Severity: model.SeverityWarning,
				Verification: model.VerificationVerified, Confidence: 1, Address: resource.Address,
				Title: "Deposed resource instance", Explanation: "The engine exposed a deposed resource instance retained during replacement.",
				Evidence:       []model.Evidence{{Kind: "state_status", Message: "status=DEPOSED"}},
				Recommendation: "Resolve the incomplete replacement before undertaking a compatibility transition.",
			})
		case model.ResourceUnknown:
			findings = append(findings, model.Finding{
				ID: "PP-COMP-005", Category: model.CategoryStateIntegrity, Severity: model.SeverityUnknown,
				Verification: model.VerificationUnverified, Confidence: 0, Address: resource.Address,
				Title: "Unknown resource status", Explanation: "The engine exposed lifecycle metadata PlanParity could not classify.",
				Evidence:       []model.Evidence{{Kind: "state_status", Message: "status=UNKNOWN"}},
				Recommendation: "Inspect this instance with the originating engine before proceeding.",
			})
		}
		if resource.ProviderAddress == "" {
			findings = append(findings, model.Finding{
				ID: "PP-PROV-002", Category: model.CategoryProviderSchema, Severity: model.SeverityWarning,
				Verification: model.VerificationVerified, Confidence: 1, Address: resource.Address,
				Title: "Provider address is missing", Explanation: fmt.Sprintf("%s has no provider address in normalized state.", resource.Address),
				Evidence:       []model.Evidence{{Kind: "provider_address", Message: "provider address is empty"}},
				Recommendation: "Verify that the state was decoded by a supported engine version.",
			})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Address == findings[j].Address {
			return findings[i].ID < findings[j].ID
		}
		return findings[i].Address < findings[j].Address
	})
	return findings
}
