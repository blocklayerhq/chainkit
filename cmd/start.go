package cmd

import (
	"context"

	"github.com/blocklayerhq/chainkit/pkg/project"
	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}
		start(p)
	},
}

func init() {
	startCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(startCmd)
}

func startExplorer(ctx context.Context, p *project.Project) {
	cmd := []string{
		"run", "--rm",
		"-p", "8080:8080",
		"samalba/cosmos-explorer-localdev:latest",
	}
	if err := docker(ctx, p.RootDir, cmd...); err != nil {
		ui.Fatal("Failed to start the Explorer: %v", err)
	}
}

func start(p *project.Project) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.Info("Starting %s", p.Name)

	// Initialize if needed.
	if err := initialize(ctx, p); err != nil {
		ui.Fatal("Initialization failed: %v", err)
	}

	ui.Success("Application is live at:     %s", ui.Emphasize("http://localhost:26657/"))
	ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize("http://localhost:8080/"))
	defer cancel()
	go startExplorer(ctx, p)
	if err := dockerRun(ctx, p.RootDir, p.Name, "start"); err != nil {
		ui.Fatal("Failed to start the application: %v", err)
	}
}
