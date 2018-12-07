package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/ipsn/go-ipfs/core"
	"github.com/ipsn/go-ipfs/core/coreapi"
	iface "github.com/ipsn/go-ipfs/core/coreapi/interface"
	cid "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-cid"
	iaddr "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-addr"
	config "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-config"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-files"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-kad-dht"
	net "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-net"
	pstore "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-peerstore"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multiaddr"
	"github.com/ipsn/go-ipfs/plugin/loader"
	"github.com/ipsn/go-ipfs/repo/fsrepo"
	"github.com/pkg/errors"
)

const (
	nBitsForKeypairDefault = 4096
)

var (
	// IPFS bootstrap nodes. Used to find other peers in the network.
	bootstrapPeers = []string{
		"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
		"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
		"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
		"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
	}
)

// PeerInfo contains information about one peer.
type PeerInfo struct {
	NodeID            string   `json:"node_id"`
	IP                []string `json:"ips"`
	TendermintP2PPort int      `json:"tendermint_p2p_port"`
}

// Server is the discovery server
type Server struct {
	root string
	port int
	node *core.IpfsNode

	dht         *dht.IpfsDHT
	connectedCh chan (struct{})

	api iface.CoreAPI
}

// New returns a new discovery server
func New(root string, port int) *Server {
	return &Server{
		root:        root,
		port:        port,
		connectedCh: make(chan struct{}),
	}
}

// Stop must be called after start
func (s *Server) Stop() error {
	return s.node.Close()
}

// Start starts the discovery server
func (s *Server) Start(ctx context.Context) error {
	ui.Info("Initializing node...")

	daemonLocked, err := fsrepo.LockedByOtherProcess(s.root)
	if err != nil {
		return err
	}
	if daemonLocked {
		return fmt.Errorf("another instance is already accessing %q", s.root)
	}

	plugins := path.Join(s.root, "plugins")
	if _, err = loader.LoadPlugins(plugins); err != nil {
		return err
	}

	if !fsrepo.IsInitialized(s.root) {
		if err := s.ipfsInit(); err != nil {
			return err
		}
	}

	repo, err := fsrepo.Open(s.root)
	if err != nil {
		return err
	}

	err = repo.SetConfigKey("Addresses.Swarm", []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", s.port),
		fmt.Sprintf("/ip6/::/tcp/%d", s.port),
	})
	if err != nil {
		return err
	}

	s.node, err = core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Repo:   repo,
	})
	if err != nil {
		return err
	}

	s.api = coreapi.NewCoreAPI(s.node)
	s.dht, err = dht.New(ctx, s.node.PeerHost)
	if err != nil {
		return err
	}

	go s.dhtConnect(ctx)

	return nil
}

func (s *Server) ipfsInit() error {
	conf, err := config.Init(os.Stdout, nBitsForKeypairDefault)
	if err != nil {
		return err
	}
	conf.Addresses.API = []string{}
	conf.Addresses.Gateway = []string{}

	return fsrepo.Init(s.root, conf)
}

func (s *Server) dhtConnect(ctx context.Context) {
	defer close(s.connectedCh)
	for _, peerAddr := range bootstrapPeers {
		addr, _ := iaddr.ParseString(peerAddr)
		peerinfo, _ := pstore.InfoFromP2pAddr(addr.Multiaddr())

		err := s.node.PeerHost.Connect(ctx, *peerinfo)
		if err != nil {
			ui.Error("Connection with bootstrap node %v failed: %v", *peerinfo, err)
			continue
		}
	}
}

// Publish publishes chain information. Returns the chain ID.
func (s *Server) Publish(ctx context.Context, manifestPath, genesisPath, imagePath string) (string, error) {
	sandbox, err := ioutil.TempDir(os.TempDir(), "chainkit-network")
	if err != nil {
		return "", err
	}

	st, err := os.Stat(sandbox)
	if err != nil {
		return "", err
	}

	if err := os.Link(manifestPath, path.Join(sandbox, "chainkit.yml")); err != nil {
		return "", err
	}
	if err := os.Link(genesisPath, path.Join(sandbox, "genesis.json")); err != nil {
		return "", err
	}
	if err := os.Link(imagePath, path.Join(sandbox, "image.tgz")); err != nil {
		return "", err
	}

	f, err := files.NewSerialFile("network", sandbox, false, st)
	if err != nil {
		return "", err
	}

	p, err := s.api.Unixfs().Add(ctx, f)
	if err != nil {
		return "", err
	}

	return p.String(), nil
}

// Join joins a network.
func (s *Server) Join(ctx context.Context, chainID, manifestPath string) (*project.Project, []byte, error) {
	manifest, genesis, image, err := s.getNetworkMetadata(ctx, chainID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to join network")
	}
	defer manifest.Close()
	defer genesis.Close()
	defer image.Close()

	genesisData, err := ioutil.ReadAll(genesis)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to read genesis file")
	}

	f, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to overwrite manifest file")
	}
	if _, err := io.Copy(f, manifest); err != nil {
		return nil, nil, errors.Wrap(err, "unable to write manifest")
	}

	if err := util.DockerLoad(ctx, image); err != nil {
		return nil, nil, errors.Wrap(err, "unable to load image")
	}

	p, err := project.Parse(manifestPath)
	if err != nil {
		return nil, nil, err
	}

	return p, genesisData, nil
}

// Join joins a network.
func (s *Server) getNetworkMetadata(ctx context.Context, chainID string) (io.ReadCloser, io.ReadCloser, io.ReadCloser, error) {
	manifestPath, err := iface.ParsePath(path.Join(chainID, "chainkit.yml"))
	if err != nil {
		return nil, nil, nil, err
	}
	manifestFile, err := s.api.Unixfs().Get(ctx, manifestPath)
	if err != nil {
		return nil, nil, nil, err
	}

	genesisPath, err := iface.ParsePath(path.Join(chainID, "genesis.json"))
	if err != nil {
		return nil, nil, nil, err
	}
	genesisFile, err := s.api.Unixfs().Get(ctx, genesisPath)
	if err != nil {
		return nil, nil, nil, err
	}

	imagePath, err := iface.ParsePath(path.Join(chainID, "image.tgz"))
	imageFile, err := s.api.Unixfs().Get(ctx, imagePath)
	if err != nil {
		return nil, nil, nil, err
	}

	return manifestFile, genesisFile, imageFile, nil
}

// Announce announces our presence as a network node.
func (s *Server) Announce(ctx context.Context, chainID string, peer *PeerInfo) error {
	// Wait for the DHT to be connected before searching.
	<-s.connectedCh

	id, err := cid.Decode(filepath.Base(chainID))
	if err != nil {
		return err
	}

	s.node.PeerHost.SetStreamHandler("/chainkit/0.1.0", func(stream net.Stream) {
		defer stream.Close()
		enc := json.NewEncoder(stream)
		if err := enc.Encode(peer); err != nil {
			ui.Error("failed to encode: %v", err)
			return
		}
	})

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.dht.Provide(cctx, id, true); err != nil {
		return err
	}
	return nil
}

// Peers looks for peers in the network
func (s *Server) Peers(ctx context.Context, chainID string) (<-chan *PeerInfo, error) {
	// Wait for the DHT to be connected before searching.
	<-s.connectedCh

	id, err := cid.Decode(filepath.Base(chainID))
	if err != nil {
		return nil, err
	}

	ch := make(chan *PeerInfo)
	go func() {
		tctx, cancel := context.WithTimeout(ctx, 10*time.Second)

		defer cancel()
		defer close(ch)

		peers := s.dht.FindProvidersAsync(tctx, id, 10)
		for p := range peers {
			if p.ID != s.node.PeerHost.ID() && len(p.Addrs) > 0 {
				stream, err := s.node.PeerHost.NewStream(ctx, p.ID, "/chainkit/0.1.0")
				if err != nil {
					continue
				}
				dec := json.NewDecoder(stream)
				peer := &PeerInfo{}
				if err := dec.Decode(peer); err != nil {
					ui.Error("failed to decode: %v", err)
					continue
				}

				if peer.IP == nil {
					peer.IP = []string{}
				}
				for _, addr := range p.Addrs {
					v, err := addr.ValueForProtocol(multiaddr.P_IP4)
					if err != nil || v == "" {
						continue
					}

					peer.IP = append(peer.IP, v)
				}

				ch <- peer
			}
		}
	}()

	return ch, nil
}
