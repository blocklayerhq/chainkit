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
	"sort"
	"sync"
	"time"

	"github.com/blocklayerhq/chainkit/ui"
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

// Server is the discovery server
type Server struct {
	root string
	port int
	node *core.IpfsNode
	api  iface.CoreAPI
	dht  *dht.IpfsDHT
}

// New returns a new discovery server
func New(root string, port int) *Server {
	return &Server{
		root: root,
		port: port,
	}
}

// Start starts the discovery server
func (s *Server) Start(ctx context.Context) error {
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
	s.dhtConnect(ctx)

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

// Stop must be called after start
func (s *Server) Stop() error {
	return s.node.Close()
}

// ListenAddresses returns the IPFS listening addresses for the server
func (s *Server) ListenAddresses() []string {
	ifaceAddrs, err := s.node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		panic(err)
	}

	var addrs []string
	for _, addr := range ifaceAddrs {
		addrs = append(addrs, addr.String())
	}
	sort.Sort(sort.StringSlice(addrs))
	return addrs
}

// AnnounceAddresses returns the announce addresses of IPFS
func (s *Server) AnnounceAddresses() []string {
	var addrs []string
	for _, addr := range s.node.PeerHost.Addrs() {
		addrs = append(addrs, addr.String())
	}
	sort.Sort(sort.StringSlice(addrs))
	return addrs
}

// Publish publishes chain information. Returns the chain ID.
func (s *Server) Publish(ctx context.Context, genesisPath, imagePath string) (string, error) {
	sandbox, err := ioutil.TempDir(os.TempDir(), "chainkit-network")
	if err != nil {
		return "", err
	}

	st, err := os.Stat(sandbox)
	if err != nil {
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

// Announce announces our presence as a network node.
func (s *Server) Announce(ctx context.Context, chainID string, peer *PeerInfo) error {
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

func (s *Server) dhtConnect(ctx context.Context) {
	connect := func(peerAddr string) error {
		addr, _ := iaddr.ParseString(peerAddr)
		peerinfo, _ := pstore.InfoFromP2pAddr(addr.Multiaddr())

		err := s.node.PeerHost.Connect(ctx, *peerinfo)
		if err != nil {
			ui.Error("%v", err)
			return err
		}
		ui.Verbose("Connection established with bootstrap node: %v", *peerinfo)
		return nil
	}

	wg := sync.WaitGroup{}
	for _, peerAddr := range bootstrapPeers {
		wg.Add(1)
		go func(peerAddr string) {
			defer wg.Done()
			connect(peerAddr)
		}(peerAddr)
	}
	wg.Wait()
}

// Join joins a network.
func (s *Server) Join(ctx context.Context, chainID string) (io.ReadCloser, io.ReadCloser, error) {
	genesisPath, err := iface.ParsePath(path.Join(chainID, "genesis.json"))
	if err != nil {
		return nil, nil, err
	}
	genesisFile, err := s.api.Unixfs().Get(ctx, genesisPath)
	if err != nil {
		return nil, nil, err
	}

	imagePath, err := iface.ParsePath(path.Join(chainID, "image.tgz"))
	imageFile, err := s.api.Unixfs().Get(ctx, imagePath)
	if err != nil {
		return nil, nil, err
	}

	// genesis, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	return nil, err
	// }

	return genesisFile, imageFile, nil
}

// PeerInfo contains information about one peer.
type PeerInfo struct {
	NodeID            string   `json:"node_id"`
	IP                []string `json:"ips"`
	TendermintP2PPort int      `json:"tendermint_p2p_port"`
}

// SearchPeers looks for peers in the network
func (s *Server) SearchPeers(ctx context.Context, chainID string) (<-chan *PeerInfo, error) {
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
					ui.Error("unable to make stream: %v", err)
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
