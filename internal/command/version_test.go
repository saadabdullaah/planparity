package command

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(output.String(), "PlanParity dev") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestVersionCommandJSON(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"version", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version --json failed: %v", err)
	}
	if !strings.Contains(output.String(), `"version": "dev"`) {
		t.Fatalf("unexpected JSON: %q", output.String())
	}
}
