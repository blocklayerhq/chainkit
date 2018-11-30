package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/blocklayerhq/chainkit/discovery"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p, err := project.Load(getCwd(cmd))
		if err != nil {
			ui.Fatal("%v", err)
		}
		join, err := cmd.Flags().GetString("join")
		if err != nil {
			ui.Fatal("%v", err)
		}
		start(p, join)
	},
}

func init() {
	startCmd.Flags().String("cwd", ".", "specifies the current working directory")
	startCmd.Flags().String("join", "", "join a network")

	rootCmd.AddCommand(startCmd)
}

func startExplorer(ctx context.Context, p *project.Project) {
	cmd := []string{
		"run", "--rm",
		"-p", "8080:8080",
		"samalba/cosmos-explorer-localdev:latest",
	}
	if err := docker(ctx, p.RootDir, cmd...); err != nil {
		ui.Fatal("Failed to start the Explorer: %v", err)
	}
}

func start(p *project.Project, join string) {
	ctx, cancel := context.WithCancel(context.Background())
	ui.Info("Starting %s", p.Name)

	// Initialize if needed.
	if err := initialize(ctx, p); err != nil {
		ui.Fatal("Initialization failed: %v", err)
	}

	ipfsRoot := path.Join(p.RootDir, "data", fmt.Sprintf("%sd", p.Name), "ipfs")
	s := discovery.New(ipfsRoot)
	if err := s.Start(ctx); err != nil {
		ui.Fatal("%v", err)
	}
	defer s.Stop()

	for _, addr := range s.ListenAddresses() {
		ui.Verbose("IPFS Swarm listening on %s", addr)
	}

	for _, addr := range s.AnnounceAddresses() {
		ui.Verbose("IPFS Swarm announcing %s", addr)
	}

	genesisPath := path.Join(p.RootDir, "data", fmt.Sprintf("%sd", p.Name), "config", "genesis.json")

	// Start a network.
	if join == "" {
		chainID, err := s.Announce(ctx, genesisPath)
		if err != nil {
			ui.Fatal("%v", err)
		}
		ui.Success("Network is live at: %v", chainID)
	} else {
		ui.Info("Joining network %s", join)
		genesisData, peerCh, err := s.Join(ctx, join)
		if err != nil {
			ui.Fatal("%v", err)
		}

		ui.Success("Retrieved genesis data")

		if err := ioutil.WriteFile(genesisPath, genesisData, 0644); err != nil {
			ui.Fatal("Unable to write genesis file: %v", err)
		}

		peer := <-peerCh
		ui.Info("Peer: %v", peer.Addrs)
	}

	ui.Success("Application is live at:     %s", ui.Emphasize("http://localhost:26657/"))
	ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize("http://localhost:8080/"))
	defer cancel()
	go startExplorer(ctx, p)
	if err := dockerRun(ctx, p.RootDir, p.Image, "start"); err != nil {
		ui.Fatal("Failed to start the application: %v", err)
	}
}
