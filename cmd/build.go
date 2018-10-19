package cmd

import (
	"path/filepath"

	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)
		build(name, rootDir)
	},
}

func init() {
	buildCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(buildCmd)
}

func build(name, rootDir string) {
	ui.Info("Building %s", name)
	if err := docker(rootDir, "build", "-q", "-t", name, rootDir); err != nil {
		ui.Fatal("Failed to build the application: %v", err)
	}
}
