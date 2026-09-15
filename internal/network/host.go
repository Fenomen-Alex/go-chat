package network

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-chat/internal/config"
	"go-chat/internal/crypto"
	"go-chat/internal/logging"
	"go-chat/internal/storage"

	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
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

const (
	maxConcurrentStreams = 256
	defaultMsgRate       = 20.0
	defaultMsgBurst      = 60.0
)

type Node struct {
	Host      host.Host
	Config    *config.NetworkConfig
	Logger    *logging.Logger
	Store     *storage.Store
	RefreshCh chan struct{}
	Ctx       context.Context
	Cancel    context.CancelFunc
	mdns      mdns.Service

	streams            atomic.Int32
	sessionKeys        map[string][]byte
	sessionMu          sync.RWMutex
	limiters           map[string]*msgLimiter
	limitersMu         sync.Mutex
	syncSessions       map[string]time.Time
	syncMu             sync.Mutex
	identityPrivateKey []byte
}

// msgLimiter is a small token bucket used to throttle inbound messages per
// remote peer so a connected peer cannot flood the local database or CPU.
type msgLimiter struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

func (l *msgLimiter) allow(rate, burst float64) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	elapsed := now.Sub(l.last).Seconds()
	l.tokens += elapsed * rate
	if l.tokens > burst {
		l.tokens = burst
	}
	l.last = now
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

func NewNode(privKey libp2pcrypto.PrivKey, cfg *config.NetworkConfig, log *logging.Logger, store *storage.Store, refreshCh chan struct{}) (*Node, error) {
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
		Host:               h,
		Config:             cfg,
		Logger:             log,
		Store:              store,
		RefreshCh:          refreshCh,
		Ctx:                ctx,
		Cancel:             cancel,
		sessionKeys:        make(map[string][]byte),
		limiters:           make(map[string]*msgLimiter),
		syncSessions:       make(map[string]time.Time),
		identityPrivateKey: ed25519PrivateKeyBytes(privKey),
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
	if !n.acquireStream() {
		n.Logger.Warn("too many concurrent streams from %s, rejecting", s.Conn().RemotePeer().String())
		_ = s.Close()
		return
	}
	defer n.releaseStream()
	n.Logger.Debug("new stream from: %s", s.Conn().RemotePeer().String())
	NewStreamHandler(n).Handle(s)
}

func (n *Node) acquireStream() bool {
	if n.streams.Load() >= maxConcurrentStreams {
		return false
	}
	n.streams.Add(1)
	return true
}

func (n *Node) releaseStream() {
	n.streams.Add(-1)
}

func (n *Node) allowMessage(peerID string) bool {
	n.limitersMu.Lock()
	l := n.limiters[peerID]
	if l == nil {
		l = &msgLimiter{tokens: defaultMsgBurst, last: time.Now()}
		n.limiters[peerID] = l
	}
	n.limitersMu.Unlock()
	return l.allow(defaultMsgRate, defaultMsgBurst)
}

func (n *Node) allowSyncSession(peerID string) bool {
	now := time.Now()
	n.syncMu.Lock()
	defer n.syncMu.Unlock()
	if last, ok := n.syncSessions[peerID]; ok && now.Sub(last) < 30*time.Second {
		return false
	}
	n.syncSessions[peerID] = now
	return true
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
	var wg sync.WaitGroup
	for _, p := range n.Host.Network().Peers() {
		if strings.HasPrefix(msg.ChannelID, "dm_") {
			if msg.SenderID != "" && !strings.Contains(msg.ChannelID, msg.SenderID) {
				continue
			}
			if !strings.Contains(msg.ChannelID, p.String()) {
				continue
			}
		}
		wg.Add(1)
		go func(peerID peer.ID) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(n.Ctx, 15*time.Second)
			defer cancel()
			s, err := n.Host.NewStream(ctx, peerID, ProtocolID)
			if err != nil {
				n.Logger.Debug("broadcast to %s: %v", peerID.String(), err)
				return
			}
			defer s.Close()
			NewStreamHandler(n).SendMessage(s, msg)
		}(p)
	}
	wg.Wait()
}

func (n *Node) ConnectedPeers() []peer.ID {
	return n.Host.Network().Peers()
}

func (n *Node) ConnectedCount() int {
	return len(n.Host.Network().Peers())
}

func (n *Node) GetSessionKey(peerID string) ([]byte, bool) {
	n.sessionMu.RLock()
	defer n.sessionMu.RUnlock()
	key, ok := n.sessionKeys[peerID]
	return key, ok
}

func (n *Node) SetSessionKey(peerID string, key []byte) {
	n.sessionMu.Lock()
	defer n.sessionMu.Unlock()
	n.sessionKeys[peerID] = key
}

func (n *Node) exchangeKeys(s network.Stream, peerID peer.ID, r *bufio.Reader) {
	ephPriv, ephPub, err := crypto.GenerateEphemeralKeypair()
	if err != nil {
		n.Logger.Warn("generate ephemeral keypair: %v", err)
		return
	}

	pubB64 := base64.StdEncoding.EncodeToString(ephPub)
	handler := NewStreamHandler(n)
	handler.SendMessage(s, &Message{
		Type:      "key_exchange",
		SenderID:  n.Host.ID().String(),
		Content:   pubB64,
		Timestamp: time.Now().UnixMilli(),
	})

	s.SetReadDeadline(time.Now().Add(10 * time.Second))
	ackData, err := readLine(r, 1<<20)
	if err != nil {
		n.Logger.Debug("key exchange ack not received from %s: %v", peerID.String(), err)
		return
	}

	var ackMsg Message
	if err := json.Unmarshal(ackData, &ackMsg); err != nil || ackMsg.Type != "key_exchange_ack" {
		n.Logger.Debug("invalid key exchange ack from %s", peerID.String())
		return
	}

	if !VerifyKeyExchangeSignature(n, &ackMsg, peerID.String()) {
		n.Logger.Debug("invalid key exchange ack signature from %s", peerID.String())
		return
	}

	if ackMsg.Timestamp != 0 {
		ackTime := time.UnixMilli(ackMsg.Timestamp)
		now := time.Now().UTC()
		if ackTime.Before(now.Add(-maxClockSkew)) || ackTime.After(now.Add(maxClockSkew)) {
			n.Logger.Debug("key exchange ack timestamp out of range from %s", peerID.String())
			return
		}
	}

	peerPub, err := base64.StdEncoding.DecodeString(ackMsg.Content)
	if err != nil || len(peerPub) != 32 {
		n.Logger.Debug("invalid key exchange pub key from %s", peerID.String())
		return
	}

	shared, err := crypto.ComputeSharedSecret(ephPriv, peerPub)
	if err != nil {
		n.Logger.Warn("compute shared secret: %v", err)
		return
	}

	salt := append(ephPub, peerPub...)
	key, err := crypto.DeriveSessionKeys(shared, salt)
	if err != nil {
		n.Logger.Warn("derive session keys: %v", err)
		return
	}

	n.SetSessionKey(peerID.String(), key)
	n.Logger.Debug("key exchange complete with %s", peerID.String())
}

func (n *Node) SyncWithPeer(ctx context.Context, peerID peer.ID) error {
	s, err := n.Host.NewStream(ctx, peerID, ProtocolID)
	if err != nil {
		return fmt.Errorf("open sync stream: %w", err)
	}
	defer s.Close()

	handler := NewStreamHandler(n)
	r := bufio.NewReader(s)

	n.exchangeKeys(s, peerID, r)

	handler.SendMessage(s, &Message{
		Type:      "sync_request",
		SenderID:  n.Host.ID().String(),
		Timestamp: time.Now().UnixMilli(),
	})

	if n.Store != nil {
		handler.sendSyncState(s)

		msgs, err := n.Store.ListAllMessages(10000)
		if err != nil {
			n.Logger.Warn("list all messages for sync: %v", err)
		}
		for _, msg := range msgs {
			if !handler.peerCanAccess(peerID.String(), msg.ChannelID) {
				continue
			}
			handler.SendMessage(s, &Message{
				Type:        "message",
				SenderID:    msg.SenderPeerID,
				ChannelID:   msg.ChannelID,
				MessageID:   msg.MessageID,
				Content:     msg.Content,
				ContentType: msg.ContentType,
				Timestamp:   msg.CreatedAt.UnixMilli(),
				Signature:   msg.Signature,
			})
		}
	}

	handler.SendMessage(s, &Message{
		Type:      "sync_complete",
		SenderID:  n.Host.ID().String(),
		Timestamp: time.Now().UnixMilli(),
	})

	timeoutCnt := 0
	for {
		s.SetReadDeadline(time.Now().Add(60 * time.Second))
		data, err := readLine(r, 1<<20)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				timeoutCnt++
				if timeoutCnt >= 3 {
					n.Logger.Debug("sync read idle timeout from %s, exiting", peerID.String())
					break
				}
				continue
			}
			break
		}
		timeoutCnt = 0

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		handler.DecryptMessage(&msg, peerID.String())

		switch msg.Type {
		case "sync_org":
			if n.Store != nil {
				handler.handleSyncOrg(&msg, peerID.String())
			}
		case "sync_channel":
			if n.Store != nil {
				handler.handleSyncChannel(&msg, s)
			}
		case "sync_peer":
			if n.Store != nil {
				handler.handleSyncPeer(&msg, peerID.String())
			}
		case "message":
			if n.Store != nil {
				handler.handleSyncMessage(&msg, peerID.String())
			}
		case "sync_channel_member":
			if n.Store != nil {
				handler.handleSyncChannelMember(&msg, peerID.String())
			}
		case "sync_complete":
			n.Logger.Debug("received sync_complete from %s", peerID.String())
			return nil
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

func ed25519PrivateKeyBytes(privKey libp2pcrypto.PrivKey) []byte {
	if ed25519Priv, ok := privKey.(*libp2pcrypto.Ed25519PrivateKey); ok {
		raw, err := ed25519Priv.Raw()
		if err == nil {
			return raw
		}
	}
	return nil
}

func VerifyKeyExchangeSignature(node *Node, msg *Message, peerID string) bool {
	if len(msg.Signature) == 0 {
		return false
	}
	pubKey, ok := GetPublicKeyFromPeerstore(node.Host, peerID)
	if !ok {
		return false
	}
	return VerifyMessageSignature(pubKey, msg)
}
