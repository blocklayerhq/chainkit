package cmd

import (
	"context"
	"path/filepath"

	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var cliCmd = &cobra.Command{
	Use:   "cli args ...",
	Short: "Run a command from the application CLI",
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)
		cli(name, rootDir, args)
	},
}

func init() {
	cliCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(cliCmd)
}

func cli(name, rootDir string, args []string) {
	ctx := context.Background()
	cmd := []string{
		"exec",
		"-it",
		name,
		name + "cli",
	}
	cmd = append(cmd, args...)
	if err := docker(ctx, rootDir, cmd...); err != nil {
		ui.Fatal("Failed to start the cli (is the application running?): %v", err)
	}
}
