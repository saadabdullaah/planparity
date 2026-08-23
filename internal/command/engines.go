package command

import (
	"fmt"
	"time"

	"github.com/saadabdullaah/planparity/internal/engine"
	"github.com/saadabdullaah/planparity/internal/engine/opentofu"
	"github.com/saadabdullaah/planparity/internal/engine/terraform"
	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/runner"
)

func resolveEngine(kind model.EngineKind, binaryPath string, timeout time.Duration) (engine.Engine, error) {
	binary, err := engine.Discover(kind, binaryPath)
	if err != nil {
		return nil, err
	}
	executor := runner.ProcessExecutor{Timeout: timeout}
	switch kind {
	case model.EngineTerraform:
		return terraform.New(binary, executor), nil
	case model.EngineOpenTofu:
		return opentofu.New(binary, executor), nil
	default:
		return nil, fmt.Errorf("unsupported engine %q", kind)
	}
}

func parseEngineKind(value string) (model.EngineKind, error) {
	switch value {
	case "terraform":
		return model.EngineTerraform, nil
	case "tofu", "opentofu":
		return model.EngineOpenTofu, nil
	default:
		return "", fmt.Errorf("engine must be terraform or tofu")
	}
}

func publicFailure(err error, operation string) *model.PublicFailure {
	if failure, ok := err.(*engine.Failure); ok {
		result := failure.Public()
		return &result
	}
	return &model.PublicFailure{Code: "UNEXPECTED_ERROR", Operation: operation, Message: err.Error()}
}
