package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
	"github.com/saadabdullaah/planparity/internal/report"
	"github.com/spf13/cobra"
)

func newInspectCommand() *cobra.Command {
	var engineName string
	var enginePath string
	var statePath string
	var format string
	var outputPath string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect local state without exposing attribute values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind, err := parseEngineKind(strings.ToLower(engineName))
			if err != nil {
				return err
			}
			if statePath == "" {
				return errors.New("--state is required")
			}
			if format != "terminal" && format != "json" {
				return errors.New("--format must be terminal or json")
			}
			service := app.NewService(nil, timeout)
			result, err := service.Inspect(cmd.Context(), app.InspectRequest{Engine: kind, EnginePath: enginePath, StatePath: statePath})
			if err != nil {
				return fmt.Errorf("inspect state: %w", err)
			}
			writer, closeWriter, err := commandWriter(cmd.OutOrStdout(), outputPath)
			if err != nil {
				return err
			}
			defer closeWriter()
			if format == "json" {
				return report.JSON(writer, result)
			}
			return report.InspectionTerminal(writer, result)
		},
	}
	cmd.Flags().StringVar(&engineName, "engine", "terraform", "engine: terraform or tofu")
	cmd.Flags().StringVar(&enginePath, "engine-bin", "", "explicit engine executable path")
	cmd.Flags().StringVar(&statePath, "state", "", "local state file to inspect")
	cmd.Flags().StringVar(&format, "format", "terminal", "output format: terminal or json")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write the report to a file")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "engine command timeout")
	return cmd
}

func commandWriter(fallback io.Writer, path string) (io.Writer, func(), error) {
	if path == "" {
		return fallback, func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("create output: %w", err)
	}
	return file, func() { _ = file.Close() }, nil
}
