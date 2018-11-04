package cmd

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)
		start(name, rootDir)
	},
}

func init() {
	startCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(startCmd)
}

func startExplorer(ctx context.Context, name, rootDir string) {
	cmd := []string{
		"run", "--rm",
		"-p", "8080:8080",
		"samalba/cosmos-explorer-localdev:latest",
	}
	if err := docker(ctx, rootDir, cmd...); err != nil {
		ui.Fatal("Failed to start the Explorer: %v", err)
	}
}

func start(name, rootDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.Info("Starting %s", name)

	// Initialize if needed.
	if _, err := os.Stat(path.Join(rootDir, "data")); os.IsNotExist(err) {
		ui.Info("Generating configuration and gensis")
		if err := dockerRun(ctx, rootDir, name, "init"); err != nil {
			ui.Fatal("Initialization failed: %v", err)
		}
		if err := ui.Tree(path.Join(rootDir, "data"), nil); err != nil {
			ui.Fatal("%v", err)
		}
	}

	ui.Success("Application is live at:     %s", ui.Emphasize("http://localhost:26657/"))
	ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize("http://localhost:8080/"))
	defer cancel()
	go startExplorer(ctx, name, rootDir)
	if err := dockerRun(ctx, rootDir, name, "start"); err != nil {
		ui.Fatal("Failed to start the application: %v", err)
	}
}
