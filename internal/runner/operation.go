// Package runner executes a closed set of Terraform/OpenTofu operations.
package runner

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Operation is a closed command allowlist entry.
type Operation string

const (
	OperationVersion         Operation = "VERSION"
	OperationValidate        Operation = "VALIDATE"
	OperationShow            Operation = "SHOW"
	OperationProviderSchemas Operation = "PROVIDER_SCHEMAS"
	OperationInit            Operation = "INIT"
	OperationPlan            Operation = "PLAN"
)

// Request contains typed fields used to construct an allowlisted command.
type Request struct {
	Binary           string
	WorkingDirectory string
	Operation        Operation
	TargetPath       string
	OutputPath       string
	StatePath        string
	Environment      []string
	TimeoutLabel     string
}

// Arguments returns a fixed command shape for an allowlisted operation.
func Arguments(request Request) ([]string, error) {
	switch request.Operation {
	case OperationVersion:
		return []string{"version", "-json"}, nil
	case OperationValidate:
		return []string{"validate", "-json", "-no-color"}, nil
	case OperationShow:
		if request.TargetPath == "" {
			return nil, errors.New("show requires a target path")
		}
		return []string{"show", "-json", filepath.Clean(request.TargetPath)}, nil
	case OperationProviderSchemas:
		return []string{"providers", "schema", "-json"}, nil
	case OperationInit:
		return []string{"init", "-backend=false", "-input=false", "-no-color"}, nil
	case OperationPlan:
		if request.OutputPath == "" || request.StatePath == "" {
			return nil, errors.New("plan requires state and output paths")
		}
		return []string{
			"plan",
			"-refresh=false",
			"-input=false",
			"-no-color",
			"-lock=false",
			"-state=" + filepath.Clean(request.StatePath),
			"-out=" + filepath.Clean(request.OutputPath),
		}, nil
	default:
		return nil, fmt.Errorf("operation %q is not allowlisted", request.Operation)
	}
}
