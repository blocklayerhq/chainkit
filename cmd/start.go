package cmd

import (
	"context"
	"path/filepath"
	"time"

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
		"-e", "NGINX_PORT=8080",
		"--network", "container:" + name,
		"samalba/cosmos-explorer-localhost" + ":latest",
	}
	go func() {
		// FIXME: wait for the container to be created
		time.Sleep(3 * time.Second)
		if err := docker(ctx, rootDir, cmd...); err != nil {
			ui.Fatal("Failed to start the Explorer: %v", err)
		}
	}()
}

func start(name, rootDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.Info("Starting %s", name)
	ui.Success("Application is live at:     %s", ui.Emphasize("http://localhost:26657/"))
	ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize("http://localhost:8080/"))
	defer cancel()
	startExplorer(ctx, name, rootDir)
	if err := dockerRun(ctx, rootDir, name, "start"); err != nil {
		ui.Fatal("Failed to start the application: %v", err)
	}
}
