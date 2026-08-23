package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareWorkspaceCopiesAndCleans(t *testing.T) {
	config := t.TempDir()
	state := filepath.Join(t.TempDir(), "source.tfstate")
	if err := os.WriteFile(filepath.Join(config, "main.tf"), []byte("terraform {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config, "terraform.tfstate"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(config, ".terraform"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config, ".terraform", "cache"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state, []byte("synthetic-state"), 0o600); err != nil {
		t.Fatal(err)
	}

	workspace, err := PrepareWorkspace(config, state)
	if err != nil {
		t.Fatal(err)
	}
	root := workspace.Root
	if _, err := os.Stat(filepath.Join(workspace.ConfigDirectory, "main.tf")); err != nil {
		t.Fatal("main.tf not copied")
	}
	if _, err := os.Stat(filepath.Join(workspace.ConfigDirectory, ".terraform")); !os.IsNotExist(err) {
		t.Fatal("managed data was copied")
	}
	copiedState, err := os.ReadFile(workspace.StatePath)
	if err != nil || string(copiedState) != "synthetic-state" {
		t.Fatalf("bad copied state: %q %v", copiedState, err)
	}
	if err := workspace.VerifyOriginals(); err != nil {
		t.Fatal(err)
	}
	if err := workspace.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("workspace still exists: %v", err)
	}
}

func TestVerifyOriginalsDetectsMutation(t *testing.T) {
	config := t.TempDir()
	state := filepath.Join(t.TempDir(), "source.tfstate")
	_ = os.WriteFile(filepath.Join(config, "main.tf"), []byte("terraform {}"), 0o600)
	_ = os.WriteFile(state, []byte("before"), 0o600)
	workspace, err := PrepareWorkspace(config, state)
	if err != nil {
		t.Fatal(err)
	}
	defer workspace.Close()
	_ = os.WriteFile(state, []byte("after"), 0o600)
	if err := workspace.VerifyOriginals(); err == nil {
		t.Fatal("input mutation was not detected")
	}
}

func TestPrepareWorkspaceSupportsSpacesAndLongPaths(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "configuration with spaces", strings.Repeat("nested", 18))
	if err := os.MkdirAll(config, 0o700); err != nil {
		t.Fatal(err)
	}
	stateDirectory := filepath.Join(root, "state files with spaces")
	if err := os.MkdirAll(stateDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(stateDirectory, "source state.tfstate")
	if err := os.WriteFile(filepath.Join(config, "main.tf"), []byte("terraform {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state, []byte("synthetic-state"), 0o600); err != nil {
		t.Fatal(err)
	}
	workspace, err := PrepareWorkspace(config, state)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = workspace.Close() })
	if err := workspace.VerifyOriginals(); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareWorkspaceRejectsLinkedInputs(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(config, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, "state.tfstate")
	if err := os.WriteFile(state, []byte("synthetic-state"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.tf"), []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.tf"), filepath.Join(config, "escape.tf")); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	if _, err := PrepareWorkspace(config, state); err == nil || !strings.Contains(err.Error(), "link") {
		t.Fatalf("linked configuration was not rejected: %v", err)
	}
}

func TestFileHashRejectsOversizedInputWithoutReadingIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.tfstate")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxCopiedFileSize + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := FileHash(path); err == nil || !strings.Contains(err.Error(), "input limit") {
		t.Fatalf("oversized input was not rejected: %v", err)
	}
}
