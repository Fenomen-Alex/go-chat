package network

import (
	"time"

	"go-chat/internal/storage"

	"github.com/libp2p/go-libp2p/core/peer"
)

type mdnsNotifee struct {
	node *Node
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == m.node.Host.ID() {
		return
	}
	m.node.Logger.Info("mDNS discovered peer: %s", pi.ID.String())

	ctx := m.node.Ctx
	if err := m.node.Host.Connect(ctx, pi); err != nil {
		m.node.Logger.Debug("mDNS connect to %s: %v", pi.ID.String(), err)
		return
	}
	m.node.Logger.Info("mDNS connected to: %s", pi.ID.String())

	if m.node.Store != nil {
		_ = m.node.Store.SavePeer(&storage.Peer{
			PeerID:      pi.ID.String(),
			DisplayName: pi.ID.String(),
			Status:      "online",
			LastSeen:    time.Now().UTC(),
		})
	}

	_ = m.node.SyncWithPeer(ctx, pi.ID)
}
