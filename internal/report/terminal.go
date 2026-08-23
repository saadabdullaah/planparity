package report

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/saadabdullaah/planparity/internal/model"
)

// InspectionTerminal writes a value-free state inventory.
func InspectionTerminal(writer io.Writer, result model.InspectionResult) error {
	snapshot := result.Snapshot
	if _, err := fmt.Fprintln(writer, "PlanParity State Inspection"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Engine:     %s %s\n", terminalText(snapshot.Engine.Name), terminalText(snapshot.Engine.Version)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "State hash: %s\n", snapshot.SourceHash); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Resources:  %d\nModules:    %d\nProviders:  %d\nOutputs:    %d\n",
		len(snapshot.Resources), len(snapshot.Modules), len(snapshot.Providers), len(snapshot.Outputs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "Values:     HIDDEN"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nResources"); err != nil {
		return err
	}
	for _, resource := range snapshot.Resources {
		if _, err := fmt.Fprintf(writer, "  %-10s %-12s %s\n", terminalText(string(resource.Status)), terminalText(string(resource.Mode)), terminalText(resource.Address)); err != nil {
			return err
		}
	}
	if len(result.Findings) == 0 {
		_, err := fmt.Fprintln(writer, "\nFindings: none")
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nFindings"); err != nil {
		return err
	}
	for _, finding := range result.Findings {
		if _, err := fmt.Fprintf(writer, "  %s %s %s\n    %s\n", terminalText(string(finding.Severity)), terminalText(finding.ID), terminalText(finding.Title), terminalText(finding.Explanation)); err != nil {
			return err
		}
	}
	return nil
}

// ComparisonTerminal writes a concise deterministic comparison report.
func ComparisonTerminal(writer io.Writer, comparison model.StateComparison) error {
	if _, err := fmt.Fprintln(writer, "PlanParity State Comparison"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Added:           %d\nRemoved:         %d\nChanged:         %d\nProbable moves:  %d\nProvider changes:%d\n",
		comparison.Summary.Added, comparison.Summary.Removed, comparison.Summary.Changed,
		comparison.Summary.ProbableMoves, comparison.Summary.ProviderChanges); err != nil {
		return err
	}
	if len(comparison.ProbableMoves) > 0 {
		if _, err := fmt.Fprintln(writer, "\nProbable moves"); err != nil {
			return err
		}
		for _, move := range comparison.ProbableMoves {
			ambiguity := ""
			if move.Ambiguous {
				ambiguity = " AMBIGUOUS"
			}
			if _, err := fmt.Fprintf(writer, "  %.0f%%%-10s %s -> %s\n", move.Confidence*100, ambiguity, terminalText(move.FromAddress), terminalText(move.ToAddress)); err != nil {
				return err
			}
		}
	}
	if len(comparison.Findings) == 0 {
		_, err := fmt.Fprintln(writer, "\nFindings: none")
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nFindings"); err != nil {
		return err
	}
	for _, finding := range comparison.Findings {
		if _, err := fmt.Fprintf(writer, "  %s %s %s\n    %s\n", terminalText(string(finding.Severity)), terminalText(finding.ID), terminalText(finding.Title), terminalText(finding.Explanation)); err != nil {
			return err
		}
	}
	return nil
}

// RehearsalTerminal writes a safety-forward plan, upgrade, or parity result.
func RehearsalTerminal(writer io.Writer, result model.RehearsalResult) error {
	if _, err := fmt.Fprintf(writer, "PlanParity %s Rehearsal\n\n", terminalText(string(result.Kind))); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "Execution mode: %s\n", result.ExecutionMode); err != nil {
		return err
	}
	if result.Candidate != nil {
		summary := result.Candidate.Summary
		if _, err := fmt.Fprintf(writer, "\nPlan summary\n  Creates:      %d\n  Updates:      %d\n  Deletes:      %d\n  Replacements: %d\n  No-ops:       %d\n",
			summary.Creates, summary.Updates, summary.Deletes, summary.Replacements, summary.NoOps); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\nCompatibility"); err != nil {
		return err
	}
	for _, check := range result.Checks {
		if _, err := fmt.Fprintf(writer, "  %-30s %s\n", terminalText(check.Name), terminalText(string(check.Verification))); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\nFindings"); err != nil {
		return err
	}
	for _, finding := range result.Findings {
		if _, err := fmt.Fprintf(writer, "  %s %s\n    %s\n", terminalText(string(finding.Severity)), terminalText(finding.ID), terminalText(finding.Title)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\nOriginal state modified: NO\nApply executed: NO"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "\nPlanParity copied the workspace, disabled refresh, removed known credentials, and avoided apply operations. This is not an operating system security sandbox and does not guarantee zero network activity."); err != nil {
		return err
	}
	return nil
}

func terminalText(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsControl(character) || unicode.Is(unicode.Cf, character) {
			return -1
		}
		return character
	}, value)
}
