package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go-chat/internal/channel"
	"go-chat/internal/config"
	"go-chat/internal/crypto"
	"go-chat/internal/discovery"
	"go-chat/internal/logging"
	"go-chat/internal/network"
	"go-chat/internal/organization"
	"go-chat/internal/peermgr"
	"go-chat/internal/storage"

	lp2ppeer "github.com/libp2p/go-libp2p/core/peer"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
)

type App struct {
	Config    *config.Config
	Logger    *logging.Logger
	Store     *storage.Store
	Node      *network.Node
	RefreshCh chan struct{}

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

	store, err := storage.New(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}

	app := &App{
		Config:    cfg,
		Logger:    log,
		Store:     store,
		RefreshCh: make(chan struct{}, 64),
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
		now := time.Now().UTC()
		return a.Store.SavePeer(&storage.Peer{
			PeerID:      id.PeerID,
			DisplayName: id.DisplayName,
			Status:      "online",
			FirstSeen:   now,
			LastSeen:    now,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
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
		hostname, err := os.Hostname()
		if err != nil {
			a.Logger.Warn("get hostname: %v", err)
			hostname = "unknown"
		}
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
	return a.Store.SavePeer(&storage.Peer{
		PeerID:      peerID,
		DisplayName: displayName,
		Status:      "online",
		FirstSeen:   now,
		LastSeen:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (a *App) startNetworking() error {
	node, err := network.NewNode(a.libp2pKey, &a.Config.Network, a.Logger, a.Store, a.RefreshCh)
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
	if err := a.Store.SaveIdentity(a.identity); err != nil {
		a.Logger.Warn("save identity: %v", err)
	}

	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:      "sync_peer",
			SenderID:  a.PeerID(),
			Content:   name,
			Timestamp: time.Now().UnixMilli(),
		})
	}
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
	if p.DisplayName == "" || p.DisplayName == "me" || strings.HasPrefix(p.DisplayName, "me_") {
		if len(peerID) > 12 {
			return peerID[:12]
		}
		return peerID
	}
	return p.DisplayName
}

func (a *App) IsReservedDisplayName(name string) bool {
	return name == "" || name == "me" || strings.HasPrefix(name, "me_")
}

func (a *App) SendMessage(channelID, content, contentType string) error {
	msgID := fmt.Sprintf("msg_%s_%d", a.PeerID(), time.Now().UnixNano())

	msg := &storage.Message{
		MessageID:     msgID,
		ChannelID:     channelID,
		SenderPeerID:  a.PeerID(),
		Content:       content,
		ContentType:   contentType,
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	peerID, err := a.Node.Connect(ctx, addrStr)
	if err != nil {
		return err
	}
	syncCtx, syncCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer syncCancel()
	if err := a.Node.SyncWithPeer(syncCtx, peerID); err != nil {
		a.Logger.Warn("sync with peer: %v", err)
	}
	return nil
}

func (a *App) DisconnectAll() {
	if a.Node == nil {
		return
	}
	for _, p := range a.Node.Host.Network().Peers() {
		if err := a.Node.Disconnect(p); err != nil {
			a.Logger.Warn("disconnect: %v", err)
		}
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
				if p.Status != "online" {
					p.Status = "online"
					if err := a.Store.UpdatePeerStatus(p.PeerID, "online"); err != nil {
						a.Logger.Warn("update peer status online: %v", err)
					}
				}
			} else if p.Status == "online" {
				p.Status = "offline"
				if err := a.Store.UpdatePeerStatus(p.PeerID, "offline"); err != nil {
					a.Logger.Warn("update peer status offline: %v", err)
				}
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

func (a *App) CreatePrivateChannel(name, description string) (*storage.Channel, error) {
	ch, err := a.channelManager.CreatePrivateChannel(name, description, a.PeerID())
	if err != nil {
		return nil, err
	}
	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:        "sync_channel",
			SenderID:    a.PeerID(),
			ChannelID:   ch.ChannelID,
			Content:     name,
			ChannelType: "private",
			Timestamp:   time.Now().UnixMilli(),
		})
		a.Node.Broadcast(&network.Message{
			Type:         "sync_channel_member",
			SenderID:     a.PeerID(),
			ChannelID:    ch.ChannelID,
			MemberPeerID: a.PeerID(),
			MemberRole:   "owner",
			Timestamp:    time.Now().UnixMilli(),
		})
	}
	return ch, nil
}

func (a *App) InviteToChannel(channelID, peerID string) (string, error) {
	ch, err := a.channelManager.GetChannel(channelID)
	if err != nil {
		return "", err
	}
	if ch == nil {
		return "", fmt.Errorf("channel not found")
	}
	if ch.ChannelType != "private" {
		return "", fmt.Errorf("invites are only for private channels")
	}
	member, err := a.channelManager.IsMember(channelID, a.PeerID())
	if err != nil {
		return "", err
	}
	if !member {
		return "", fmt.Errorf("you are not a member of this channel")
	}

	inviteID := fmt.Sprintf("inv_%d", time.Now().UnixNano())
	inv := &storage.Invite{
		InviteID:     inviteID,
		SenderPeerID: a.PeerID(),
		TargetPeerID: peerID,
		ChannelID:    channelID,
		InviteType:   "channel",
		OneTime:      true,
		CreatedAt:    time.Now().UTC(),
	}
	if err := a.Store.SaveInvite(inv); err != nil {
		return "", fmt.Errorf("save invite: %w", err)
	}

	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:         "sync_invite",
			SenderID:     a.PeerID(),
			InviteID:     inviteID,
			ChannelID:    channelID,
			TargetPeerID: peerID,
			Content:      fmt.Sprintf("Invite to %s", ch.Name),
			Timestamp:    time.Now().UnixMilli(),
		})
	}

	return inviteID, nil
}

func (a *App) AcceptInvite(inviteID string) (string, error) {
	inv, err := a.Store.GetInvite(inviteID)
	if err != nil {
		return "", err
	}
	if inv == nil {
		return "", fmt.Errorf("invite not found")
	}
	if inv.TargetPeerID != a.PeerID() {
		return "", fmt.Errorf("this invite is not for you")
	}

	if err := a.channelManager.AddMember(inv.ChannelID, a.PeerID(), "member"); err != nil {
		return "", fmt.Errorf("add channel member: %w", err)
	}

	if a.Node != nil {
		a.Node.Broadcast(&network.Message{
			Type:         "sync_channel_member",
			SenderID:     a.PeerID(),
			ChannelID:    inv.ChannelID,
			MemberPeerID: a.PeerID(),
			MemberRole:   "member",
			Timestamp:    time.Now().UnixMilli(),
		})
	}

	if err := a.Store.DeleteInvite(inviteID); err != nil {
		a.Logger.Warn("delete accepted invite: %v", err)
	}

	return inv.ChannelID, nil
}

func (a *App) ListPendingInvites() ([]*storage.Invite, error) {
	return a.Store.ListPendingInvites(a.PeerID())
}

func (a *App) DeleteChannel(channelID string) error {
	return a.Store.ArchiveChannel(channelID)
}

func (a *App) ListChannels() ([]*storage.Channel, error) {
	allChannels, err := a.channelManager.ListChannels("")
	if err != nil {
		return nil, err
	}
	var filtered []*storage.Channel
	for _, ch := range allChannels {
		if ch.ChannelType == "dm" {
			continue
		}
		if ch.ChannelType == "private" {
			member, err := a.channelManager.IsMember(ch.ChannelID, a.PeerID())
			if err != nil || !member {
				continue
			}
		}
		filtered = append(filtered, ch)
	}
	return filtered, nil
}

func (a *App) ListDMChannels() ([]*storage.Channel, error) {
	return a.channelManager.ListDMChannels()
}

func (a *App) ListMessages(channelID string, limit, offset int) ([]*storage.Message, error) {
	return a.Store.ListMessages(channelID, limit, offset)
}

func (a *App) CountChannelMessages(channelID string) (int, error) {
	return a.Store.CountChannelMessages(channelID)
}

func (a *App) OpenDM(peerID string) (string, error) {
	myID := a.PeerID()
	channelID := fmt.Sprintf("dm_%s_%s", myID, peerID)
	if myID > peerID {
		channelID = fmt.Sprintf("dm_%s_%s", peerID, myID)
	}

	ch, err := a.channelManager.GetChannel(channelID)
	if err != nil {
		return "", err
	}
	if ch == nil {
		dn := a.GetPeerDisplayName(peerID)
		now := time.Now().UTC()
		ch = &storage.Channel{
			ChannelID:   channelID,
			Name:        fmt.Sprintf("DM-%s", dn),
			ChannelType: "dm",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := a.Store.SaveChannel(ch); err != nil {
			return "", fmt.Errorf("save dm channel: %w", err)
		}
		if a.Node != nil {
			a.Node.Broadcast(&network.Message{
				Type:        "sync_channel",
				SenderID:    a.PeerID(),
				ChannelID:   channelID,
				Content:     ch.Name,
				ChannelType: "dm",
				Timestamp:   now.UnixMilli(),
			})
		}
	}

	a.activeChan = channelID
	return channelID, nil
}

func (a *App) IsDefaultName() bool {
	if a.identity == nil {
		return true
	}
	return strings.HasPrefix(a.identity.DisplayName, "user-")
}

func (a *App) SaveConnection(addr string) {
	conns, err := a.Store.ListConnections()
	if err != nil {
		a.Logger.Warn("list connections: %v", err)
	}
	for _, c := range conns {
		if c.Address == addr {
			c.LastConnectedAt = time.Now().UTC()
			if err := a.Store.SaveConnection(c); err != nil {
				a.Logger.Warn("save connection: %v", err)
			}
			return
		}
	}
	now := time.Now().UTC()
	if err := a.Store.SaveConnection(&storage.Connection{
		Address:         addr,
		LastConnectedAt: now,
		CreatedAt:       now,
	}); err != nil {
		a.Logger.Warn("save connection: %v", err)
	}
}

func (a *App) ListConnections() ([]*storage.Connection, error) {
	return a.Store.ListConnections()
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
