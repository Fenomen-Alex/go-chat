package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-chat/internal/channel"
	"go-chat/internal/config"
	"go-chat/internal/crypto"
	"go-chat/internal/discovery"
	"go-chat/internal/logging"
	"go-chat/internal/network"
	"go-chat/internal/organization"
	"go-chat/internal/peermgr"
	"go-chat/internal/protocol"
	"go-chat/internal/storage"

	lp2ppeer "github.com/libp2p/go-libp2p/core/peer"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
)

type App struct {
	Config  *config.Config
	Logger  *logging.Logger
	Store   *storage.Store
	Node    *network.Node

	peerManager    *peermgr.Manager
	orgManager     *organization.Manager
	channelManager *channel.Manager
	discoverer     *discovery.Discoverer

	identity   *storage.Identity
	libp2pKey  libp2pCrypto.PrivKey
	activeChan string
}

func New(cfgPath string) (*App, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := logging.New(cfg.Logging.Level, cfg.Logging.File, cfg.Logging.Rotate)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	store, err := storage.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}

	app := &App{
		Config: cfg,
		Logger: log,
		Store:  store,
	}

	app.peerManager = peermgr.NewManager(store)
	app.orgManager = organization.NewManager(store)
	app.channelManager = channel.NewManager(store)

	if err := app.loadOrCreateIdentity(); err != nil {
		return nil, fmt.Errorf("setup identity: %w", err)
	}

	if err := app.startNetworking(); err != nil {
		return nil, fmt.Errorf("start networking: %w", err)
	}

	return app, nil
}

func (a *App) loadOrCreateIdentity() error {
	id, err := a.Store.GetIdentity()
	if err != nil {
		return fmt.Errorf("get identity: %w", err)
	}

	if id != nil {
		a.identity = id
		priv, err := libp2pCrypto.UnmarshalPrivateKey(id.PrivateKey)
		if err != nil {
			return fmt.Errorf("unmarshal private key: %w", err)
		}
		a.libp2pKey = priv
		a.Logger.Info("loaded identity: %s (%s)", id.DisplayName, id.PeerID)
		return nil
	}

	keypair, err := crypto.GenerateIdentityKeypair()
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}

	privKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(keypair.PrivateKey)
	if err != nil {
		return fmt.Errorf("unmarshal ed25519: %w", err)
	}

	pid, err := lp2ppeer.IDFromPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("peer id from key: %w", err)
	}
	peerID := pid.String()

	privBytes, err := libp2pCrypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}

	displayName := a.Config.Identity.DisplayName
	if displayName == "" {
		hostname, _ := os.Hostname()
		displayName = fmt.Sprintf("user-%s", hostname)
	}

	now := time.Now().UTC()
	id = &storage.Identity{
		DisplayName: displayName,
		PeerID:      peerID,
		PrivateKey:  privBytes,
		PublicKey:   keypair.PublicKey,
		AvatarColor: "#5865F2",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := a.Store.SaveIdentity(id); err != nil {
		return fmt.Errorf("save identity: %w", err)
	}

	a.identity = id
	a.libp2pKey = privKey
	a.Logger.Info("created identity: %s (%s)", displayName, peerID)
	return nil
}

func (a *App) startNetworking() error {
	node, err := network.NewNode(a.libp2pKey, &a.Config.Network, a.Logger, a.Store)
	if err != nil {
		return fmt.Errorf("create network node: %w", err)
	}
	a.Node = node

	a.discoverer = discovery.New(node.Host, &a.Config.Network, a.Logger)

	ctx := context.Background()
	if err := a.discoverer.Bootstrap(ctx); err != nil {
		a.Logger.Warn("bootstrap: %v", err)
	}

	return nil
}

func (a *App) Identity() *storage.Identity {
	return a.identity
}

func (a *App) PeerID() string {
	if a.identity == nil {
		return ""
	}
	return a.identity.PeerID
}

func (a *App) MyAddr() string {
	if a.Node == nil {
		return ""
	}
	return a.Node.MyAddr()
}

func (a *App) AllAddrs() []string {
	if a.Node == nil {
		return nil
	}
	return a.Node.AllAddrs()
}

func (a *App) SetDisplayName(name string) {
	a.identity.DisplayName = name
	a.identity.UpdatedAt = time.Now().UTC()
	_ = a.Store.SaveIdentity(a.identity)
}

func (a *App) GetPeerDisplayName(peerID string) string {
	if a.identity != nil && a.identity.PeerID == peerID {
		return a.identity.DisplayName
	}
	p, err := a.peerManager.GetPeer(peerID)
	if err != nil || p == nil {
		if len(peerID) > 12 {
			return peerID[:12]
		}
		return peerID
	}
	return p.DisplayName
}

func (a *App) SendMessage(channelID, content, contentType string) error {
	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	cipher := crypto.NewCipher(make([]byte, 32))
	env, err := protocol.NewEncryptedMessage(a.PeerID(), channelID, msgID, content, contentType, "", cipher)
	if err != nil {
		return fmt.Errorf("create encrypted message: %w", err)
	}

	msg := &storage.Message{
		MessageID:     msgID,
		ChannelID:     channelID,
		SenderPeerID:  a.PeerID(),
		Content:       content,
		ContentType:   contentType,
		Encrypted:     env.EncryptedData != nil,
		DeliveryState: "sent",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := a.Store.SaveMessage(msg); err != nil {
		return fmt.Errorf("save message: %w", err)
	}

	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:        "message",
			SenderID:    a.PeerID(),
			ChannelID:   channelID,
			MessageID:   msgID,
			Content:     content,
			ContentType: contentType,
			Timestamp:   time.Now().UnixMilli(),
		})
	}

	return nil
}

func (a *App) Connect(addrStr string) error {
	if a.Node == nil {
		return fmt.Errorf("network not initialized")
	}
	peerID, err := a.Node.Connect(context.Background(), addrStr)
	if err != nil {
		return err
	}
	_ = a.Node.SyncWithPeer(context.Background(), peerID)
	return nil
}

func (a *App) DisconnectAll() {
	if a.Node == nil {
		return
	}
	for _, p := range a.Node.Host.Network().Peers() {
		_ = a.Node.Disconnect(p)
	}
}

func (a *App) ListPeers() ([]*storage.Peer, error) {
	peers, err := a.peerManager.ListPeers()
	if err != nil {
		return nil, err
	}

	if a.Node != nil {
		connected := make(map[string]bool)
		for _, p := range a.Node.ConnectedPeers() {
			connected[p.String()] = true
		}

		for _, p := range peers {
			if connected[p.PeerID] {
				p.Status = "online"
			}
		}

		for _, pid := range a.Node.ConnectedPeers() {
			pidStr := pid.String()
			found := false
			for _, p := range peers {
				if p.PeerID == pidStr {
					found = true
					break
				}
			}
			if !found {
				peers = append(peers, &storage.Peer{
					PeerID:      pidStr,
					DisplayName: pidStr,
					Status:      "online",
				})
			}
		}
	}

	return peers, nil
}

func (a *App) ConnectedCount() int {
	if a.Node == nil {
		return 0
	}
	return a.Node.ConnectedCount()
}

func (a *App) CreateOrg(name, description string) (*storage.Organization, error) {
	return a.orgManager.CreateOrganization(name, description, a.PeerID())
}

func (a *App) ListOrgs() ([]*storage.Organization, error) {
	return a.Store.ListOrganizations()
}

func (a *App) CreateChannel(name, channelType string) (*storage.Channel, error) {
	ch, err := a.channelManager.CreateChannel("", name, channelType, "", "")
	if err != nil {
		return nil, err
	}
	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:        "sync_channel",
			SenderID:    a.PeerID(),
			ChannelID:   ch.ChannelID,
			Content:     name,
			ChannelType: channelType,
			Timestamp:   time.Now().UnixMilli(),
		})
	}
	return ch, nil
}

func (a *App) DeleteChannel(channelID string) error {
	return a.Store.ArchiveChannel(channelID)
}

func (a *App) ListChannels() ([]*storage.Channel, error) {
	return a.channelManager.ListChannels("")
}

func (a *App) ListMessages(channelID string, limit, offset int) ([]*storage.Message, error) {
	return a.Store.ListMessages(channelID, limit, offset)
}

func (a *App) OpenDM(peerID string) error {
	myID := a.PeerID()
	channelID := fmt.Sprintf("dm_%s_%s", myID, peerID)
	if myID > peerID {
		channelID = fmt.Sprintf("dm_%s_%s", peerID, myID)
	}

	ch, err := a.channelManager.GetChannel(channelID)
	if err != nil {
		return err
	}
	if ch == nil {
		now := time.Now().UTC()
		ch = &storage.Channel{
			ChannelID:   channelID,
			Name:        fmt.Sprintf("DM-%s", peerID[:8]),
			ChannelType: "dm",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := a.Store.SaveChannel(ch); err != nil {
			return fmt.Errorf("save dm channel: %w", err)
		}
	}

	a.activeChan = channelID
	return nil
}

func (a *App) Close() error {
	a.Logger.Info("shutting down...")
	if a.Node != nil {
		if err := a.Node.Close(); err != nil {
			a.Logger.Error("close network: %v", err)
		}
	}
	if a.Store != nil {
		if err := a.Store.Close(); err != nil {
			a.Logger.Error("close store: %v", err)
		}
	}
	return a.Logger.Close()
}
