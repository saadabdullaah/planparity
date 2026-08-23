package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/saadabdullaah/planparity/internal/model"
)

// Discover finds and validates a supported engine binary.
func Discover(kind model.EngineKind, explicitPath string) (string, error) {
	commandName := "terraform"
	if kind == model.EngineOpenTofu {
		commandName = "tofu"
	}
	path := explicitPath
	if path == "" {
		found, err := exec.LookPath(commandName)
		if err != nil {
			return "", &Failure{Code: "ENGINE_NOT_FOUND", Operation: "discover", Message: fmt.Sprintf("%s was not found on PATH", commandName), Cause: err}
		}
		path = found
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", &Failure{Code: "INVALID_ENGINE_PATH", Operation: "discover", Message: "engine path is invalid", Cause: err}
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		absolute = resolved
	}
	if !allowedBinaryName(kind, filepath.Base(absolute)) {
		return "", &Failure{Code: "INVALID_ENGINE_PATH", Operation: "discover", Message: fmt.Sprintf("%s path must name the expected executable", commandName)}
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.Mode().IsRegular() {
		return "", &Failure{Code: "INVALID_ENGINE_PATH", Operation: "discover", Message: fmt.Sprintf("%s path is not a regular file", commandName), Cause: err}
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		return "", &Failure{Code: "INVALID_ENGINE_PATH", Operation: "discover", Message: fmt.Sprintf("%s is not executable", commandName)}
	}
	return absolute, nil
}

func allowedBinaryName(kind model.EngineKind, name string) bool {
	name = strings.ToLower(name)
	if kind == model.EngineTerraform {
		return name == "terraform" || name == "terraform.exe"
	}
	return name == "tofu" || name == "tofu.exe"
}
