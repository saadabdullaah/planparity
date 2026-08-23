package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	enginepkg "github.com/saadabdullaah/planparity/internal/engine"
	"github.com/saadabdullaah/planparity/internal/model"
)

type fakeResolver struct{ adapter enginepkg.Engine }

func (resolver fakeResolver) Resolve(model.EngineKind, string, time.Duration) (enginepkg.Engine, error) {
	return resolver.adapter, nil
}

type fakeEngine struct {
	raw      []byte
	planRaw  []byte
	lastPlan enginepkg.PlanRequest
}

func (engine *fakeEngine) Name() string           { return "Terraform" }
func (engine *fakeEngine) Kind() model.EngineKind { return model.EngineTerraform }
func (engine *fakeEngine) BinaryPath() string     { return "terraform" }
func (engine *fakeEngine) Version(context.Context) (model.EngineMetadata, error) {
	return model.EngineMetadata{Kind: model.EngineTerraform, Name: "Terraform", Version: "1.12.1"}, nil
}
func (engine *fakeEngine) Capabilities(context.Context) (model.Capabilities, error) {
	return model.Capabilities{}, nil
}
func (engine *fakeEngine) Validate(context.Context, enginepkg.Workspace) (enginepkg.ValidationResult, error) {
	return enginepkg.ValidationResult{Valid: true}, nil
}
func (engine *fakeEngine) ShowState(context.Context, string) (enginepkg.RawStateResult, error) {
	return enginepkg.RawStateResult{JSON: append([]byte(nil), engine.raw...), Engine: model.EngineMetadata{Kind: model.EngineTerraform, Name: "Terraform", Version: "1.12.1"}}, nil
}
func (engine *fakeEngine) ProviderSchemas(context.Context, enginepkg.Workspace) (enginepkg.RawProviderSchemas, error) {
	return enginepkg.RawProviderSchemas{}, nil
}
func (engine *fakeEngine) Plan(_ context.Context, request enginepkg.PlanRequest) (enginepkg.RawPlanResult, error) {
	engine.lastPlan = request
	return enginepkg.RawPlanResult{JSON: append([]byte(nil), engine.planRaw...), Engine: model.EngineMetadata{Kind: model.EngineTerraform, Name: "Terraform", Version: "1.12.1"}}, nil
}

func TestRehearsePlanUsesCopiedWorkspaceAndCleansIt(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	stateJSON, err := os.ReadFile(filepath.Join(root, "testdata", "terraform", "show-basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	planJSON, err := os.ReadFile(filepath.Join(root, "testdata", "plans", "replacement.json"))
	if err != nil {
		t.Fatal(err)
	}
	config := t.TempDir()
	state := filepath.Join(t.TempDir(), "terraform.tfstate")
	if err := os.WriteFile(filepath.Join(config, "main.tf"), []byte("terraform {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state, []byte(`{"synthetic":"state"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	adapter := &fakeEngine{raw: stateJSON, planRaw: planJSON}
	service := NewService(fakeResolver{adapter: adapter}, time.Second)
	result, err := service.RehearsePlan(context.Background(), PlanRehearsalRequest{
		Engine: model.EngineTerraform, ConfigPath: config, StatePath: state, AllowRegistryAccess: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if adapter.lastPlan.Workspace.Path == config || adapter.lastPlan.StatePath == state {
		t.Fatal("engine received original input paths")
	}
	workspaceRoot := filepath.Dir(adapter.lastPlan.Workspace.Path)
	if _, err := os.Stat(workspaceRoot); !os.IsNotExist(err) {
		t.Fatalf("temporary workspace remains: %s", workspaceRoot)
	}
	if result.Candidate == nil || result.Candidate.Summary.Replacements != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9") {
		t.Fatal("canary escaped rehearsal")
	}
	if !strings.Contains(string(encoded), "PP-DEST-002") {
		t.Fatal("replacement finding missing")
	}
}

func TestInspectPreservesInputAndRedactsCanary(t *testing.T) {
	t.Parallel()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, "testdata", "terraform", "show-basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	temporaryState := filepath.Join(t.TempDir(), "terraform.tfstate")
	if err := os.WriteFile(temporaryState, []byte(`{"synthetic":"state"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(temporaryState)
	service := NewService(fakeResolver{adapter: &fakeEngine{raw: raw}}, time.Second)
	result, err := service.Inspect(context.Background(), InspectRequest{Engine: model.EngineTerraform, StatePath: temporaryState})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(temporaryState)
	if string(before) != string(after) {
		t.Fatal("inspection modified supplied state")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9") {
		t.Fatal("canary escaped inspection result")
	}
}
