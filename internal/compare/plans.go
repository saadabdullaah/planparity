package compare

import (
	"reflect"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Plans compares normalized action sets from two engines or versions.
func Plans(baseline, candidate model.PlanResult) model.PlanComparison {
	result := model.PlanComparison{
		BaselineSummary: baseline.Summary, CandidateSummary: candidate.Summary,
		OnlyBaseline: []model.PlanChange{}, OnlyCandidate: []model.PlanChange{}, Changed: []model.PlanChange{},
	}
	baselineMap := planChangeMap(baseline.Changes)
	candidateMap := planChangeMap(candidate.Changes)
	for _, change := range baseline.Changes {
		other, exists := candidateMap[change.Address]
		if !exists {
			result.OnlyBaseline = append(result.OnlyBaseline, change)
			continue
		}
		if !planChangesEquivalent(change, other) {
			result.Changed = append(result.Changed, other)
		}
	}
	for _, change := range candidate.Changes {
		if _, exists := baselineMap[change.Address]; !exists {
			result.OnlyCandidate = append(result.OnlyCandidate, change)
		}
	}
	sort.Slice(result.OnlyBaseline, func(i, j int) bool { return result.OnlyBaseline[i].Address < result.OnlyBaseline[j].Address })
	sort.Slice(result.OnlyCandidate, func(i, j int) bool { return result.OnlyCandidate[i].Address < result.OnlyCandidate[j].Address })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].Address < result.Changed[j].Address })
	result.Equivalent = len(result.OnlyBaseline) == 0 && len(result.OnlyCandidate) == 0 && len(result.Changed) == 0 && reflect.DeepEqual(result.BaselineSummary, result.CandidateSummary)
	return result
}

func planChangeMap(changes []model.PlanChange) map[string]model.PlanChange {
	result := make(map[string]model.PlanChange, len(changes))
	for _, change := range changes {
		result[change.Address] = change
	}
	return result
}

func planChangesEquivalent(first, second model.PlanChange) bool {
	return first.ResourceType == second.ResourceType && first.Provider == second.Provider && first.Replacement == second.Replacement &&
		reflect.DeepEqual(first.Actions, second.Actions) && reflect.DeepEqual(first.ChangedPaths, second.ChangedPaths) && reflect.DeepEqual(first.UnknownPaths, second.UnknownPaths)
}

// ProviderSchemas compares provider and type fingerprints.
func ProviderSchemas(baseline, candidate model.ProviderSchemaSet) []model.ProviderSchemaDifference {
	left := providerSchemaMap(baseline.Providers)
	right := providerSchemaMap(candidate.Providers)
	addressSet := make(map[string]struct{}, len(left)+len(right))
	for address := range left {
		addressSet[address] = struct{}{}
	}
	for address := range right {
		addressSet[address] = struct{}{}
	}
	addresses := make([]string, 0, len(addressSet))
	for address := range addressSet {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	result := make([]model.ProviderSchemaDifference, 0)
	for _, address := range addresses {
		before, beforeExists := left[address]
		after, afterExists := right[address]
		difference := model.ProviderSchemaDifference{Address: address, ChangedResourceTypes: []string{}, ChangedDataSourceTypes: []string{}}
		switch {
		case !beforeExists:
			difference.Change = "added"
		case !afterExists:
			difference.Change = "removed"
		case before.Fingerprint != after.Fingerprint:
			difference.Change = "changed"
		}
		if beforeExists && afterExists {
			difference.ChangedResourceTypes = changedSchemaTypes(before.ResourceTypes, after.ResourceTypes)
			difference.ChangedDataSourceTypes = changedSchemaTypes(before.DataSourceTypes, after.DataSourceTypes)
			if difference.Change == "" && (len(difference.ChangedResourceTypes) > 0 || len(difference.ChangedDataSourceTypes) > 0) {
				difference.Change = "changed"
			}
		}
		if difference.Change != "" {
			result = append(result, difference)
		}
	}
	return result
}

func providerSchemaMap(providers []model.ProviderSchemaFingerprint) map[string]model.ProviderSchemaFingerprint {
	result := make(map[string]model.ProviderSchemaFingerprint, len(providers))
	for _, provider := range providers {
		result[provider.Address] = provider
	}
	return result
}

func changedSchemaTypes(before, after []model.TypeSchemaFingerprint) []string {
	left := make(map[string]model.TypeSchemaFingerprint, len(before))
	right := make(map[string]model.TypeSchemaFingerprint, len(after))
	keys := make(map[string]struct{}, len(before)+len(after))
	for _, schema := range before {
		left[schema.Type] = schema
		keys[schema.Type] = struct{}{}
	}
	for _, schema := range after {
		right[schema.Type] = schema
		keys[schema.Type] = struct{}{}
	}
	result := make([]string, 0)
	for key := range keys {
		if !reflect.DeepEqual(left[key], right[key]) {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
