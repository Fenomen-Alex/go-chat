package network

import (
	"context"
	"fmt"
	"time"

	"go-chat/internal/config"
	"go-chat/internal/logging"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const ProtocolID protocol.ID = "/go-chat/1.0.0"

type Node struct {
	Host   host.Host
	Config *config.NetworkConfig
	Logger *logging.Logger
	Ctx    context.Context
	Cancel context.CancelFunc
	mdns   mdns.Service
}

func NewNode(privKey crypto.PrivKey, cfg *config.NetworkConfig, log *logging.Logger) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	var staticRelays []peer.AddrInfo
	for _, relayAddr := range cfg.RelayPeers {
		if relayAddr == "" {
			continue
		}
		addr, err := multiaddr.NewMultiaddr(relayAddr)
		if err != nil {
			log.Warn("invalid relay addr %s: %v", relayAddr, err)
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Warn("parse relay addr %s: %v", relayAddr, err)
			continue
		}
		staticRelays = append(staticRelays, *pi)
	}

	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port),
		),
		libp2p.EnableHolePunching(),
		libp2p.NATPortMap(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	if len(staticRelays) > 0 {
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(staticRelays...))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}

	node := &Node{
		Host:   h,
		Config: cfg,
		Logger: log,
		Ctx:    ctx,
		Cancel: cancel,
	}

	h.SetStreamHandler(ProtocolID, node.handleStream)

	if cfg.EnableRelay {
		go node.startRelay()
	}

	for _, relayAddr := range cfg.RelayPeers {
		if relayAddr == "" {
			continue
		}
		go func(addr string) {
			if err := node.Connect(ctx, addr); err != nil {
				log.Warn("relay connect %s: %v", addr, err)
				return
			}
			log.Info("connected to relay: %s", addr)
		}(relayAddr)
	}

	if cfg.EnableMDNS {
		node.startMDNS()
	}

	log.Info("host created: %s", h.ID().String())
	for _, addr := range h.Addrs() {
		log.Info("listening on: %s/p2p/%s", addr, h.ID().String())
	}

	return node, nil
}

func (n *Node) startRelay() {
	_, err := relay.New(n.Host)
	if err != nil {
		n.Logger.Warn("relay setup: %v", err)
	}
}

func (n *Node) startMDNS() {
	service := mdns.NewMdnsService(n.Host, "go-chat", &mdnsNotifee{node: n})
	service.Start()
	n.mdns = service
}

func (n *Node) handleStream(s network.Stream) {
	n.Logger.Debug("new stream from: %s", s.Conn().RemotePeer().String())
	NewStreamHandler(n).Handle(s)
}

func (n *Node) Connect(ctx context.Context, addrStr string) error {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return fmt.Errorf("parse multiaddr: %w", err)
	}

	pi, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("addr info: %w", err)
	}

	if err := n.Host.Connect(ctx, *pi); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	n.Logger.Info("connected to: %s", pi.ID.String())
	return nil
}

func (n *Node) Disconnect(peerID peer.ID) error {
	if err := n.Host.Network().ClosePeer(peerID); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}
	return nil
}

func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.Host.Addrs()
}

func (n *Node) AddrInfo() *peer.AddrInfo {
	return &peer.AddrInfo{
		ID:    n.Host.ID(),
		Addrs: n.Host.Addrs(),
	}
}

func (n *Node) MyAddr() string {
	id := n.Host.ID().String()
	for _, addr := range n.Host.Addrs() {
		return fmt.Sprintf("%s/p2p/%s", addr, id)
	}
	return ""
}

func (n *Node) AllAddrs() []string {
	id := n.Host.ID().String()
	var out []string
	for _, addr := range n.Host.Addrs() {
		out = append(out, fmt.Sprintf("%s/p2p/%s", addr, id))
	}
	return out
}

func (n *Node) Close() error {
	n.Cancel()
	if n.mdns != nil {
		n.mdns.Close()
	}
	return n.Host.Close()
}

func (n *Node) ReconnectWithBackoff(ctx context.Context, pi peer.AddrInfo) error {
	const (
		maxAttempts = 5
		baseDelay   = 2 * time.Second
		maxDelay    = 60 * time.Second
	)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := n.Host.Connect(ctx, pi)
		if err == nil {
			n.Logger.Info("reconnected to: %s", pi.ID.String())
			return nil
		}

		n.Logger.Warn("reconnect attempt %d/%d failed: %v", attempt+1, maxAttempts, err)

		if attempt < maxAttempts-1 {
			delay := time.Duration(min(1<<uint(attempt), 30)) * baseDelay
			if delay > maxDelay {
				delay = maxDelay
			}
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("reconnect failed after %d attempts", maxAttempts)
}
