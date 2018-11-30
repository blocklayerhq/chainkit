package discovery

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	pstore "github.com/ipsn/go-ipfs/gxlibs/github.com/libp2p/go-libp2p-peerstore"
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
	node *core.IpfsNode
	api  iface.CoreAPI
	dht  *dht.IpfsDHT
}

// New returns a new discovery server
func New(root string) *Server {
	return &Server{
		root: root,
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

	return nil
}

func (s *Server) ipfsInit() error {
	conf, err := config.Init(os.Stdout, nBitsForKeypairDefault)
	if err != nil {
		return err
	}
	port, err := getFreePort()
	if err != nil {
		return err
	}
	conf.Addresses.Swarm = []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
		fmt.Sprintf("/ip6/::/tcp/%d", port),
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

// Announce broadcasts chain information. Returns the chain ID.
func (s *Server) Announce(ctx context.Context, genesisPath string) (string, error) {
	st, err := os.Stat(genesisPath)
	if err != nil {
		return "", err
	}
	f, err := files.NewSerialFile("genesis.json", genesisPath, false, st)
	if err != nil {
		return "", err
	}

	p, err := s.api.Unixfs().Add(ctx, f)
	if err != nil {
		return "", err
	}

	go s.broadcast(ctx, p.Cid())

	return p.String(), nil
}

func (s *Server) dhtConnect(ctx context.Context) {
	for _, peerAddr := range bootstrapPeers {
		addr, _ := iaddr.ParseString(peerAddr)
		peerinfo, _ := pstore.InfoFromP2pAddr(addr.Multiaddr())

		err := s.node.PeerHost.Connect(ctx, *peerinfo)
		if err != nil {
			ui.Error("%v", err)
			continue
		}
		ui.Verbose("Connection established with bootstrap node: %v", *peerinfo)
	}
}

func (s *Server) broadcast(ctx context.Context, chainID cid.Cid) error {
	s.dhtConnect(ctx)
	return s.dht.Provide(ctx, chainID, true)
}

// Join joins a network.
func (s *Server) Join(ctx context.Context, chainID string) ([]byte, <-chan pstore.PeerInfo, error) {
	id, err := cid.Decode(filepath.Base(chainID))
	if err != nil {
		return nil, nil, err
	}

	fpath, err := iface.ParsePath(chainID)
	if err != nil {
		return nil, nil, err
	}

	file, err := s.api.Unixfs().Get(ctx, fpath)
	if err != nil {
		return nil, nil, err
	}

	genesis, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}

	return genesis, s.searchPeers(ctx, id), nil
}

func (s *Server) searchPeers(ctx context.Context, id cid.Cid) <-chan pstore.PeerInfo {
	s.dhtConnect(ctx)

	ch := make(chan pstore.PeerInfo)
	go func() {
		tctx, cancel := context.WithTimeout(ctx, 1*time.Minute)

		defer cancel()
		defer close(ch)

		for p := range s.dht.FindProvidersAsync(tctx, id, 10) {
			if p.ID != s.node.PeerHost.ID() && len(p.Addrs) > 0 {
				ch <- p
			}
		}
	}()

	return ch
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}
