package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxCopiedFileSize  = 256 << 20 // 256 MiB.
	maxCopiedFileCount = 100_000
)

// PreparedWorkspace is an isolated copy used for rehearsal execution.
type PreparedWorkspace struct {
	Root            string
	ConfigDirectory string
	StatePath       string
	PlanPath        string
	HomeDirectory   string
	DataDirectory   string
	OriginalConfig  string
	OriginalState   string
	configHash      string
	stateHash       string
}

// PrepareWorkspace copies configuration and state without following links.
func PrepareWorkspace(configPath, statePath string) (*PreparedWorkspace, error) {
	return prepareWorkspace(configPath, statePath)
}

// PrepareWorkspaceWithinRoot copies browser-selected inputs only after proving
// that both relative components are local to the server-authorized root.
func PrepareWorkspaceWithinRoot(configPath, statePath, pathRoot string) (*PreparedWorkspace, error) {
	if pathRoot == "" {
		return nil, errors.New("authorized path root is required")
	}
	root, err := filepath.Abs(pathRoot)
	if err != nil {
		return nil, errors.New("authorized path root is invalid")
	}
	configRelative, err := filepath.Rel(root, configPath)
	if err != nil || !filepath.IsLocal(configRelative) {
		return nil, errors.New("configuration path is outside the authorized root")
	}
	stateRelative, err := filepath.Rel(root, statePath)
	if err != nil || !filepath.IsLocal(stateRelative) {
		return nil, errors.New("state path is outside the authorized root")
	}
	return prepareWorkspace(filepath.Join(root, configRelative), filepath.Join(root, stateRelative))
}

func prepareWorkspace(configPath, statePath string) (*PreparedWorkspace, error) {
	config, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve configuration: %w", err)
	}
	state, err := filepath.Abs(statePath)
	if err != nil {
		return nil, fmt.Errorf("resolve state: %w", err)
	}
	configInfo, err := os.Stat(config)
	if err != nil || !configInfo.IsDir() {
		return nil, errors.New("configuration path must be a directory")
	}
	stateInfo, err := os.Lstat(state)
	if err != nil || !stateInfo.Mode().IsRegular() {
		return nil, errors.New("state path must be a regular file")
	}
	if stateInfo.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("state path cannot be a symbolic link")
	}
	configHash, err := TreeHash(config)
	if err != nil {
		return nil, fmt.Errorf("hash configuration: %w", err)
	}
	stateHash, err := FileHash(state)
	if err != nil {
		return nil, fmt.Errorf("hash state: %w", err)
	}

	tempRoot := os.Getenv("PLANPARITY_TEMP_DIR")
	if tempRoot != "" {
		tempRoot, err = filepath.Abs(tempRoot)
		if err != nil {
			return nil, fmt.Errorf("resolve temporary root: %w", err)
		}
		if err := os.MkdirAll(tempRoot, 0o700); err != nil {
			return nil, fmt.Errorf("create temporary root: %w", err)
		}
	}
	root, err := os.MkdirTemp(tempRoot, "planparity-")
	if err != nil {
		return nil, fmt.Errorf("create temporary workspace: %w", err)
	}
	prepared := &PreparedWorkspace{
		Root: root, ConfigDirectory: filepath.Join(root, "workspace"), StatePath: filepath.Join(root, "workspace", "terraform.tfstate"),
		PlanPath: filepath.Join(root, "plan.tfplan"), HomeDirectory: filepath.Join(root, "home"), DataDirectory: filepath.Join(root, "data"),
		OriginalConfig: config, OriginalState: state, configHash: configHash, stateHash: stateHash,
	}
	if err := os.MkdirAll(prepared.ConfigDirectory, 0o700); err != nil {
		_ = prepared.Close()
		return nil, err
	}
	if err := os.MkdirAll(prepared.HomeDirectory, 0o700); err != nil {
		_ = prepared.Close()
		return nil, err
	}
	if err := os.MkdirAll(prepared.DataDirectory, 0o700); err != nil {
		_ = prepared.Close()
		return nil, err
	}
	if err := copyConfiguration(config, prepared.ConfigDirectory, state); err != nil {
		_ = prepared.Close()
		return nil, err
	}
	if err := copyFile(state, prepared.StatePath, 0o600); err != nil {
		_ = prepared.Close()
		return nil, fmt.Errorf("copy state: %w", err)
	}
	return prepared, nil
}

// VerifyOriginals confirms supplied inputs remain byte-identical.
func (workspace *PreparedWorkspace) VerifyOriginals() error {
	configHash, err := TreeHash(workspace.OriginalConfig)
	if err != nil {
		return err
	}
	stateHash, err := FileHash(workspace.OriginalState)
	if err != nil {
		return err
	}
	if configHash != workspace.configHash {
		return errors.New("original configuration changed during rehearsal")
	}
	if stateHash != workspace.stateHash {
		return errors.New("original state changed during rehearsal")
	}
	return nil
}

// Close removes all copied configuration, state, plan, home, and data files.
func (workspace *PreparedWorkspace) Close() error {
	if workspace == nil || workspace.Root == "" {
		return nil
	}
	return os.RemoveAll(workspace.Root)
}

func copyConfiguration(source, destination, excludedState string) error {
	fileCount := 0
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if samePath(absolute, excludedState) {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("configuration entry escapes source root")
		}
		if relative == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("configuration link %q is not supported", relative)
		}
		if shouldExclude(relative, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("configuration entry %q is not a regular file", relative)
		}
		fileCount++
		if fileCount > maxCopiedFileCount {
			return fmt.Errorf("configuration exceeds the %d file copy limit", maxCopiedFileCount)
		}
		if info.Size() > maxCopiedFileSize {
			return fmt.Errorf("configuration file %q exceeds the copy limit", relative)
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func shouldExclude(relative string, directory bool) bool {
	name := filepath.Base(relative)
	if directory {
		switch name {
		case ".git", ".terraform", ".tofu", ".planparity", "node_modules":
			return true
		}
	}
	lower := strings.ToLower(name)
	return lower == "terraform.tfstate" || strings.HasPrefix(lower, "terraform.tfstate.") || strings.HasSuffix(lower, ".tfplan")
}

func copyFile(source, destination string, mode fs.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// FileHash returns a SHA-256 hash without retaining file data.
func FileHash(path string) (string, error) {
	// Browser-originated paths are contained by the server's startup path root;
	// CLI-originated paths are explicit capabilities supplied by the local user.
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("hash input must be a regular file")
	}
	if info.Size() > maxCopiedFileSize {
		return "", fmt.Errorf("file exceeds the %d byte input limit", maxCopiedFileSize)
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// TreeHash deterministically hashes regular file paths, modes, and contents.
func TreeHash(root string) (string, error) {
	entries := make([]string, 0)
	fileCount := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("configuration link %q is not supported", path)
		}
		if entry.IsDir() {
			if path != root {
				relative, _ := filepath.Rel(root, path)
				if shouldExclude(relative, true) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("configuration entry %q is not regular", path)
		}
		fileCount++
		if fileCount > maxCopiedFileCount {
			return fmt.Errorf("configuration exceeds the %d file input limit", maxCopiedFileCount)
		}
		if info.Size() > maxCopiedFileSize {
			return fmt.Errorf("configuration file %q exceeds the input limit", path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if shouldExclude(relative, false) {
			return nil
		}
		entries = append(entries, relative)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)
	hash := sha256.New()
	for _, relative := range entries {
		path := filepath.Join(root, relative)
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%s\x00%o\x00", filepath.ToSlash(relative), info.Mode().Perm())
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func samePath(first, second string) bool {
	first = filepath.Clean(first)
	second = filepath.Clean(second)
	if os.PathSeparator == '\\' {
		return strings.EqualFold(first, second)
	}
	return first == second
}
