package cmd

import (
	"context"
	"os"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}
		if err := initialize(ctx, p); err != nil {
			ui.Fatal("Initialization failed: %v", err)
		}

	},
}

func init() {
	initCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(initCmd)
}

func initialize(ctx context.Context, p *project.Project) error {
	_, err := os.Stat(p.StateDir())

	// Skip initialization if already initialized.
	if err == nil {
		return nil
	}

	// Make sure we got an ErrNotExist - fail otherwise.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	ui.Info("Generating configuration and genesis files")
	if err := dockerRun(ctx, p, "init"); err != nil {
		return err
	}
	if err := ui.Tree(p.StateDir(), nil); err != nil {
		return err
	}

	return nil
}
