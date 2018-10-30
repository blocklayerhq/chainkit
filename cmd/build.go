package cmd

import (
	"context"
	"path/filepath"

	"github.com/blocklayerhq/chainkit/pkg/builder"
	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			ui.Fatal("unable to resolve flag: %v", err)
		}
		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)
		build(name, rootDir, verbose)
	},
}

func init() {
	buildCmd.Flags().String("cwd", ".", "specifies the current working directory")
	buildCmd.Flags().BoolP("verbose", "v", false, "enable verbose mode")

	rootCmd.AddCommand(buildCmd)
}

func build(name, rootDir string, verbose bool) {
	ctx := context.Background()
	ui.Info("Building %s", name)
	b := builder.New(rootDir, name)
	opts := builder.BuildOpts{
		Verbose: verbose,
	}
	if err := b.Build(ctx, opts); err != nil {
		ui.Fatal("Failed to build the application: %v", err)
	}
	ui.Success("Build successfull")
}
