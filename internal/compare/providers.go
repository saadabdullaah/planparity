package compare

import (
	"reflect"
	"sort"

	"github.com/saadabdullaah/planparity/internal/model"
)

func compareProviders(before, after []model.ProviderReference) []model.ProviderChange {
	beforeMap := make(map[string]model.ProviderReference, len(before))
	afterMap := make(map[string]model.ProviderReference, len(after))
	addresses := make(map[string]struct{}, len(before)+len(after))
	for _, provider := range before {
		beforeMap[provider.Address] = provider
		addresses[provider.Address] = struct{}{}
	}
	for _, provider := range after {
		afterMap[provider.Address] = provider
		addresses[provider.Address] = struct{}{}
	}
	result := make([]model.ProviderChange, 0)
	for address := range addresses {
		beforeProvider, existedBefore := beforeMap[address]
		afterProvider, existsAfter := afterMap[address]
		change := ""
		switch {
		case !existedBefore:
			change = "added"
		case !existsAfter:
			change = "removed"
		case !reflect.DeepEqual(beforeProvider.SchemaVersions, afterProvider.SchemaVersions):
			change = "schema versions changed"
		}
		if change != "" {
			result = append(result, model.ProviderChange{Address: address, BeforeSchemaVersions: emptyUint(beforeProvider.SchemaVersions), AfterSchemaVersions: emptyUint(afterProvider.SchemaVersions), Change: change})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result
}

func emptyUint(values []uint64) []uint64 {
	if values == nil {
		return []uint64{}
	}
	return values
}
