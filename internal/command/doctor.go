package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	var terraformPath string
	var tofuPath string
	var asJSON bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Detect Terraform and OpenTofu binaries and capabilities",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			statuses := []model.EngineStatus{
				checkEngine(cmd, model.EngineTerraform, terraformPath, timeout),
				checkEngine(cmd, model.EngineOpenTofu, tofuPath, timeout),
			}
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(struct {
					SchemaVersion string               `json:"schema_version"`
					Engines       []model.EngineStatus `json:"engines"`
				}{SchemaVersion: model.SchemaV1, Engines: statuses})
			}
			fmt.Fprintln(cmd.OutOrStdout(), "PlanParity Doctor")
			fmt.Fprintln(cmd.OutOrStdout())
			for _, status := range statuses {
				if !status.Available {
					fmt.Fprintf(cmd.OutOrStdout(), "%-10s NOT AVAILABLE  %s\n", status.Name, status.Error.Message)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-10s AVAILABLE      %s (%s)\n", status.Name, status.Version, status.Platform)
				fmt.Fprintf(cmd.OutOrStdout(), "           %s\n", status.Path)
				fmt.Fprintf(cmd.OutOrStdout(), "           state-json=%t plan-json=%t schemas=%t validate-json=%t refresh-disabled=%t\n",
					status.Capabilities.MachineReadableState,
					status.Capabilities.MachineReadablePlan,
					status.Capabilities.ProviderSchema,
					status.Capabilities.ValidateJSON,
					status.Capabilities.RefreshDisabledPlan,
				)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&terraformPath, "terraform-bin", "", "explicit Terraform executable path")
	cmd.Flags().StringVar(&tofuPath, "tofu-bin", "", "explicit OpenTofu executable path")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print machine-readable status")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "per-engine detection timeout")
	return cmd
}

func checkEngine(cmd *cobra.Command, kind model.EngineKind, binaryPath string, timeout time.Duration) model.EngineStatus {
	name := "Terraform"
	if kind == model.EngineOpenTofu {
		name = "OpenTofu"
	}
	status := model.EngineStatus{Kind: kind, Name: name, CheckedAt: time.Now().UTC()}
	adapter, err := resolveEngine(kind, binaryPath, timeout)
	if err != nil {
		status.Error = publicFailure(err, "discover")
		return status
	}
	status.Path = adapter.BinaryPath()
	metadata, err := adapter.Version(cmd.Context())
	if err != nil {
		status.Error = publicFailure(err, "version")
		return status
	}
	capabilities, err := adapter.Capabilities(cmd.Context())
	if err != nil {
		status.Error = publicFailure(err, "capabilities")
		return status
	}
	status.Available = true
	status.Version = metadata.Version
	status.Platform = metadata.Platform
	status.Capabilities = capabilities
	return status
}
