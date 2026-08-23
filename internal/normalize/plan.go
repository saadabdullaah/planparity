package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/saadabdullaah/planparity/internal/model"
)

type rawPlan struct {
	FormatVersion    string          `json:"format_version"`
	TerraformVersion string          `json:"terraform_version"`
	ResourceChanges  []rawPlanChange `json:"resource_changes"`
	Diagnostics      []rawDiagnostic `json:"diagnostics"`
}

type rawPlanChange struct {
	Address      string    `json:"address"`
	Type         string    `json:"type"`
	ProviderName string    `json:"provider_name"`
	Change       rawChange `json:"change"`
}

type rawChange struct {
	Actions      []string        `json:"actions"`
	Before       json.RawMessage `json:"before"`
	After        json.RawMessage `json:"after"`
	AfterUnknown json.RawMessage `json:"after_unknown"`
	ReplacePaths [][]any         `json:"replace_paths"`
}

type rawDiagnostic struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
}

// Plan converts show -json plan output into a value-free action model.
func Plan(raw []byte, kind model.EngineKind) (model.PlanResult, error) {
	var document rawPlan
	if err := json.Unmarshal(raw, &document); err != nil {
		return model.PlanResult{}, &Error{Code: ErrorInvalidJSON, Message: "engine plan output is not valid JSON", Cause: err}
	}
	if err := validateFormatVersion(document.FormatVersion); err != nil {
		return model.PlanResult{}, err
	}
	engineName := "Terraform"
	if kind == model.EngineOpenTofu {
		engineName = "OpenTofu"
	}
	sourceHash := sha256.Sum256(raw)
	result := model.PlanResult{
		SchemaVersion: model.SchemaV1,
		Engine:        model.EngineMetadata{Kind: kind, Name: engineName, Version: document.TerraformVersion},
		Changes:       make([]model.PlanChange, 0, len(document.ResourceChanges)),
		Diagnostics:   make([]model.Diagnostic, 0, len(document.Diagnostics)),
		SourceHash:    hex.EncodeToString(sourceHash[:]),
	}
	for _, change := range document.ResourceChanges {
		actions, replacement := normalizeActions(change.Change.Actions)
		changedPaths, err := changedValuePaths(change.Change.Before, change.Change.After)
		if err != nil {
			return model.PlanResult{}, &Error{Code: ErrorInvalidJSON, Message: fmt.Sprintf("resource %q has invalid change structure", change.Address), Cause: err}
		}
		for _, path := range change.Change.ReplacePaths {
			changedPaths = append(changedPaths, renderPath(path))
		}
		sort.Strings(changedPaths)
		changedPaths = uniqueStrings(changedPaths)
		planChange := model.PlanChange{
			Address:      change.Address,
			ResourceType: change.Type,
			Provider:     change.ProviderName,
			Actions:      actions,
			Replacement:  replacement,
			ChangedPaths: changedPaths,
			UnknownPaths: markerPaths(change.Change.AfterUnknown),
		}
		result.Changes = append(result.Changes, planChange)
		incrementSummary(&result.Summary, actions)
	}
	for _, diagnostic := range document.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, model.Diagnostic{
			Severity: cleanDiagnostic(diagnostic.Severity, 32),
			Summary:  cleanDiagnostic(diagnostic.Summary, 512),
			Detail:   cleanDiagnostic(diagnostic.Detail, 2048),
		})
	}
	sort.Slice(result.Changes, func(i, j int) bool { return result.Changes[i].Address < result.Changes[j].Address })
	return result, nil
}

func normalizeActions(actions []string) ([]model.PlanAction, bool) {
	hasCreate := false
	hasDelete := false
	for _, action := range actions {
		hasCreate = hasCreate || action == "create"
		hasDelete = hasDelete || action == "delete"
	}
	if hasCreate && hasDelete {
		return []model.PlanAction{model.PlanReplace}, true
	}
	result := make([]model.PlanAction, 0, len(actions))
	for _, action := range actions {
		switch action {
		case "no-op":
			result = append(result, model.PlanNoOp)
		case "create":
			result = append(result, model.PlanCreate)
		case "read":
			result = append(result, model.PlanRead)
		case "update":
			result = append(result, model.PlanUpdate)
		case "delete":
			result = append(result, model.PlanDelete)
		default:
			result = append(result, model.PlanUnknown)
		}
	}
	if len(result) == 0 {
		result = []model.PlanAction{model.PlanUnknown}
	}
	return result, false
}

func incrementSummary(summary *model.PlanSummary, actions []model.PlanAction) {
	for _, action := range actions {
		switch action {
		case model.PlanCreate:
			summary.Creates++
		case model.PlanRead:
			summary.Reads++
		case model.PlanUpdate:
			summary.Updates++
		case model.PlanDelete:
			summary.Deletes++
		case model.PlanReplace:
			summary.Replacements++
		case model.PlanNoOp:
			summary.NoOps++
		default:
			summary.Unknown++
		}
	}
}

func changedValuePaths(beforeRaw, afterRaw json.RawMessage) ([]string, error) {
	before, err := decodeJSONValue(beforeRaw)
	if err != nil {
		return nil, err
	}
	after, err := decodeJSONValue(afterRaw)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	collectChangedPaths(before, after, "", &paths)
	sort.Strings(paths)
	return uniqueStrings(paths), nil
}

func collectChangedPaths(before, after any, path string, paths *[]string) {
	beforeObject, beforeIsObject := before.(map[string]any)
	afterObject, afterIsObject := after.(map[string]any)
	if beforeIsObject && afterIsObject {
		keySet := make(map[string]struct{}, len(beforeObject)+len(afterObject))
		for key := range beforeObject {
			keySet[key] = struct{}{}
		}
		for key := range afterObject {
			keySet[key] = struct{}{}
		}
		keys := make([]string, 0, len(keySet))
		for key := range keySet {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectChangedPaths(beforeObject[key], afterObject[key], joinPath(path, key), paths)
		}
		return
	}
	if !reflect.DeepEqual(before, after) {
		if path == "" {
			path = "$"
		}
		*paths = append(*paths, path)
	}
}

func renderPath(parts []any) string {
	var builder strings.Builder
	for index, part := range parts {
		switch typed := part.(type) {
		case string:
			if index > 0 {
				builder.WriteByte('.')
			}
			builder.WriteString(typed)
		case float64:
			fmt.Fprintf(&builder, "[%d]", int(typed))
		default:
			fmt.Fprintf(&builder, "[%v]", typed)
		}
	}
	return builder.String()
}

func cleanDiagnostic(value string, limit int) string {
	var builder strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\t' || (!unicode.IsControl(r) && r != '\x1b') {
			builder.WriteRune(r)
		}
		if builder.Len() >= limit {
			break
		}
	}
	return strings.TrimSpace(builder.String())
}
