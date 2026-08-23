package engine

import (
	"errors"
	"strings"
	"testing"

	"github.com/saadabdullaah/planparity/internal/runner"
)

func TestExecutionFailureDoesNotExposeEngineStderr(t *testing.T) {
	canary := "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9"
	internal := &runner.ExecutionError{
		Operation: runner.OperationPlan,
		ExitCode:  1,
		Stderr:    []byte("provider error included " + canary),
		Cause:     errors.New("exit status 1"),
	}
	public := executionFailure("plan", internal)
	if strings.Contains(public.Error(), canary) || strings.Contains(public.Error(), "provider error included") {
		t.Fatalf("public engine error exposed stderr: %s", public)
	}
}

func TestSanitizeMessageStripsTerminalControls(t *testing.T) {
	t.Parallel()
	message := sanitizeMessage("safe\x1b[31m\r\nline\u202e", 512)
	for _, forbidden := range []string{"\x1b", "\r", "\u202e"} {
		if strings.Contains(message, forbidden) {
			t.Fatalf("sanitized message retained %q: %q", forbidden, message)
		}
	}
}
