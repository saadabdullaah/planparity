// Package model contains PlanParity's engine-neutral, value-free domain contracts.
package model

import "time"

// EngineKind identifies a supported external engine.
type EngineKind string

const (
	EngineTerraform EngineKind = "terraform"
	EngineOpenTofu  EngineKind = "opentofu"
)

// EngineMetadata records the engine that produced a normalized result.
type EngineMetadata struct {
	Kind     EngineKind `json:"kind"`
	Name     string     `json:"name"`
	Version  string     `json:"version"`
	Platform string     `json:"platform,omitempty"`
}

// Capabilities describes machine-readable engine operations available to an adapter.
type Capabilities struct {
	MachineReadableState bool `json:"machine_readable_state"`
	MachineReadablePlan  bool `json:"machine_readable_plan"`
	ProviderSchema       bool `json:"provider_schema"`
	ValidateJSON         bool `json:"validate_json"`
	RefreshDisabledPlan  bool `json:"refresh_disabled_plan"`
}

// EngineStatus is the doctor and system-endpoint view of an engine.
type EngineStatus struct {
	Kind         EngineKind     `json:"kind"`
	Name         string         `json:"name"`
	Available    bool           `json:"available"`
	Path         string         `json:"path,omitempty"`
	Version      string         `json:"version,omitempty"`
	Platform     string         `json:"platform,omitempty"`
	Capabilities Capabilities   `json:"capabilities"`
	CheckedAt    time.Time      `json:"checked_at"`
	Error        *PublicFailure `json:"error,omitempty"`
}

// PublicFailure is a sanitized, typed error suitable for reports and APIs.
type PublicFailure struct {
	Code      string `json:"code"`
	Operation string `json:"operation,omitempty"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
