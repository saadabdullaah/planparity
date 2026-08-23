package normalize

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func TestPlanNormalizesReplacementWithoutValues(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(fixturePath(t, "plans", "replacement.json"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := Plan(raw, model.EngineTerraform)
	if err != nil {
		t.Fatalf("normalize plan: %v", err)
	}
	if result.Summary.Replacements != 1 || result.Summary.Creates != 1 {
		t.Fatalf("unexpected summary: %#v", result.Summary)
	}
	if !result.Changes[0].Replacement || result.Changes[0].Actions[0] != model.PlanReplace {
		t.Fatalf("replacement not classified: %#v", result.Changes[0])
	}
	if !contains(result.Changes[0].ChangedPaths, "filename") {
		t.Fatalf("changed path missing: %#v", result.Changes[0].ChangedPaths)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secretCanary) {
		t.Fatal("secret canary escaped plan normalization")
	}
	for _, forbidden := range []string{`"before"`, `"after"`, `"value"`, `"values"`} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("forbidden field %s escaped normalization", forbidden)
		}
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func FuzzPlanNeverPanics(f *testing.F) {
	f.Add([]byte(`{"format_version":"1.0","resource_changes":[]}`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		_, _ = Plan(raw, model.EngineOpenTofu)
	})
}
