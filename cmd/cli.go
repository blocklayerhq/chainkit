package cmd

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
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

func getContainerID(ctx context.Context, p *project.Project) string {
	cmd := []string{
		"ps", "-q",
		"-f", "label=chainkit.cosmos.daemon",
		"-f", "label=chainkit.project=" + p.Name,
	}
	var b bytes.Buffer
	bwriter := bufio.NewWriter(&b)
	if err := util.RunWithFD(ctx, os.Stdin, bwriter, os.Stderr, "docker", cmd...); err != nil {
		ui.Fatal("Failed to start the cli (can't find the daemon container, is the application running?): %v", err)
		return ""
	}
	// FIXME: if there are multiple chainkit containers running, only the first one will be detected.
	containerID := strings.Split(b.String(), "\n")[0]
	return containerID
}

func cli(p *project.Project, args []string) {
	ctx := context.Background()
	containerID := getContainerID(ctx, p)
	cmd := []string{
		"exec",
		"-it",
		containerID,
		p.Binaries.CLI,
	}
	cmd = append(cmd, args...)
	if err := util.Run(ctx, "docker", cmd...); err != nil {
		ui.Fatal("Failed to start the cli (is the application running?): %v", err)
	}
}
