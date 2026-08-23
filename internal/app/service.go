// Package app coordinates engine adapters, normalization, comparison, and rehearsals.
package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	comparepkg "github.com/saadabdullaah/planparity/internal/compare"
	"github.com/saadabdullaah/planparity/internal/engine"
	"github.com/saadabdullaah/planparity/internal/engine/opentofu"
	"github.com/saadabdullaah/planparity/internal/engine/terraform"
	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/normalize"
	"github.com/saadabdullaah/planparity/internal/risk"
	"github.com/saadabdullaah/planparity/internal/runner"
)

// Resolver constructs a validated adapter for a requested engine.
type Resolver interface {
	Resolve(model.EngineKind, string, time.Duration) (engine.Engine, error)
}

// DefaultResolver uses local executable discovery and the restricted process executor.
type DefaultResolver struct{}

func (DefaultResolver) Resolve(kind model.EngineKind, binaryPath string, timeout time.Duration) (engine.Engine, error) {
	binary, err := engine.Discover(kind, binaryPath)
	if err != nil {
		return nil, err
	}
	executor := runner.ProcessExecutor{Timeout: timeout}
	if kind == model.EngineTerraform {
		return terraform.New(binary, executor), nil
	}
	if kind == model.EngineOpenTofu {
		return opentofu.New(binary, executor), nil
	}
	return nil, fmt.Errorf("unsupported engine %q", kind)
}

// Service holds reusable application orchestration used by CLI and API.
type Service struct {
	resolver Resolver
	timeout  time.Duration
}

// EngineStatuses detects both supported engines for the local system endpoint.
func (service *Service) EngineStatuses(ctx context.Context, terraformPath, tofuPath string) []model.EngineStatus {
	return []model.EngineStatus{
		service.engineStatus(ctx, model.EngineTerraform, terraformPath),
		service.engineStatus(ctx, model.EngineOpenTofu, tofuPath),
	}
}

func (service *Service) engineStatus(ctx context.Context, kind model.EngineKind, binaryPath string) model.EngineStatus {
	name := "Terraform"
	if kind == model.EngineOpenTofu {
		name = "OpenTofu"
	}
	status := model.EngineStatus{Kind: kind, Name: name, CheckedAt: time.Now().UTC()}
	adapter, err := service.resolver.Resolve(kind, binaryPath, service.timeout)
	if err != nil {
		status.Error = &model.PublicFailure{Code: "ENGINE_NOT_AVAILABLE", Operation: "discover", Message: name + " is not available."}
		return status
	}
	metadata, err := adapter.Version(ctx)
	if err != nil {
		status.Error = &model.PublicFailure{Code: "ENGINE_VERSION_FAILED", Operation: "version", Message: name + " version could not be read."}
		return status
	}
	capabilities, err := adapter.Capabilities(ctx)
	if err != nil {
		status.Error = &model.PublicFailure{Code: "ENGINE_CAPABILITIES_FAILED", Operation: "capabilities", Message: name + " capabilities could not be detected."}
		return status
	}
	status.Available = true
	status.Path = adapter.BinaryPath()
	status.Version = metadata.Version
	status.Platform = metadata.Platform
	status.Capabilities = capabilities
	return status
}

// NewService constructs an application service.
func NewService(resolver Resolver, timeout time.Duration) *Service {
	if resolver == nil {
		resolver = DefaultResolver{}
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &Service{resolver: resolver, timeout: timeout}
}

// InspectRequest describes local state inspection.
type InspectRequest struct {
	Engine     model.EngineKind
	EnginePath string
	StatePath  string
}

// CompareRequest describes deterministic comparison of two local state files.
type CompareRequest struct {
	Engine                model.EngineKind
	EnginePath            string
	BeforePath            string
	AfterPath             string
	ProbableMoveThreshold float64
}

// PlanRehearsalRequest describes one restricted, refresh-disabled plan.
type PlanRehearsalRequest struct {
	Engine              model.EngineKind
	EnginePath          string
	ConfigPath          string
	StatePath           string
	InheritProxy        bool
	AllowRegistryAccess bool
	PathRoot            string
}

// UpgradeRehearsalRequest compares two explicit engine versions over identical inputs.
type UpgradeRehearsalRequest struct {
	FromEngine          model.EngineKind
	FromEnginePath      string
	ToEngine            model.EngineKind
	ToEnginePath        string
	ConfigPath          string
	StatePath           string
	InheritProxy        bool
	AllowRegistryAccess bool
	PathRoot            string
}

// MigrateRehearsalRequest compares Terraform and OpenTofu over identical inputs.
type MigrateRehearsalRequest struct {
	TerraformPath       string
	OpenTofuPath        string
	ConfigPath          string
	StatePath           string
	InheritProxy        bool
	AllowRegistryAccess bool
	PathRoot            string
}

// Inspect verifies input immutability and immediately normalizes raw engine output.
func (service *Service) Inspect(ctx context.Context, request InspectRequest) (model.InspectionResult, error) {
	statePath, beforeHash, err := inputHash(request.StatePath)
	if err != nil {
		return model.InspectionResult{}, err
	}
	adapter, err := service.resolver.Resolve(request.Engine, request.EnginePath, service.timeout)
	if err != nil {
		return model.InspectionResult{}, err
	}
	raw, err := adapter.ShowState(ctx, statePath)
	if err != nil {
		return model.InspectionResult{}, err
	}
	snapshot, normalizeErr := normalize.State(raw.JSON, request.Engine)
	clearBytes(raw.JSON)
	if normalizeErr != nil {
		return model.InspectionResult{}, normalizeErr
	}
	afterPath, afterHash, err := inputHash(statePath)
	if err != nil {
		return model.InspectionResult{}, err
	}
	if afterPath != statePath || beforeHash != afterHash {
		return model.InspectionResult{}, errors.New("supplied state changed during inspection")
	}
	snapshot.SourceHash = beforeHash
	snapshot.Engine = raw.Engine
	return model.InspectionResult{SchemaVersion: model.SchemaV1, Snapshot: snapshot, Findings: risk.Inspect(snapshot)}, nil
}

// Compare inspects each input through the same engine and compares normalized models.
func (service *Service) Compare(ctx context.Context, request CompareRequest) (model.StateComparison, error) {
	before, err := service.Inspect(ctx, InspectRequest{Engine: request.Engine, EnginePath: request.EnginePath, StatePath: request.BeforePath})
	if err != nil {
		return model.StateComparison{}, fmt.Errorf("inspect before state: %w", err)
	}
	after, err := service.Inspect(ctx, InspectRequest{Engine: request.Engine, EnginePath: request.EnginePath, StatePath: request.AfterPath})
	if err != nil {
		return model.StateComparison{}, fmt.Errorf("inspect after state: %w", err)
	}
	return comparepkg.States(before.Snapshot, after.Snapshot, comparepkg.Options{ProbableMoveThreshold: request.ProbableMoveThreshold}), nil
}

// RehearsePlan runs an engine only against copied inputs and returns sanitized output.
func (service *Service) RehearsePlan(ctx context.Context, request PlanRehearsalRequest) (result model.RehearsalResult, returnErr error) {
	var workspace *runner.PreparedWorkspace
	var err error
	if request.PathRoot == "" {
		workspace, err = runner.PrepareWorkspace(request.ConfigPath, request.StatePath)
	} else {
		workspace, err = runner.PrepareWorkspaceWithinRoot(request.ConfigPath, request.StatePath, request.PathRoot)
	}
	if err != nil {
		return result, err
	}
	defer func() {
		if verifyErr := workspace.VerifyOriginals(); verifyErr != nil && returnErr == nil {
			returnErr = verifyErr
		}
		if cleanupErr := workspace.Close(); cleanupErr != nil && returnErr == nil {
			returnErr = fmt.Errorf("cleanup rehearsal workspace: %w", cleanupErr)
		}
	}()
	adapter, err := service.resolver.Resolve(request.Engine, request.EnginePath, service.timeout)
	if err != nil {
		return result, err
	}

	stateRaw, err := adapter.ShowState(ctx, workspace.StatePath)
	if err != nil {
		return result, fmt.Errorf("verify copied state readability: %w", err)
	}
	_, stateNormalizeErr := normalize.State(stateRaw.JSON, request.Engine)
	clearBytes(stateRaw.JSON)
	if stateNormalizeErr != nil {
		return result, fmt.Errorf("normalize copied state: %w", stateNormalizeErr)
	}

	environment := runner.SafeEnvironment(runner.EnvironmentOptions{
		InheritProxy: request.InheritProxy, TemporaryHome: workspace.HomeDirectory, DataDirectory: workspace.DataDirectory,
	})
	planRaw, err := adapter.Plan(ctx, engine.PlanRequest{
		Workspace: engine.Workspace{Path: workspace.ConfigDirectory, Environment: environment}, StatePath: workspace.StatePath, PlanPath: workspace.PlanPath,
	})
	if err != nil {
		return result, err
	}
	plan, normalizeErr := normalize.Plan(planRaw.JSON, request.Engine)
	clearBytes(planRaw.JSON)
	if normalizeErr != nil {
		return result, normalizeErr
	}
	plan.Engine = planRaw.Engine
	checks := []model.CompatibilityCheck{
		{Name: "state_readable", Verification: model.VerificationVerified, Message: "The copied state was decoded by the selected engine."},
		{Name: "plan_generated", Verification: model.VerificationVerified, Message: "A refresh-disabled plan was generated in the copied workspace."},
		{Name: "original_inputs_unchanged", Verification: model.VerificationVerified, Message: "Original configuration and state hashes remained unchanged."},
		{Name: "real_cloud_api_behavior", Verification: model.VerificationUnverified, Message: "Cloud behavior was not tested."},
	}
	findings := risk.Plan(plan)
	var schemaSet *model.ProviderSchemaSet
	schemaRaw, schemaErr := adapter.ProviderSchemas(ctx, engine.Workspace{Path: workspace.ConfigDirectory, Environment: environment})
	if schemaErr == nil {
		normalizedSchemas, normalizeSchemaErr := normalize.ProviderSchemas(schemaRaw.JSON, schemaRaw.Engine)
		clearBytes(schemaRaw.JSON)
		if normalizeSchemaErr == nil {
			schemaSet = &normalizedSchemas
			checks = append(checks, model.CompatibilityCheck{Name: "provider_schemas", Verification: model.VerificationVerified, Message: "Provider schemas were fingerprinted without exposing schema content."})
		} else {
			schemaErr = normalizeSchemaErr
		}
	}
	if schemaErr != nil {
		checks = append(checks, model.CompatibilityCheck{Name: "provider_schemas", Verification: model.VerificationFailed, Message: "Provider schemas could not be normalized."})
		findings = append(findings, model.Finding{
			ID: "PP-PROV-003", Category: model.CategoryProviderSchema, Severity: model.SeverityWarning,
			Verification: model.VerificationFailed, Confidence: 0, Title: "Provider schemas unavailable",
			Explanation: "The selected engine could not provide a normalized provider schema set.", Evidence: []model.Evidence{},
			Recommendation: "Review engine initialization diagnostics and provider availability.",
		})
	}

	limitations := []string{
		"PlanParity is not an operating-system sandbox and cannot guarantee zero network activity.",
		"Real cloud API behavior remains unverified because refresh is disabled and known credentials are removed.",
	}
	if !request.AllowRegistryAccess {
		limitations = append(limitations, "Registry access was not authorized; direct provider network behavior still cannot be prevented by this process-level boundary.")
	}
	result = model.RehearsalResult{
		SchemaVersion: model.SchemaV1, Kind: model.RehearsalPlan, ExecutionMode: model.LocalRestricted,
		Candidate: &plan, CandidateSchemas: schemaSet, SchemaDifferences: []model.ProviderSchemaDifference{},
		Checks: checks, Findings: findings, Limitations: limitations,
	}
	return result, nil
}

// RehearseUpgrade compares baseline and candidate engine versions.
func (service *Service) RehearseUpgrade(ctx context.Context, request UpgradeRehearsalRequest) (model.RehearsalResult, error) {
	baseline, err := service.RehearsePlan(ctx, PlanRehearsalRequest{
		Engine: request.FromEngine, EnginePath: request.FromEnginePath, ConfigPath: request.ConfigPath, StatePath: request.StatePath,
		InheritProxy: request.InheritProxy, AllowRegistryAccess: request.AllowRegistryAccess, PathRoot: request.PathRoot,
	})
	if err != nil {
		return model.RehearsalResult{}, fmt.Errorf("baseline rehearsal: %w", err)
	}
	candidate, err := service.RehearsePlan(ctx, PlanRehearsalRequest{
		Engine: request.ToEngine, EnginePath: request.ToEnginePath, ConfigPath: request.ConfigPath, StatePath: request.StatePath,
		InheritProxy: request.InheritProxy, AllowRegistryAccess: request.AllowRegistryAccess, PathRoot: request.PathRoot,
	})
	if err != nil {
		return model.RehearsalResult{}, fmt.Errorf("candidate rehearsal: %w", err)
	}
	return combineRehearsals(model.RehearsalUpgrade, baseline, candidate), nil
}

// RehearseMigrate compares Terraform baseline behavior with OpenTofu candidate behavior.
func (service *Service) RehearseMigrate(ctx context.Context, request MigrateRehearsalRequest) (model.RehearsalResult, error) {
	baseline, err := service.RehearsePlan(ctx, PlanRehearsalRequest{
		Engine: model.EngineTerraform, EnginePath: request.TerraformPath, ConfigPath: request.ConfigPath, StatePath: request.StatePath,
		InheritProxy: request.InheritProxy, AllowRegistryAccess: request.AllowRegistryAccess, PathRoot: request.PathRoot,
	})
	if err != nil {
		return model.RehearsalResult{}, fmt.Errorf("Terraform rehearsal: %w", err)
	}
	candidate, err := service.RehearsePlan(ctx, PlanRehearsalRequest{
		Engine: model.EngineOpenTofu, EnginePath: request.OpenTofuPath, ConfigPath: request.ConfigPath, StatePath: request.StatePath,
		InheritProxy: request.InheritProxy, AllowRegistryAccess: request.AllowRegistryAccess, PathRoot: request.PathRoot,
	})
	if err != nil {
		return model.RehearsalResult{}, fmt.Errorf("OpenTofu rehearsal: %w", err)
	}
	return combineRehearsals(model.RehearsalMigrate, baseline, candidate), nil
}

func combineRehearsals(kind model.RehearsalKind, baseline, candidate model.RehearsalResult) model.RehearsalResult {
	planComparison := comparepkg.Plans(*baseline.Candidate, *candidate.Candidate)
	schemaDifferences := []model.ProviderSchemaDifference{}
	providerVerification := model.VerificationUnsupported
	providerMessage := "Provider schemas were unavailable for one or both runs."
	if baseline.CandidateSchemas != nil && candidate.CandidateSchemas != nil {
		schemaDifferences = comparepkg.ProviderSchemas(*baseline.CandidateSchemas, *candidate.CandidateSchemas)
		providerVerification = model.VerificationVerified
		providerMessage = "Provider schema fingerprints are equivalent."
		if len(schemaDifferences) > 0 {
			providerVerification = model.VerificationFailed
			providerMessage = fmt.Sprintf("%d provider schema fingerprint differences were detected.", len(schemaDifferences))
		}
	}
	planVerification := model.VerificationVerified
	planMessage := "Normalized plan behavior is equivalent."
	if !planComparison.Equivalent {
		planVerification = model.VerificationFailed
		planMessage = "Normalized plan behavior differs."
	}
	findings := append([]model.Finding{}, candidate.Findings...)
	findings = append(findings, risk.Parity(*baseline.Candidate, *candidate.Candidate, planComparison, schemaDifferences)...)
	findings = deduplicateFindings(findings)
	return model.RehearsalResult{
		SchemaVersion: model.SchemaV1, Kind: kind, ExecutionMode: model.LocalRestricted,
		Baseline: baseline.Candidate, Candidate: candidate.Candidate, Comparison: &planComparison,
		BaselineSchemas: baseline.CandidateSchemas, CandidateSchemas: candidate.CandidateSchemas, SchemaDifferences: schemaDifferences,
		Checks: []model.CompatibilityCheck{
			{Name: "baseline_state_readable", Verification: model.VerificationVerified, Message: "The baseline engine decoded the copied state."},
			{Name: "candidate_state_readable", Verification: model.VerificationVerified, Message: "The candidate engine decoded the copied state."},
			{Name: "provider_schema_equivalent", Verification: providerVerification, Message: providerMessage},
			{Name: "plan_equivalent", Verification: planVerification, Message: planMessage},
			{Name: "original_inputs_unchanged", Verification: model.VerificationVerified, Message: "Both isolated runs preserved original input hashes."},
			{Name: "real_cloud_api_behavior", Verification: model.VerificationUnverified, Message: "Cloud behavior was not tested."},
		},
		Findings: findings, Limitations: candidate.Limitations,
	}
}

func deduplicateFindings(findings []model.Finding) []model.Finding {
	seen := make(map[string]bool)
	result := make([]model.Finding, 0, len(findings))
	for _, finding := range findings {
		key := finding.ID + "\x00" + finding.Address + "\x00" + finding.Title
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, finding)
	}
	return result
}

func inputHash(path string) (string, string, error) {
	if path == "" {
		return "", "", errors.New("state path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve state path: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", "", fmt.Errorf("read state metadata: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", "", errors.New("state path must be a regular file")
	}
	content, err := os.ReadFile(absolute)
	if err != nil {
		return "", "", fmt.Errorf("read state: %w", err)
	}
	hash := sha256.Sum256(content)
	clearBytes(content)
	return absolute, hex.EncodeToString(hash[:]), nil
}

func clearBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
