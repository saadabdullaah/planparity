package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func TestDiscoverAcceptsOnlyTheExpectedExecutableName(t *testing.T) {
	t.Parallel()
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	root := t.TempDir()
	valid := filepath.Join(root, "terraform"+suffix)
	if err := os.WriteFile(valid, []byte("synthetic"), 0o700); err != nil {
		t.Fatal(err)
	}
	resolved, err := Discover(model.EngineTerraform, valid)
	if err != nil {
		t.Fatal(err)
	}
	resolvedInfo, err := os.Stat(resolved)
	if err != nil {
		t.Fatal(err)
	}
	validInfo, err := os.Stat(valid)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(resolvedInfo, validInfo) {
		t.Fatalf("resolved path = %q, want %q", resolved, valid)
	}
	wrong := filepath.Join(root, "terraform-wrapper"+suffix)
	if err := os.WriteFile(wrong, []byte("synthetic"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(model.EngineTerraform, wrong); err == nil {
		t.Fatal("unexpected executable name was accepted")
	}
}

func TestDiscoverRejectsADeceptiveEngineSymlink(t *testing.T) {
	t.Parallel()
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	root := t.TempDir()
	target := filepath.Join(root, "not-an-engine"+suffix)
	link := filepath.Join(root, "terraform"+suffix)
	if err := os.WriteFile(target, []byte("synthetic"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	if _, err := Discover(model.EngineTerraform, link); err == nil {
		t.Fatal("deceptive engine link was accepted")
	}
}
