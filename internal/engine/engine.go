// Package engine defines the shared Terraform/OpenTofu adapter boundary.
package engine

import (
	"context"
	"time"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Workspace is a prepared configuration directory.
type Workspace struct {
	Path        string
	Environment []string
}

// ValidationResult contains sanitized validation status and diagnostics.
type ValidationResult struct {
	Valid       bool               `json:"valid"`
	Diagnostics []model.Diagnostic `json:"diagnostics"`
}

// RawStateResult remains inside the backend and must be normalized immediately.
type RawStateResult struct {
	JSON     []byte
	Engine   model.EngineMetadata
	Duration time.Duration
}

// RawProviderSchemas is internal-only provider schema output.
type RawProviderSchemas struct {
	JSON     []byte
	Engine   model.EngineMetadata
	Duration time.Duration
}

// RawPlanResult remains inside the backend and must be normalized immediately.
type RawPlanResult struct {
	JSON     []byte
	Engine   model.EngineMetadata
	Duration time.Duration
}

// PlanRequest describes a prepared, copied rehearsal workspace.
type PlanRequest struct {
	Workspace Workspace
	StatePath string
	PlanPath  string
}

// Engine is implemented equally by Terraform and OpenTofu adapters.
type Engine interface {
	Name() string
	Kind() model.EngineKind
	BinaryPath() string

	Version(context.Context) (model.EngineMetadata, error)
	Capabilities(context.Context) (model.Capabilities, error)
	Validate(context.Context, Workspace) (ValidationResult, error)
	ShowState(context.Context, string) (RawStateResult, error)
	ProviderSchemas(context.Context, Workspace) (RawProviderSchemas, error)
	Plan(context.Context, PlanRequest) (RawPlanResult, error)
}
