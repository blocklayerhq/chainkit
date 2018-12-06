package node

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/blocklayerhq/chainkit/config"
	"github.com/blocklayerhq/chainkit/discovery"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Node is a chainkit Node
type Node struct {
	config *config.Config

	parentCtx context.Context
	cancelCtx context.CancelFunc
	doneCh    chan struct{}

	server    *server
	discovery *discovery.Server
}

// New creates a new Node
func New(config *config.Config, discovery *discovery.Server) *Node {
	return &Node{
		config:    config,
		server:    newServer(config),
		discovery: discovery,
	}
}

// Stop stops the node and returns once fully stopped.
func (n *Node) Stop() {
	n.cancelCtx()
	<-n.doneCh
}

// Start starts the node. It will not return until it finishes
// starting.
func (n *Node) Start(ctx context.Context, p *project.Project, genesis []byte) error {
	n.parentCtx, n.cancelCtx = context.WithCancel(ctx)

	n.doneCh = make(chan struct{})
	defer close(n.doneCh)

	if err := n.init(ctx, p, genesis); err != nil {
		return err
	}

	for _, addr := range n.discovery.ListenAddresses() {
		ui.Verbose("IPFS Swarm listening on %s", addr)
	}

	for _, addr := range n.discovery.AnnounceAddresses() {
		ui.Verbose("IPFS Swarm announcing %s", addr)
	}

	// Create a network.
	chainID, err := n.createNetwork(n.parentCtx, p)
	if err != nil {
		return err
	}
	ui.Success("Network is live at: %v", chainID)

	if err := n.server.start(n.parentCtx, p); err != nil {
		return err
	}

	peer, err := n.server.peerInfo(n.parentCtx)
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(n.parentCtx)

	// Monitor the server
	g.Go(func() error {
		return n.server.wait()
	})

	// Start the explorer.
	g.Go(func() error {
		return startExplorer(gctx, n.config, p)
	})

	// Announce
	g.Go(func() error {
		return n.announce(gctx, chainID, peer)
	})

	// Discover Peers
	g.Go(func() error {
		return n.discoverPeers(gctx, chainID)
	})

	ui.Success("Node is up and running:     %s", peer.NodeID)
	ui.Success("Application is live at:     %s", ui.Emphasize(fmt.Sprintf("http://localhost:%d/", n.config.Ports.TendermintRPC)))
	ui.Success("Cosmos Explorer is live at: %s", ui.Emphasize(fmt.Sprintf("http://localhost:%d/?rpc_port=%d", n.config.Ports.Explorer, n.config.Ports.TendermintRPC)))

	return g.Wait()
}

// init initializes the server if needed and updates the runtime config.
func (n *Node) init(ctx context.Context, p *project.Project, genesis []byte) error {
	moniker, err := os.Hostname()
	if err != nil {
		return errors.Wrap(err, "unable to determine hostname")
	}

	// Initialize if needed.
	if err := initialize(ctx, n.config, p); err != nil {
		return errors.Wrap(err, "initialization failed")
	}

	err = updateConfig(
		n.config.ConfigFile(),
		map[string]string{
			// Set custom moniker. Needed to join nodes together.
			"moniker": fmt.Sprintf("%q", moniker),
			// Needed to join local/private networks.
			"addr_book_strict": "false",
			// Needed to enable dial_seeds
			"unsafe": "true",
			// Info logs are just too verbose.
			"log_level": fmt.Sprintf("%q", "*:error"),
		},
	)
	if err != nil {
		return err
	}

	if genesis == nil {
		return nil
	}

	if err := ioutil.WriteFile(n.config.GenesisPath(), genesis, 0644); err != nil {
		return errors.Wrap(err, "unable to overwrite genesis file")
	}

	return nil
}

func (n *Node) createNetwork(ctx context.Context, p *project.Project) (string, error) {
	f, err := ioutil.TempFile(os.TempDir(), "chainkit-image")
	if err != nil {
		return "", errors.Wrap(err, "unable to create temporary file")
	}
	if err := util.RunWithFD(ctx, os.Stdin, f, os.Stderr, "docker", "save", p.Image); err != nil {
		return "", errors.Wrap(err, "unable to save image")
	}
	f.Close()

	ui.Verbose("Image saved at %s", f.Name())

	chainID, err := n.discovery.Publish(ctx, n.config.ManifestPath(), n.config.GenesisPath(), f.Name())
	if err != nil {
		return "", errors.Wrap(err, "unable to create network")
	}

	return chainID, nil
}

func (n *Node) announce(ctx context.Context, chainID string, peer *discovery.PeerInfo) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := n.discovery.Announce(ctx, chainID, peer)
		if err == nil {
			ui.Success("Node successfully announced")
			return nil
		}
		ui.Error("Failed to announce: %v", err)
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *Node) discoverPeers(ctx context.Context, chainID string) error {
	ui.Info("Discovering peer nodes...")

	seenNodes := make(map[string]struct{})

	for {
		// Make sure the context was not cancelled.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		peerCh, err := n.discovery.SearchPeers(ctx, chainID)
		if err != nil {
			return err
		}

		for peer := range peerCh {
			if _, ok := seenNodes[peer.NodeID]; ok {
				continue
			}
			ui.Info("Discovered node %q", peer.NodeID)
			if err := n.server.dialSeeds(ctx, peer); err != nil {
				ui.Error("Failed to dial peer: %v", err)
				continue
			}

			seenNodes[peer.NodeID] = struct{}{}
		}

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
