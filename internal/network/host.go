package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-chat/internal/config"
	"go-chat/internal/logging"
	"go-chat/internal/storage"

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
	Host      host.Host
	Config    *config.NetworkConfig
	Logger    *logging.Logger
	Store     *storage.Store
	RefreshCh chan struct{}
	Ctx       context.Context
	Cancel    context.CancelFunc
	mdns      mdns.Service
}

func NewNode(privKey crypto.PrivKey, cfg *config.NetworkConfig, log *logging.Logger, store *storage.Store, refreshCh chan struct{}) (*Node, error) {
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
		opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(staticRelays))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}

	node := &Node{
		Host:      h,
		Config:    cfg,
		Logger:    log,
		Store:     store,
		RefreshCh: refreshCh,
		Ctx:       ctx,
		Cancel:    cancel,
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
			if _, err := node.Connect(ctx, addr); err != nil {
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

func (n *Node) Connect(ctx context.Context, addrStr string) (peer.ID, error) {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return "", fmt.Errorf("parse multiaddr: %w", err)
	}

	pi, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return "", fmt.Errorf("addr info: %w", err)
	}

	if err := n.Host.Connect(ctx, *pi); err != nil {
		return "", fmt.Errorf("connect: %w", err)
	}

	n.Logger.Info("connected to: %s", pi.ID.String())
	return pi.ID, nil
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

func (n *Node) Broadcast(msg *Message) {
	for _, p := range n.Host.Network().Peers() {
		s, err := n.Host.NewStream(n.Ctx, p, ProtocolID)
		if err != nil {
			n.Logger.Debug("broadcast to %s: %v", p.String(), err)
			continue
		}
		NewStreamHandler(n).SendMessage(s, msg)
		s.Close()
	}
}

func (n *Node) ConnectedPeers() []peer.ID {
	return n.Host.Network().Peers()
}

func (n *Node) ConnectedCount() int {
	return len(n.Host.Network().Peers())
}

func (n *Node) SyncWithPeer(ctx context.Context, peerID peer.ID) error {
	s, err := n.Host.NewStream(ctx, peerID, ProtocolID)
	if err != nil {
		return fmt.Errorf("open sync stream: %w", err)
	}
	defer s.Close()

	handler := NewStreamHandler(n)

	handler.SendMessage(s, &Message{
		Type:     "sync_request",
		SenderID: n.Host.ID().String(),
	})

	if n.Store != nil {
		orgs, _ := n.Store.ListOrganizations()
		for _, org := range orgs {
			handler.SendMessage(s, &Message{
				Type:      "sync_org",
				SenderID:  n.Host.ID().String(),
				OrgID:     org.OrgID,
				Content:   org.Name,
				Timestamp: org.CreatedAt.UnixMilli(),
			})
		}

		channels, _ := n.Store.ListAllChannels()
		for _, ch := range channels {
			if strings.HasPrefix(ch.ChannelID, "dm_") && !strings.Contains(ch.ChannelID, peerID.String()) {
				continue
			}
			handler.SendMessage(s, &Message{
				Type:        "sync_channel",
				SenderID:    n.Host.ID().String(),
				OrgID:       ch.OrgID,
				ChannelID:   ch.ChannelID,
				Content:     ch.Name,
				ChannelType: ch.ChannelType,
				Timestamp:   ch.CreatedAt.UnixMilli(),
			})
		}

		allPeers, _ := n.Store.ListPeers()
		for _, p := range allPeers {
			if p.DisplayName == "" || p.DisplayName == p.PeerID || p.DisplayName == "me" || strings.HasPrefix(p.DisplayName, "me_") {
				continue
			}
			handler.SendMessage(s, &Message{
				Type:      "sync_peer",
				SenderID:  p.PeerID,
				Content:   p.DisplayName,
				Timestamp: time.Now().UnixMilli(),
			})
		}

		myName := n.Host.ID().String()
		if identity, _ := n.Store.GetIdentity(); identity != nil {
			myName = identity.DisplayName
		}
		handler.SendMessage(s, &Message{
			Type:      "sync_peer",
			SenderID:  n.Host.ID().String(),
			Content:   myName,
			Timestamp: time.Now().UnixMilli(),
		})

		msgs, _ := n.Store.ListAllMessages(50)
		for _, msg := range msgs {
			handler.SendMessage(s, &Message{
				Type:       "message",
				SenderID:   msg.SenderPeerID,
				ChannelID:  msg.ChannelID,
				MessageID:  msg.MessageID,
				Content:    msg.Content,
				Timestamp:  msg.CreatedAt.UnixMilli(),
			})
		}
	}

	r := bufio.NewReader(s)
	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "sync_org":
			if n.Store != nil {
				handler.handleSyncOrg(&msg)
			}
		case "sync_channel":
			if n.Store != nil {
				handler.handleSyncChannel(&msg, s)
			}
		case "sync_peer":
			if n.Store != nil {
				handler.handleSyncPeer(&msg)
			}
		case "message":
			if n.Store != nil {
				handler.handleSyncMessage(&msg)
			}
		}
		if n.RefreshCh != nil {
			select {
			case n.RefreshCh <- struct{}{}:
			default:
			}
		}
	}

	n.Logger.Info("sync complete with %s", peerID.String())
	return nil
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
