// Package terraform constructs Terraform engine adapters.
package terraform

import (
	"github.com/saadabdullaah/planparity/internal/engine"
	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/runner"
)

// New returns a Terraform adapter for a validated binary path.
func New(binary string, executor runner.Executor) engine.Engine {
	return engine.NewCommandEngine(model.EngineTerraform, "Terraform", binary, executor)
}
