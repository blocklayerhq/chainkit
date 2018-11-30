package cmd

import (
	"context"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

var cliCmd = &cobra.Command{
	Use:                "cli args ...",
	Short:              "Run a command from the application CLI",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}
		cli(p, args)
	},
}

func init() {
	cliCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(cliCmd)
}

func cli(p *project.Project, args []string) {
	ctx := context.Background()
	cmd := []string{
		"exec",
		"-it",
		p.Name,
		p.Binaries.CLI,
	}
	cmd = append(cmd, args...)
	if err := docker(ctx, p.RootDir, cmd...); err != nil {
		ui.Fatal("Failed to start the cli (is the application running?): %v", err)
	}
}
