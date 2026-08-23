package command

import (
	"encoding/json"
	"fmt"

	"github.com/saadabdullaah/planparity/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print PlanParity version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := buildinfo.Current()
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "PlanParity %s (%s, %s, %s)\n", info.Version, info.Commit, info.GoVersion, info.Platform)
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print machine-readable version information")
	return cmd
}
