package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
	"github.com/saadabdullaah/planparity/internal/report"
	"github.com/spf13/cobra"
)

func newCompareCommand() *cobra.Command {
	var engineName, enginePath, beforePath, afterPath, format, outputPath string
	var threshold float64
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare two local state files deterministically",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind, err := parseEngineKind(strings.ToLower(engineName))
			if err != nil {
				return err
			}
			if beforePath == "" || afterPath == "" {
				return errors.New("--before and --after are required")
			}
			if threshold < 0.5 || threshold > 1 {
				return errors.New("--probable-move-threshold must be between 0.5 and 1")
			}
			if format != "terminal" && format != "json" {
				return errors.New("--format must be terminal or json")
			}
			service := app.NewService(nil, timeout)
			result, err := service.Compare(cmd.Context(), app.CompareRequest{
				Engine: kind, EnginePath: enginePath, BeforePath: beforePath, AfterPath: afterPath, ProbableMoveThreshold: threshold,
			})
			if err != nil {
				return fmt.Errorf("compare states: %w", err)
			}
			writer, closeWriter, err := commandWriter(cmd.OutOrStdout(), outputPath)
			if err != nil {
				return err
			}
			defer closeWriter()
			if format == "json" {
				return report.JSON(writer, result)
			}
			return report.ComparisonTerminal(writer, result)
		},
	}
	cmd.Flags().StringVar(&engineName, "engine", "terraform", "engine: terraform or tofu")
	cmd.Flags().StringVar(&enginePath, "engine-bin", "", "explicit engine executable path")
	cmd.Flags().StringVar(&beforePath, "before", "", "earlier local state file")
	cmd.Flags().StringVar(&afterPath, "after", "", "later local state file")
	cmd.Flags().Float64Var(&threshold, "probable-move-threshold", 0.85, "minimum probable-move confidence")
	cmd.Flags().StringVar(&format, "format", "terminal", "output format: terminal or json")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write the report to a file")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "per-engine command timeout")
	return cmd
}
