package command

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
	localserver "github.com/saadabdullaah/planparity/internal/server"
	"github.com/spf13/cobra"
)

func newUICommand() *cobra.Command {
	var port int
	var open bool
	var terraformPath, tofuPath, pathRoot string
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Start the authenticated localhost web interface",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			service := app.NewService(nil, 5*time.Minute)
			server, err := localserver.New(service, localserver.Options{
				Host: "127.0.0.1", Port: port, TerraformPath: terraformPath, OpenTofuPath: tofuPath, AllowedPathRoot: pathRoot,
			})
			if err != nil {
				return err
			}
			if err := server.Listen(); err != nil {
				return fmt.Errorf("listen on localhost: %w", err)
			}
			url := server.URL()
			fmt.Fprintf(cmd.OutOrStdout(), "PlanParity UI\n\nListening: %s\nPath root: %s\nExecution mode: LOCAL-RESTRICTED\nResults: memory-only\n\nPress Ctrl+C to stop.\n", url, server.PathRoot())
			if open {
				if err := openBrowser(url); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Browser could not be opened automatically: %v\n", err)
				}
			}
			return server.Serve(cmd.Context())
		},
	}
	cmd.Flags().IntVar(&port, "port", 8765, "localhost TCP port (0 selects an available port)")
	cmd.Flags().BoolVar(&open, "open-browser", true, "open the authenticated local URL")
	cmd.Flags().StringVar(&terraformPath, "terraform-bin", "", "explicit Terraform executable path for system detection")
	cmd.Flags().StringVar(&tofuPath, "tofu-bin", "", "explicit OpenTofu executable path for system detection")
	cmd.Flags().StringVar(&pathRoot, "path-root", ".", "limit browser state and configuration paths to this directory tree")
	return cmd
}

func openBrowser(url string) error {
	command, err := browserCommand(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return command.Start()
}

func browserCommand(goos, url string) (*exec.Cmd, error) {
	switch goos {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url), nil
	case "linux":
		return exec.Command("xdg-open", url), nil
	default:
		return nil, fmt.Errorf("automatic browser opening is unsupported on %s", goos)
	}
}
