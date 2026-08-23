package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStateSnapshotJSONHasNoValueField(t *testing.T) {
	t.Parallel()
	snapshot := StateSnapshot{
		SchemaVersion: SchemaV1,
		FormatVersion: "1.0",
		Resources: []ResourceInstance{{
			Address:       "local_file.example",
			Mode:          ResourceManaged,
			ResourceType:  "local_file",
			ResourceName:  "example",
			Status:        ResourceNormal,
			Dependencies:  []string{},
			Structure:     AttributeStructure{Keys: []string{"content"}},
			SchemaVersion: 0,
		}},
		Providers: []ProviderReference{},
		Modules:   []ModuleReference{},
		Outputs:   []OutputMetadata{},
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"value"`, `"values"`, `"before"`, `"after"`} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("public state contains forbidden field %s: %s", forbidden, encoded)
		}
	}
}
