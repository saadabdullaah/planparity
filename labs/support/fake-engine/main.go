// Command fake-engine is a deterministic Terraform/OpenTofu protocol fixture.
// It exists only for PlanParity's offline demo labs and never contacts a provider.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "fixture engine:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) < 2 {
		return errors.New("an engine operation is required")
	}
	kind, profile := identity(arguments[0])
	switch arguments[1] {
	case "version":
		return writeJSON(versionDocument(kind, profile))
	case "validate":
		return writeJSON(map[string]any{"valid": true, "diagnostics": []any{}})
	case "init":
		return nil
	case "plan":
		output := flagValue(arguments[2:], "-out=")
		if output == "" {
			return errors.New("plan output path is missing")
		}
		return os.WriteFile(output, []byte(profile+":"+kind), 0o600)
	case "show":
		if len(arguments) < 4 {
			return errors.New("show target is missing")
		}
		target := arguments[len(arguments)-1]
		if strings.HasSuffix(strings.ToLower(target), ".tfplan") {
			return writeJSON(planDocument(kind, profile))
		}
		file, err := os.Open(target)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(os.Stdout, file)
		return err
	case "providers":
		if len(arguments) >= 4 && arguments[2] == "schema" {
			return writeJSON(schemaDocument(profile))
		}
	}
	return fmt.Errorf("unsupported operation %q", strings.Join(arguments[1:], " "))
}

func identity(executable string) (kind, profile string) {
	name := strings.TrimSuffix(strings.ToLower(filepath.Base(executable)), ".exe")
	kind = "terraform"
	if name == "tofu" {
		kind = "tofu"
	}
	profile = strings.ToLower(filepath.Base(filepath.Dir(executable)))
	return kind, profile
}

func versionDocument(kind, profile string) map[string]any {
	version := "1.9.0-lab"
	platform := "fixture"
	if profile == "candidate" {
		version = "2.0.0-lab"
	}
	if kind == "tofu" {
		version = "1.10.0-lab"
	}
	return map[string]any{"terraform_version": version, "platform": platform}
}

func planDocument(kind, profile string) map[string]any {
	actions := []string{"no-op"}
	replacePaths := []any{}
	after := map[string]any{"input": "stable"}
	if profile == "candidate" {
		actions = []string{"delete", "create"}
		replacePaths = []any{[]string{"input"}}
		after = map[string]any{"input": "candidate"}
	} else if kind == "tofu" {
		actions = []string{"update"}
		after = map[string]any{"input": "tofu"}
	}
	change := map[string]any{
		"actions":       actions,
		"before":        map[string]any{"input": "stable"},
		"after":         after,
		"after_unknown": map[string]any{},
		"replace_paths": replacePaths,
	}
	return map[string]any{
		"format_version":    "1.2",
		"terraform_version": versionDocument(kind, profile)["terraform_version"],
		"resource_changes": []any{map[string]any{
			"address": "terraform_data.example", "mode": "managed", "type": "terraform_data",
			"name": "example", "provider_name": "terraform.io/builtin/terraform", "change": change,
		}},
	}
}

func schemaDocument(profile string) map[string]any {
	attribute := map[string]any{"type": "dynamic", "optional": true}
	if profile == "candidate" {
		attribute["description_kind"] = "plain"
	}
	return map[string]any{
		"format_version": "1.0",
		"provider_schemas": map[string]any{
			"terraform.io/builtin/terraform": map[string]any{
				"provider": map[string]any{"version": 0, "block": map[string]any{"attributes": map[string]any{}}},
				"resource_schemas": map[string]any{"terraform_data": map[string]any{
					"version": 0, "block": map[string]any{"attributes": map[string]any{"input": attribute}},
				}},
				"data_source_schemas": map[string]any{},
			},
		},
	}
}

func flagValue(arguments []string, prefix string) string {
	for _, argument := range arguments {
		if strings.HasPrefix(argument, prefix) {
			return strings.TrimPrefix(argument, prefix)
		}
	}
	return ""
}

func writeJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(value)
}
