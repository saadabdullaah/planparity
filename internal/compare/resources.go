// Package compare performs deterministic comparisons of normalized models.
package compare

import (
	"fmt"
	"math"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Options configures conservative inference behavior.
type Options struct {
	ProbableMoveThreshold float64
}

// States compares two sanitized snapshots.
func States(before, after model.StateSnapshot, options Options) model.StateComparison {
	threshold := options.ProbableMoveThreshold
	if threshold <= 0 {
		threshold = 0.85
	}
	result := model.StateComparison{
		SchemaVersion: model.SchemaV1,
		BeforeHash:    before.SourceHash,
		AfterHash:     after.SourceHash,
		Added:         []model.ResourceInstance{}, Removed: []model.ResourceInstance{}, Changed: []model.ResourceChange{},
		ProbableMoves: []model.ProbableMove{}, ProviderChanges: []model.ProviderChange{}, Findings: []model.Finding{},
	}

	beforeByAddress := resourceMap(before.Resources)
	afterByAddress := resourceMap(after.Resources)
	unmatchedBefore := make([]model.ResourceInstance, 0)
	unmatchedAfter := make([]model.ResourceInstance, 0)
	for _, resource := range before.Resources {
		candidate, exists := afterByAddress[resource.Address]
		if !exists {
			unmatchedBefore = append(unmatchedBefore, resource)
			continue
		}
		change, changed := compareMatchedResource(resource, candidate)
		if changed {
			result.Changed = append(result.Changed, change)
		} else {
			result.Summary.Unchanged++
		}
	}
	for _, resource := range after.Resources {
		if _, exists := beforeByAddress[resource.Address]; !exists {
			unmatchedAfter = append(unmatchedAfter, resource)
		}
	}

	moves, movedBefore, movedAfter := inferMoves(unmatchedBefore, unmatchedAfter, threshold)
	result.ProbableMoves = moves
	for _, resource := range unmatchedBefore {
		if !movedBefore[resource.Address] {
			result.Removed = append(result.Removed, resource)
		}
	}
	for _, resource := range unmatchedAfter {
		if !movedAfter[resource.Address] {
			result.Added = append(result.Added, resource)
		}
	}
	result.ProviderChanges = compareProviders(before.Providers, after.Providers)
	sortComparison(&result)
	result.Summary.Added = len(result.Added)
	result.Summary.Removed = len(result.Removed)
	result.Summary.Changed = len(result.Changed)
	result.Summary.ProbableMoves = len(result.ProbableMoves)
	result.Summary.ProviderChanges = len(result.ProviderChanges)
	result.Findings = comparisonFindings(result)
	return result
}

func resourceMap(resources []model.ResourceInstance) map[string]model.ResourceInstance {
	result := make(map[string]model.ResourceInstance, len(resources))
	for _, resource := range resources {
		result[resource.Address] = resource
	}
	return result
}

func compareMatchedResource(before, after model.ResourceInstance) (model.ResourceChange, bool) {
	change := model.ResourceChange{Address: before.Address, Before: before, After: after, ChangedFields: []string{}}
	if before.ProviderAddress != after.ProviderAddress {
		change.ProviderChanged = true
		change.ChangedFields = append(change.ChangedFields, "provider")
	}
	if before.SchemaVersion != after.SchemaVersion {
		change.SchemaChanged = true
		change.ChangedFields = append(change.ChangedFields, "schema_version")
	}
	if before.Structure.TypeFingerprint != after.Structure.TypeFingerprint || !equalStrings(before.Structure.Keys, after.Structure.Keys) {
		change.StructureChanged = true
		change.ChangedFields = append(change.ChangedFields, "structure")
	}
	if !equalStrings(before.Dependencies, after.Dependencies) {
		change.DependenciesChanged = true
		change.ChangedFields = append(change.ChangedFields, "dependencies")
	}
	if before.Status != after.Status {
		change.ChangedFields = append(change.ChangedFields, "status")
	}
	return change, len(change.ChangedFields) > 0
}

func equalStrings(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

type moveCandidate struct {
	from    model.ResourceInstance
	to      model.ResourceInstance
	score   float64
	signals []model.MoveSignal
}

func inferMoves(removed, added []model.ResourceInstance, threshold float64) ([]model.ProbableMove, map[string]bool, map[string]bool) {
	candidatesBySource := make(map[string][]moveCandidate)
	for _, from := range removed {
		for _, to := range added {
			score, signals := moveScore(from, to)
			if score+1e-9 >= threshold {
				candidatesBySource[from.Address] = append(candidatesBySource[from.Address], moveCandidate{from: from, to: to, score: score, signals: signals})
			}
		}
	}
	sources := make([]string, 0, len(candidatesBySource))
	for source := range candidatesBySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	moves := make([]model.ProbableMove, 0)
	usedBefore := make(map[string]bool)
	usedAfter := make(map[string]bool)
	for _, source := range sources {
		candidates := candidatesBySource[source]
		sort.Slice(candidates, func(i, j int) bool {
			if math.Abs(candidates[i].score-candidates[j].score) < 1e-9 {
				return candidates[i].to.Address < candidates[j].to.Address
			}
			return candidates[i].score > candidates[j].score
		})
		best := candidates[0]
		alternatives := make([]string, 0)
		for _, candidate := range candidates[1:] {
			if math.Abs(candidate.score-best.score) < 0.0001 {
				alternatives = append(alternatives, candidate.to.Address)
			}
		}
		ambiguous := len(alternatives) > 0 || usedAfter[best.to.Address]
		move := model.ProbableMove{FromAddress: best.from.Address, ToAddress: best.to.Address, From: best.from, To: best.to, Confidence: roundConfidence(best.score), Signals: best.signals, Ambiguous: ambiguous, Alternatives: alternatives}
		moves = append(moves, move)
		if !ambiguous {
			usedBefore[best.from.Address] = true
			usedAfter[best.to.Address] = true
		}
	}
	return moves, usedBefore, usedAfter
}

func moveScore(from, to model.ResourceInstance) (float64, []model.MoveSignal) {
	dependencyMatch := dependencySimilarity(from.Dependencies, to.Dependencies) >= 0.8
	signals := []model.MoveSignal{
		{Name: "resource_type", Weight: 0.25, Match: from.ResourceType == to.ResourceType},
		{Name: "provider", Weight: 0.20, Match: from.ProviderAddress == to.ProviderAddress},
		{Name: "structure", Weight: 0.25, Match: from.Structure.TypeFingerprint == to.Structure.TypeFingerprint},
		{Name: "mode", Weight: 0.05, Match: from.Mode == to.Mode},
		{Name: "resource_name", Weight: 0.20, Match: from.ResourceName == to.ResourceName},
		{Name: "dependencies", Weight: 0.05, Match: dependencyMatch},
	}
	score := 0.0
	for _, signal := range signals {
		if signal.Match {
			score += signal.Weight
		}
	}
	return score, signals
}

func dependencySimilarity(first, second []string) float64 {
	if len(first) == 0 && len(second) == 0 {
		return 1
	}
	left := make(map[string]struct{}, len(first))
	for _, value := range first {
		left[value] = struct{}{}
	}
	intersection := 0
	union := len(left)
	for _, value := range second {
		if _, exists := left[value]; exists {
			intersection++
		} else {
			union++
		}
	}
	if union == 0 {
		return 1
	}
	return float64(intersection) / float64(union)
}

func roundConfidence(value float64) float64 { return math.Round(value*1000) / 1000 }

func sortComparison(result *model.StateComparison) {
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].Address < result.Added[j].Address })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].Address < result.Removed[j].Address })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].Address < result.Changed[j].Address })
	sort.Slice(result.ProbableMoves, func(i, j int) bool {
		if result.ProbableMoves[i].FromAddress == result.ProbableMoves[j].FromAddress {
			return result.ProbableMoves[i].ToAddress < result.ProbableMoves[j].ToAddress
		}
		return result.ProbableMoves[i].FromAddress < result.ProbableMoves[j].FromAddress
	})
	sort.Slice(result.ProviderChanges, func(i, j int) bool { return result.ProviderChanges[i].Address < result.ProviderChanges[j].Address })
}

func comparisonFindings(comparison model.StateComparison) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, resource := range comparison.Removed {
		findings = append(findings, model.Finding{
			ID: "PP-DEST-001", Category: model.CategoryDestructiveChange, Severity: model.SeverityHigh,
			Verification: model.VerificationVerified, Confidence: 1, Address: resource.Address,
			Title: "Resource removed from state", Explanation: fmt.Sprintf("%s exists only in the earlier snapshot.", resource.Address),
			Evidence:       []model.Evidence{{Kind: "state_presence", Message: "present before and absent after"}},
			Recommendation: "Confirm whether the resource was intentionally removed or its identity changed.",
		})
	}
	for _, move := range comparison.ProbableMoves {
		verification := model.VerificationInferred
		severity := model.SeverityInfo
		title := "Probable resource move"
		recommendation := "Confirm the move with a moved block or independent configuration review."
		if move.Ambiguous {
			severity = model.SeverityWarning
			title = "Ambiguous resource identity"
			recommendation = "Resolve the competing identity candidates before relying on this inference."
		}
		findings = append(findings, model.Finding{
			ID: "PP-IDEN-002", Category: model.CategoryResourceIdentity, Severity: severity,
			Verification: verification, Confidence: move.Confidence, Address: move.FromAddress,
			Title: title, Explanation: fmt.Sprintf("%s may have moved to %s.", move.FromAddress, move.ToAddress),
			Evidence: moveEvidence(move.Signals), Recommendation: recommendation,
		})
	}
	for _, provider := range comparison.ProviderChanges {
		findings = append(findings, model.Finding{
			ID: "PP-PROV-001", Category: model.CategoryProviderSchema, Severity: model.SeverityWarning,
			Verification: model.VerificationVerified, Confidence: 1, Address: provider.Address,
			Title: "Provider schema usage changed", Explanation: fmt.Sprintf("Provider %s was %s between snapshots.", provider.Address, provider.Change),
			Evidence:       []model.Evidence{{Kind: "provider_change", Message: provider.Change}},
			Recommendation: "Review provider requirements and schema migration notes.",
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

func moveEvidence(signals []model.MoveSignal) []model.Evidence {
	evidence := make([]model.Evidence, 0, len(signals))
	for _, signal := range signals {
		if signal.Match {
			evidence = append(evidence, model.Evidence{Kind: signal.Name, Message: fmt.Sprintf("matched (weight %.2f)", signal.Weight)})
		}
	}
	return evidence
}
