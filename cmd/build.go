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
		ctx := context.Background()

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			ui.Fatal("unable to resolve flag: %v", err)
		}
		noCache, err := cmd.Flags().GetBool("no-cache")
		if err != nil {
			ui.Fatal("unable to resolve flag: %v", err)
		}

		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)

		b := builder.New(rootDir, name)
		opts := builder.BuildOpts{
			Verbose: verbose,
			NoCache: noCache,
		}
		if err := b.Build(ctx, opts); err != nil {
			ui.Fatal("Failed to build the application: %v", err)
		}
	},
}

func init() {
	buildCmd.Flags().String("cwd", ".", "specifies the current working directory")
	buildCmd.Flags().BoolP("verbose", "v", false, "enable verbose mode")
	buildCmd.Flags().Bool("no-cache", false, "disable caching")

	rootCmd.AddCommand(buildCmd)
}
