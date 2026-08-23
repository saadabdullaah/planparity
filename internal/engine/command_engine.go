package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/runner"
)

// CommandEngine implements common CLI behavior while retaining an explicit kind.
type CommandEngine struct {
	kind     model.EngineKind
	name     string
	binary   string
	executor runner.Executor
}

// NewCommandEngine constructs the shared command-backed adapter implementation.
func NewCommandEngine(kind model.EngineKind, name, binary string, executor runner.Executor) *CommandEngine {
	return &CommandEngine{kind: kind, name: name, binary: binary, executor: executor}
}

func (adapter *CommandEngine) Name() string           { return adapter.name }
func (adapter *CommandEngine) Kind() model.EngineKind { return adapter.kind }
func (adapter *CommandEngine) BinaryPath() string     { return adapter.binary }

func (adapter *CommandEngine) Version(ctx context.Context) (model.EngineMetadata, error) {
	result, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: filepath.Dir(adapter.binary), Operation: runner.OperationVersion,
		Environment: runner.SafeEnvironment(runner.EnvironmentOptions{}),
	})
	if err != nil {
		return model.EngineMetadata{}, executionFailure("version", err)
	}
	var response struct {
		TerraformVersion string `json:"terraform_version"`
		Platform         string `json:"platform"`
	}
	if err := json.Unmarshal(result.Stdout, &response); err != nil || response.TerraformVersion == "" {
		return model.EngineMetadata{}, &Failure{Code: "INVALID_ENGINE_OUTPUT", Operation: "version", Message: "engine returned invalid version JSON", Cause: err}
	}
	return model.EngineMetadata{Kind: adapter.kind, Name: adapter.name, Version: response.TerraformVersion, Platform: response.Platform}, nil
}

func (adapter *CommandEngine) Capabilities(ctx context.Context) (model.Capabilities, error) {
	if _, err := adapter.Version(ctx); err != nil {
		return model.Capabilities{}, err
	}
	return model.Capabilities{MachineReadableState: true, MachineReadablePlan: true, ProviderSchema: true, ValidateJSON: true, RefreshDisabledPlan: true}, nil
}

func (adapter *CommandEngine) Validate(ctx context.Context, workspace Workspace) (ValidationResult, error) {
	environment := workspace.Environment
	if len(environment) == 0 {
		environment = runner.SafeEnvironment(runner.EnvironmentOptions{})
	}
	result, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: workspace.Path, Operation: runner.OperationValidate,
		Environment: environment,
	})
	if err != nil {
		return ValidationResult{}, executionFailure("validate", err)
	}
	var response struct {
		Valid       bool                                         `json:"valid"`
		Diagnostics []struct{ Severity, Summary, Detail string } `json:"diagnostics"`
	}
	if err := json.Unmarshal(result.Stdout, &response); err != nil {
		return ValidationResult{}, &Failure{Code: "INVALID_ENGINE_OUTPUT", Operation: "validate", Message: "engine returned invalid validation JSON", Cause: err}
	}
	diagnostics := make([]model.Diagnostic, 0, len(response.Diagnostics))
	for _, diagnostic := range response.Diagnostics {
		diagnostics = append(diagnostics, model.Diagnostic{Severity: sanitizeMessage(diagnostic.Severity, 32), Summary: sanitizeMessage(diagnostic.Summary, 512), Detail: sanitizeMessage(diagnostic.Detail, 2048)})
	}
	return ValidationResult{Valid: response.Valid, Diagnostics: diagnostics}, nil
}

func (adapter *CommandEngine) ShowState(ctx context.Context, statePath string) (RawStateResult, error) {
	result, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: filepath.Dir(statePath), Operation: runner.OperationShow, TargetPath: statePath,
		Environment: runner.SafeEnvironment(runner.EnvironmentOptions{}),
	})
	if err != nil {
		return RawStateResult{}, executionFailure("show_state", err)
	}
	metadata, err := metadataFromJSON(result.Stdout, adapter.kind, adapter.name)
	if err != nil {
		metadata, err = adapter.Version(ctx)
		if err != nil {
			return RawStateResult{}, err
		}
	}
	return RawStateResult{JSON: result.Stdout, Engine: metadata, Duration: result.Duration}, nil
}

func (adapter *CommandEngine) ProviderSchemas(ctx context.Context, workspace Workspace) (RawProviderSchemas, error) {
	environment := workspace.Environment
	if len(environment) == 0 {
		environment = runner.SafeEnvironment(runner.EnvironmentOptions{})
	}
	result, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: workspace.Path, Operation: runner.OperationProviderSchemas,
		Environment: environment,
	})
	if err != nil {
		return RawProviderSchemas{}, executionFailure("provider_schemas", err)
	}
	metadata, metadataErr := adapter.Version(ctx)
	if metadataErr != nil {
		return RawProviderSchemas{}, metadataErr
	}
	return RawProviderSchemas{JSON: result.Stdout, Engine: metadata, Duration: result.Duration}, nil
}

func (adapter *CommandEngine) Plan(ctx context.Context, request PlanRequest) (RawPlanResult, error) {
	environment := request.Workspace.Environment
	if len(environment) == 0 {
		environment = runner.SafeEnvironment(runner.EnvironmentOptions{})
	}
	if _, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: request.Workspace.Path, Operation: runner.OperationInit, Environment: environment,
	}); err != nil {
		return RawPlanResult{}, executionFailure("init", err)
	}
	if _, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: request.Workspace.Path, Operation: runner.OperationPlan,
		StatePath: request.StatePath, OutputPath: request.PlanPath, Environment: environment,
	}); err != nil {
		return RawPlanResult{}, executionFailure("plan", err)
	}
	result, err := adapter.executor.Execute(ctx, runner.Request{
		Binary: adapter.binary, WorkingDirectory: request.Workspace.Path, Operation: runner.OperationShow,
		TargetPath: request.PlanPath, Environment: environment,
	})
	if err != nil {
		return RawPlanResult{}, executionFailure("show_plan", err)
	}
	metadata, err := metadataFromJSON(result.Stdout, adapter.kind, adapter.name)
	if err != nil {
		return RawPlanResult{}, err
	}
	return RawPlanResult{JSON: result.Stdout, Engine: metadata, Duration: result.Duration}, nil
}

func metadataFromJSON(raw []byte, kind model.EngineKind, name string) (model.EngineMetadata, error) {
	var response struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal(raw, &response); err != nil || response.TerraformVersion == "" {
		return model.EngineMetadata{}, &Failure{Code: "INVALID_ENGINE_OUTPUT", Operation: "decode", Message: fmt.Sprintf("%s returned JSON without an engine version", name), Cause: err}
	}
	return model.EngineMetadata{Kind: kind, Name: name, Version: response.TerraformVersion}, nil
}

var _ Engine = (*CommandEngine)(nil)

// Compile-time reference prevents accidental removal while duration remains part of raw results.
var _ = time.Duration(0)
