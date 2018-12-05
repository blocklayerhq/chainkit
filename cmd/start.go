package cmd

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/blocklayerhq/chainkit/discovery"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
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

func startExplorer(ctx context.Context, p *project.Project) error {
	cmd := []string{
		"run", "--rm",
		"--name", fmt.Sprint(p.Image + "-explorer"),
		"-p", fmt.Sprintf("%d:8080", p.Ports.Explorer),
		ExplorerImage,
	}
	if err := docker(ctx, p, cmd...); err != nil {
		return errors.Wrap(err, "failed to start the explorer")
	}
	return nil
}

func startServer(ctx context.Context, p *project.Project) error {
	if err := dockerRun(ctx, p, "start"); err != nil {
		return errors.Wrap(err, "failed to start the application")
	}
	return nil
}

func start(p *project.Project, chainID string) {
	ctx := context.Background()
	ui.Info("Starting %s", p.Name)

	// Initialize if needed.
	if err := initialize(ctx, p); err != nil {
		ui.Fatal("Initialization failed: %v", err)
	}

	d := discovery.New(p.IPFSDir(), p.Ports.IPFS)
	if err := d.Start(ctx); err != nil {
		ui.Fatal("%v", err)
	}
	defer d.Stop()

	for _, addr := range d.ListenAddresses() {
		ui.Verbose("IPFS Swarm listening on %s", addr)
	}

	for _, addr := range d.AnnounceAddresses() {
		ui.Verbose("IPFS Swarm announcing %s", addr)
	}

	// Start a network.
	if chainID == "" {
		f, err := ioutil.TempFile(os.TempDir(), "chainkit-image")
		if err != nil {
			ui.Fatal("Unable to create temporary file: %v", err)
		}
		if err := runWithFD(ctx, p, os.Stdin, f, os.Stderr, "docker", "save", p.Image); err != nil {
			ui.Fatal("Unable to save image: %v", err)
		}
		f.Close()

		ui.Verbose("Image saved at %s", f.Name())

		chainID, err = d.Publish(ctx, p.GenesisPath(), f.Name())
		if err != nil {
			ui.Fatal("%v", err)
		}
		ui.Success("Network is live at: %v", chainID)
	} else {
		ui.Info("Joining network %s", chainID)
		genesis, image, err := d.Join(ctx, chainID)
		if err != nil {
			ui.Fatal("%v", err)
		}
		defer genesis.Close()
		defer image.Close()

		f, err := os.OpenFile(p.GenesisPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			ui.Fatal("unable to overwrite genesis file: %v", err)
		}

		if _, err := io.Copy(f, genesis); err != nil {
			ui.Fatal("unable to write genesis: %v", err)
		}

		ui.Success("Retrieved genesis data")

		if err := runWithFD(ctx, p, image, os.Stdout, os.Stderr, "docker", "load"); err != nil {
			ui.Fatal("unable to load image: %v", err)
		}
	}

	cctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	errCh := make(chan error, 2)

	// Start the node.
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- startServer(cctx, p)
	}()

	// Start the explorer.
	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- startExplorer(cctx, p)
	}()

	go func() {
		rpc := client.NewHTTP(
			fmt.Sprintf("http://localhost:%d", p.Ports.TendermintRPC),
			fmt.Sprintf("http://localhost:%d/websocket", p.Ports.TendermintRPC),
		)

		var (
			err    error
			status *ctypes.ResultStatus
		)
		// Wait for the node to come up.
		for {
			status, err = rpc.Status()
			if err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}

		ui.Info("Node %q is up and running", status.NodeInfo.ID)
		ui.Success("Application is live at:     %s", ui.Emphasize(fmt.Sprintf("http://localhost:%d/", p.Ports.TendermintRPC)))
		ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize(fmt.Sprintf("http://localhost:%d/?rpc_port=%d", p.Ports.Explorer, p.Ports.TendermintRPC)))

		peer := &discovery.PeerInfo{
			NodeID:            string(status.NodeInfo.ID),
			TendermintP2PPort: p.Ports.TendermintP2P,
		}

		// Announce
		go func() {
			defer wg.Done()
			for {
				err := d.Announce(ctx, chainID, peer)
				if err == nil {
					ui.Success("Node successfully announced")
					break
				}
				ui.Error("Failed to announce: %v", err)
				time.Sleep(5 * time.Second)
			}
		}()

		// Search Peers
		go func() {
			defer wg.Done()
			for {
				peerCh, err := d.SearchPeers(ctx, chainID)
				if err != nil {
					ui.Fatal("%v", err)
				}
				nodes := make(map[string]struct{})
				for peer := range peerCh {
					if _, ok := nodes[peer.NodeID]; ok {
						continue
					}
					ui.Info("Discovered node %q", peer.NodeID)
					if err := dialSeeds(ctx, p, peer); err != nil {
						ui.Error("Failed to dial peer: %v", err)
						continue
					}

					nodes[peer.NodeID] = struct{}{}
				}

				time.Sleep(5 * time.Second)
			}
		}()
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
	}

	// Stop all processes and wait for completion.
	cancel()
	wg.Wait()
}

func dialSeeds(ctx context.Context, p *project.Project, peer *discovery.PeerInfo) error {
	seeds := []string{}
	for _, ip := range peer.IP {
		seeds = append(seeds, fmt.Sprintf("\"%s@%s:%d\"", peer.NodeID, ip, peer.TendermintP2PPort))
	}
	seedString := fmt.Sprintf("[%s]", strings.Join(seeds, ","))

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("http://localhost:%d/dial_seeds?seeds=%s",
			p.Ports.TendermintRPC,
			url.QueryEscape(seedString),
		),
		nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("requested failed with code %d", resp.StatusCode)
	}
	return nil
}
