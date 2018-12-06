package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/blocklayerhq/chainkit/config"
	"github.com/blocklayerhq/chainkit/discovery"
	"github.com/blocklayerhq/chainkit/node"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := getCwd(cmd)
		p, err := project.Load(rootDir)
		if err != nil {
			ui.Fatal("%v", err)
		}

		ctx := context.Background()
		cfg := &config.Config{
			RootDir:        rootDir,
			PublishNetwork: true,
		}
		cfg.Ports, err = config.AllocatePorts()
		if err != nil {
			ui.Fatal("%v", err)
		}

		ui.Info("Starting %s", p.Name)
		d := discovery.New(cfg.IPFSDir(), cfg.Ports.IPFS)
		if err := d.Start(ctx); err != nil {
			ui.Fatal("Failed to initialize discovery: %v", err)
		}
		defer d.Stop()

		n := node.New(cfg, d)
		errCh := make(chan error)
		go func() {
			defer close(errCh)
			errCh <- n.Start(ctx, p, nil)
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

	rootCmd.AddCommand(startCmd)
}
