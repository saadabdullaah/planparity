// Package opentofu constructs OpenTofu engine adapters.
package opentofu

import (
	"github.com/saadabdullaah/planparity/internal/engine"
	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/runner"
)

// New returns an OpenTofu adapter for a validated binary path.
func New(binary string, executor runner.Executor) engine.Engine {
	return engine.NewCommandEngine(model.EngineOpenTofu, "OpenTofu", binary, executor)
}
