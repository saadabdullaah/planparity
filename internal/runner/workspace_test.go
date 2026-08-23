package runner

import (
	"os"
	"path/filepath"
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
