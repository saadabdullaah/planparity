package normalize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

const secretCanary = "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9"

func fixturePath(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve fixture path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(append([]string{root, "testdata"}, parts...)...)
}

func TestStateNormalizesWithoutValues(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(fixturePath(t, "terraform", "show-basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := State(raw, model.EngineTerraform)
	if err != nil {
		t.Fatalf("normalize state: %v", err)
	}
	if got, want := len(snapshot.Resources), 3; got != want {
		t.Fatalf("resources=%d want=%d", got, want)
	}
	if got, want := len(snapshot.Modules), 2; got != want {
		t.Fatalf("modules=%d want=%d", got, want)
	}
	if got, want := len(snapshot.Providers), 3; got != want {
		t.Fatalf("providers=%d want=%d", got, want)
	}
	if snapshot.Resources[0].Address != "local_file.inventory" {
		t.Fatalf("resources are not sorted: %#v", snapshot.Resources)
	}
	if len(snapshot.Resources[0].Structure.SensitivePaths) != 1 || snapshot.Resources[0].Structure.SensitivePaths[0] != "content" {
		t.Fatalf("unexpected sensitive paths: %#v", snapshot.Resources[0].Structure.SensitivePaths)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secretCanary) {
		t.Fatal("secret canary escaped normalization")
	}
	if !strings.Contains(string(encoded), `"dependencies":[]`) {
		t.Fatalf("empty dependency lists must remain arrays: %s", encoded)
	}
	for _, forbidden := range []string{`"value"`, `"values"`} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("forbidden field %s escaped normalization", forbidden)
		}
	}
}

func TestStateIsDeterministic(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(fixturePath(t, "terraform", "show-basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := State(raw, model.EngineTerraform)
	if err != nil {
		t.Fatal(err)
	}
	second, err := State(raw, model.EngineTerraform)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("identical input produced different normalized JSON")
	}
}

func TestStateRejectsMalformedAndMissingValues(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		path []string
		code ErrorCode
	}{
		{name: "malformed", path: []string{"corrupted", "not-json.json"}, code: ErrorInvalidJSON},
		{name: "missing", path: []string{"corrupted", "missing-values.json"}, code: ErrorMissingValues},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, err := os.ReadFile(fixturePath(t, test.path...))
			if err != nil {
				t.Fatal(err)
			}
			_, err = State(raw, model.EngineTerraform)
			normalizeError, ok := err.(*Error)
			if !ok || normalizeError.Code != test.code {
				t.Fatalf("error=%v want code=%s", err, test.code)
			}
		})
	}
}

func FuzzStateNeverPanics(f *testing.F) {
	f.Add([]byte(`{"format_version":"1.0","values":{"root_module":{"resources":[]}}}`))
	f.Add([]byte(`not-json`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		_, _ = State(raw, model.EngineTerraform)
	})
}

func BenchmarkStateTenThousand(b *testing.B) {
	resources := make([]rawResource, 10_000)
	for index := range resources {
		name := fmt.Sprintf("item_%05d", index)
		resources[index] = rawResource{
			Address:       "terraform_data." + name,
			Mode:          "managed",
			Type:          "terraform_data",
			Name:          name,
			ProviderName:  "terraform.io/builtin/terraform",
			SchemaVersion: 0,
			Values:        json.RawMessage(`{"id":"fixture","input":{"enabled":true}}`),
		}
	}
	raw, err := json.Marshal(rawStateShow{
		FormatVersion:    "1.0",
		TerraformVersion: "1.12.0",
		Values:           &rawValues{RootModule: &rawModule{Resources: resources}},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(raw)))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		snapshot, err := State(raw, model.EngineTerraform)
		if err != nil {
			b.Fatal(err)
		}
		if len(snapshot.Resources) != len(resources) {
			b.Fatalf("resources=%d want=%d", len(snapshot.Resources), len(resources))
		}
	}
}
