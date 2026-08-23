package compare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/normalize"
)

func comparisonFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, "testdata", "comparisons", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestStatesDetectsProbableMove(t *testing.T) {
	t.Parallel()
	before, err := normalize.State(comparisonFixture(t, "before.json"), model.EngineTerraform)
	if err != nil {
		t.Fatal(err)
	}
	after, err := normalize.State(comparisonFixture(t, "after.json"), model.EngineTerraform)
	if err != nil {
		t.Fatal(err)
	}
	result := States(before, after, Options{ProbableMoveThreshold: 0.85})
	if got, want := len(result.ProbableMoves), 1; got != want {
		t.Fatalf("moves=%d want=%d: %#v", got, want, result.ProbableMoves)
	}
	move := result.ProbableMoves[0]
	if move.FromAddress != "local_file.application" || move.ToAddress != "module.files.local_file.application" || move.Confidence != 1 {
		t.Fatalf("unexpected move: %#v", move)
	}
	if len(result.Removed) != 1 || result.Removed[0].Address != "terraform_data.removed" {
		t.Fatalf("unexpected removed: %#v", result.Removed)
	}
	if len(result.Added) != 1 || result.Added[0].Address != "terraform_data.added" {
		t.Fatalf("unexpected added: %#v", result.Added)
	}
}

func TestStatesIsDeterministic(t *testing.T) {
	t.Parallel()
	before, _ := normalize.State(comparisonFixture(t, "before.json"), model.EngineTerraform)
	after, _ := normalize.State(comparisonFixture(t, "after.json"), model.EngineTerraform)
	first, _ := json.Marshal(States(before, after, Options{}))
	second, _ := json.Marshal(States(before, after, Options{}))
	if string(first) != string(second) {
		t.Fatal("comparison output is nondeterministic")
	}
}

func TestIdenticalStatesHaveNoDifferences(t *testing.T) {
	t.Parallel()
	snapshot, err := normalize.State(comparisonFixture(t, "before.json"), model.EngineTerraform)
	if err != nil {
		t.Fatal(err)
	}
	result := States(snapshot, snapshot, Options{})
	if result.Summary.Added != 0 || result.Summary.Removed != 0 || result.Summary.Changed != 0 || result.Summary.ProbableMoves != 0 {
		t.Fatalf("identical states differ: %#v", result.Summary)
	}
}

func BenchmarkStatesTenThousand(b *testing.B) {
	resources := make([]model.ResourceInstance, 10_000)
	for index := range resources {
		resources[index] = model.ResourceInstance{Address: fmt.Sprintf("terraform_data.item[%d]", index), ResourceType: "terraform_data", ResourceName: "item", ProviderAddress: "terraform.io/builtin/terraform", Structure: model.AttributeStructure{TypeFingerprint: "shape"}, Dependencies: []string{}}
	}
	snapshot := model.StateSnapshot{Resources: resources, Providers: []model.ProviderReference{}}
	b.ResetTimer()
	for range b.N {
		_ = States(snapshot, snapshot, Options{})
	}
}
