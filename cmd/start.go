package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/blocklayerhq/chainkit/node"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

// ExplorerImage defines the container image to pull for running the Cosmos Explorer
const ExplorerImage = "samalba/cosmos-explorer-localdev:20181204"

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}
		chainID, err := cmd.Flags().GetString("join")
		if err != nil {
			ui.Fatal("%v", err)
		}

		ctx := context.Background()
		n := node.New(p)
		errCh := make(chan error)
		go func() {
			defer close(errCh)
			errCh <- n.Start(ctx, chainID)
		}()

		// Wait for the application to error out or the user to quit.
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,
			syscall.SIGTERM,
		)

		select {
		case err := <-errCh:
			if err != nil {
				ui.Error("%v", err)
			}
		case sig := <-c:
			ui.Info("Received signal %v, exiting", sig)
			n.Stop()
		}
	},
}

func init() {
	startCmd.Flags().String("cwd", ".", "specifies the current working directory")
	startCmd.Flags().String("join", "", "join a network")

	rootCmd.AddCommand(startCmd)
}
