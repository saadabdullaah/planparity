// Package command defines PlanParity's command-line interface.
package command

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// Execute constructs and runs the root command.
func Execute(ctx context.Context) error {
	return NewRootCommand().ExecuteContext(ctx)
}

// NewRootCommand returns an isolated command tree suitable for tests.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "planparity",
		Short:         "Inspect and rehearse Terraform/OpenTofu state transitions locally",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}
		},
	}

	root.PersistentFlags().Bool("verbose", false, "enable diagnostic logging (never includes raw state)")
	root.AddCommand(newVersionCommand())
	root.AddCommand(newDoctorCommand())
	root.AddCommand(newInspectCommand())
	root.AddCommand(newCompareCommand())
	root.AddCommand(newRehearseCommand())
	root.AddCommand(newUICommand())
	return root
}
