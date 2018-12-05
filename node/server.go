package node

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/blocklayerhq/chainkit/discovery"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/tendermint/tendermint/rpc/client"
)

type server struct {
	p     *project.Project
	errCh chan error
	rpc   *client.HTTP
}

func newServer(p *project.Project) *server {
	return &server{
		p:     p,
		errCh: make(chan error),
		rpc: client.NewHTTP(
			fmt.Sprintf("http://localhost:%d", p.Ports.TendermintRPC),
			fmt.Sprintf("http://localhost:%d/websocket", p.Ports.TendermintRPC),
		),
	}
}

// waitReady blocks until the node is ready.
func (s *server) waitReady(ctx context.Context) error {
	for {
		_, err := s.rpc.Status()
		if err == nil {
			return nil
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// start starts the server and returns when it's up and running.
func (s *server) start(ctx context.Context) error {
	// Spin the server on the background.
	go func() {
		defer close(s.errCh)
		// TODO: Leaving disabled for now as it helps finding leaks.
		// defer func() {
		// 	// Failsafe: Sometimes, if we stop a `docker run --rm`, it will leave
		// 	// the container behind.
		// 	util.Run(ctx, "docker", "rm", "-f", p.Name)
		// }()

		s.errCh <- util.DockerRun(ctx, s.p, "start")
	}()

	// Wait for the server to be ready.
	waitCh := make(chan error)
	go func() {
		defer close(waitCh)
		waitCh <- s.waitReady(ctx)
	}()

	// Now we wait for the server to come up, or to error out.
	select {
	case err := <-s.errCh:
		return err
	case err := <-waitCh:
		if err != nil {
			return err
		}
	}

	return nil
}

// wait waits until the server stops.
func (s *server) wait() error {
	return <-s.errCh
}

// peerInfo retrieves PeerInfo from the underlying node
func (s *server) peerInfo(ctx context.Context) (*discovery.PeerInfo, error) {
	status, err := s.rpc.Status()
	if err != nil {
		return nil, err
	}

	return &discovery.PeerInfo{
		NodeID:            string(status.NodeInfo.ID),
		TendermintP2PPort: s.p.Ports.TendermintP2P,
	}, nil
}

// dialSeeds will add the given seeds to the underlying node.
func (s *server) dialSeeds(ctx context.Context, peer *discovery.PeerInfo) error {
	seeds := []string{}
	for _, ip := range peer.IP {
		seeds = append(seeds, fmt.Sprintf("\"%s@%s:%d\"", peer.NodeID, ip, peer.TendermintP2PPort))
	}
	seedString := fmt.Sprintf("[%s]", strings.Join(seeds, ","))

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("http://localhost:%d/dial_seeds?seeds=%s",
			s.p.Ports.TendermintRPC,
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
