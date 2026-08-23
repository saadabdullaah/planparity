package command

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
	"github.com/saadabdullaah/planparity/internal/model"
	"github.com/saadabdullaah/planparity/internal/report"
	"github.com/spf13/cobra"
)

func newRehearseCommand() *cobra.Command {
	command := &cobra.Command{Use: "rehearse", Short: "Run compatibility workflows in copied local workspaces"}
	command.AddCommand(newRehearsePlanCommand())
	command.AddCommand(newRehearseUpgradeCommand())
	command.AddCommand(newRehearseMigrateCommand())
	return command
}

func newRehearsePlanCommand() *cobra.Command {
	var engineName, enginePath, configPath, statePath, format, outputPath string
	var inheritProxy, allowRegistry bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generate a refresh-disabled plan in a copied workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind, err := parseEngineKind(strings.ToLower(engineName))
			if err != nil {
				return err
			}
			if configPath == "" || statePath == "" {
				return errors.New("--config and --state are required")
			}
			if format != "terminal" && format != "json" {
				return errors.New("--format must be terminal or json")
			}
			service := app.NewService(nil, timeout)
			result, err := service.RehearsePlan(cmd.Context(), app.PlanRehearsalRequest{
				Engine: kind, EnginePath: enginePath, ConfigPath: configPath, StatePath: statePath,
				InheritProxy: inheritProxy, AllowRegistryAccess: allowRegistry,
			})
			if err != nil {
				return fmt.Errorf("rehearse plan: %w", err)
			}
			writer, closeWriter, err := commandWriter(cmd.OutOrStdout(), outputPath)
			if err != nil {
				return err
			}
			defer closeWriter()
			if format == "json" {
				return report.JSON(writer, result)
			}
			return report.RehearsalTerminal(writer, result)
		},
	}
	cmd.Flags().StringVar(&engineName, "engine", "terraform", "engine: terraform or tofu")
	cmd.Flags().StringVar(&enginePath, "engine-bin", "", "explicit engine executable path")
	cmd.Flags().StringVar(&configPath, "config", "", "local configuration directory")
	cmd.Flags().StringVar(&statePath, "state", "", "local state file")
	cmd.Flags().BoolVar(&inheritProxy, "inherit-proxy", false, "allow proxy variables in the reduced engine environment")
	cmd.Flags().BoolVar(&allowRegistry, "allow-registry-access", true, "authorize provider/module registry access during copied init")
	cmd.Flags().StringVar(&format, "format", "terminal", "output format: terminal or json")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write the report to a file")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "per-engine command timeout")
	return cmd
}

func newRehearseUpgradeCommand() *cobra.Command {
	var fromName, fromPath, toName, toPath, configPath, statePath, format, outputPath string
	var inheritProxy, allowRegistry bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Compare baseline and candidate engine versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fromKind, err := parseEngineKind(strings.ToLower(fromName))
			if err != nil {
				return fmt.Errorf("from engine: %w", err)
			}
			toKind, err := parseEngineKind(strings.ToLower(toName))
			if err != nil {
				return fmt.Errorf("to engine: %w", err)
			}
			if fromPath == "" || toPath == "" || configPath == "" || statePath == "" {
				return errors.New("--from-bin, --to-bin, --config, and --state are required")
			}
			if format != "terminal" && format != "json" {
				return errors.New("--format must be terminal or json")
			}
			service := app.NewService(nil, timeout)
			result, err := service.RehearseUpgrade(cmd.Context(), app.UpgradeRehearsalRequest{
				FromEngine: fromKind, FromEnginePath: fromPath, ToEngine: toKind, ToEnginePath: toPath,
				ConfigPath: configPath, StatePath: statePath, InheritProxy: inheritProxy, AllowRegistryAccess: allowRegistry,
			})
			if err != nil {
				return fmt.Errorf("rehearse upgrade: %w", err)
			}
			return writeRehearsal(cmd, result, format, outputPath)
		},
	}
	cmd.Flags().StringVar(&fromName, "from-engine", "terraform", "baseline engine: terraform or tofu")
	cmd.Flags().StringVar(&fromPath, "from-bin", "", "baseline engine executable path")
	cmd.Flags().StringVar(&toName, "to-engine", "terraform", "candidate engine: terraform or tofu")
	cmd.Flags().StringVar(&toPath, "to-bin", "", "candidate engine executable path")
	addRehearsalCommonFlags(cmd, &configPath, &statePath, &inheritProxy, &allowRegistry, &format, &outputPath, &timeout)
	return cmd
}

func newRehearseMigrateCommand() *cobra.Command {
	var terraformPath, tofuPath, configPath, statePath, format, outputPath string
	var inheritProxy, allowRegistry bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Compare Terraform baseline behavior with OpenTofu",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if terraformPath == "" || tofuPath == "" || configPath == "" || statePath == "" {
				return errors.New("--terraform-bin, --tofu-bin, --config, and --state are required")
			}
			if format != "terminal" && format != "json" {
				return errors.New("--format must be terminal or json")
			}
			service := app.NewService(nil, timeout)
			result, err := service.RehearseMigrate(cmd.Context(), app.MigrateRehearsalRequest{
				TerraformPath: terraformPath, OpenTofuPath: tofuPath, ConfigPath: configPath, StatePath: statePath,
				InheritProxy: inheritProxy, AllowRegistryAccess: allowRegistry,
			})
			if err != nil {
				return fmt.Errorf("rehearse migrate: %w", err)
			}
			return writeRehearsal(cmd, result, format, outputPath)
		},
	}
	cmd.Flags().StringVar(&terraformPath, "terraform-bin", "", "Terraform executable path")
	cmd.Flags().StringVar(&tofuPath, "tofu-bin", "", "OpenTofu executable path")
	addRehearsalCommonFlags(cmd, &configPath, &statePath, &inheritProxy, &allowRegistry, &format, &outputPath, &timeout)
	return cmd
}

func addRehearsalCommonFlags(cmd *cobra.Command, configPath, statePath *string, inheritProxy, allowRegistry *bool, format, outputPath *string, timeout *time.Duration) {
	cmd.Flags().StringVar(configPath, "config", "", "local configuration directory")
	cmd.Flags().StringVar(statePath, "state", "", "local state file")
	cmd.Flags().BoolVar(inheritProxy, "inherit-proxy", false, "allow proxy variables in the reduced engine environment")
	cmd.Flags().BoolVar(allowRegistry, "allow-registry-access", true, "authorize provider/module registry access during copied init")
	cmd.Flags().StringVar(format, "format", "terminal", "output format: terminal or json")
	cmd.Flags().StringVarP(outputPath, "output", "o", "", "write the report to a file")
	cmd.Flags().DurationVar(timeout, "timeout", 5*time.Minute, "per-engine command timeout")
}

func writeRehearsal(cmd *cobra.Command, result model.RehearsalResult, format, outputPath string) error {
	writer, closeWriter, err := commandWriter(cmd.OutOrStdout(), outputPath)
	if err != nil {
		return err
	}
	defer closeWriter()
	if format == "json" {
		return report.JSON(writer, result)
	}
	return report.RehearsalTerminal(writer, result)
}
