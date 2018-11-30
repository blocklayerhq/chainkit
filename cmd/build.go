package cmd

import (
	"context"

	"github.com/blocklayerhq/chainkit/builder"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
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

		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}

		b := builder.New(p)
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
