package network

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"go-chat/internal/crypto"
	"go-chat/internal/storage"

	"github.com/libp2p/go-libp2p/core/network"
)

var errLineTooLong = errors.New("line exceeds max size")

func readLine(r *bufio.Reader, max int) ([]byte, error) {
	var total int
	for {
		b, err := r.ReadSlice('\n')
		total += len(b)
		if total > max {
			return nil, errLineTooLong
		}
		if err == nil || err == bufio.ErrBufferFull {
			if err == nil {
				return b, nil
			}
			continue
		}
		return b, err
	}
}

type StreamHandler struct {
	node    *Node
	mu      sync.Mutex
	dedupMu sync.Mutex
}

func NewStreamHandler(node *Node) *StreamHandler {
	return &StreamHandler{node: node}
}

func (h *StreamHandler) Handle(s network.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer().String()
	r := bufio.NewReader(s)

	for {
		s.SetReadDeadline(time.Now().Add(30 * time.Second))
		data, err := readLine(r, 1<<20)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if err != io.EOF && err != errLineTooLong {
				h.node.Logger.Debug("stream read error from %s: %v", peerID, err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			h.node.Logger.Debug("invalid message from %s: %v", peerID, err)
			continue
		}

		if msg.Type == "key_exchange" {
			h.handleKeyExchange(s, peerID, &msg)
			continue
		}

		h.DecryptMessage(&msg, peerID)

		if msg.Type == "sync_request" {
			h.handleSyncRequest(s, peerID, r)
			continue
		}

		h.handleMessage(&msg, s)
	}
}

func (h *StreamHandler) handleSyncRequest(s network.Stream, peerID string, r *bufio.Reader) {
	h.node.Logger.Info("handling sync request from %s", peerID)

	for {
		s.SetReadDeadline(time.Now().Add(60 * time.Second))
		data, err := readLine(r, 1<<20)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				h.node.Logger.Debug("sync read timeout from %s, assuming done", peerID)
				break
			}
			h.node.Logger.Debug("sync read error from %s: %v", peerID, err)
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			h.node.Logger.Debug("invalid sync message from %s: %v", peerID, err)
			continue
		}

		h.DecryptMessage(&msg, peerID)

		if msg.Type == "sync_complete" {
			break
		}

		h.handleMessage(&msg, s)
	}

	h.sendFullState(s)
}

func (h *StreamHandler) handleMessage(msg *Message, s network.Stream) {
	remotePeerID := h.remotePeerID(s)
	h.node.Logger.Info("received %s from %s", msg.Type, remotePeerID)

	if msg.SenderID != "" && msg.SenderID != remotePeerID {
		known, err := h.node.Store.GetPeer(msg.SenderID)
		if err == nil && known == nil {
			h.node.Logger.Debug("rejected %s from %s claiming unknown peer %s", msg.Type, remotePeerID, msg.SenderID)
			return
		}
	}

	switch msg.Type {
	case "sync_org":
		if h.node.Store != nil {
			h.handleSyncOrg(msg)
		}
	case "sync_channel":
		if h.node.Store != nil {
			h.handleSyncChannel(msg, s)
			h.notifyRefresh()
		}
	case "message":
		if h.node.Store != nil {
			h.handleSyncMessage(msg)
			h.notifyRefresh()
		}
	case "sync_peer":
		if h.node.Store != nil {
			h.handleSyncPeer(msg, remotePeerID)
			h.notifyRefresh()
		}
	default:
		h.node.Logger.Debug("unknown message type: %s from %s", msg.Type, remotePeerID)
	}
}

func (h *StreamHandler) notifyRefresh() {
	if h.node.RefreshCh != nil {
		select {
		case h.node.RefreshCh <- struct{}{}:
		default:
		}
	}
}

func (h *StreamHandler) sendFullState(s network.Stream) {
	h.sendSyncState(s)

	allMsgs, err := h.node.Store.ListAllMessages(10000)
	if err == nil {
		remotePeerID := h.remotePeerID(s)
		for _, m := range allMsgs {
			if strings.HasPrefix(m.ChannelID, "dm_") && !strings.Contains(m.ChannelID, remotePeerID) {
				continue
			}
			if err := h.SendMessage(s, &Message{
				Type:        "message",
				SenderID:    m.SenderPeerID,
				ChannelID:   m.ChannelID,
				MessageID:   m.MessageID,
				Content:     m.Content,
				ContentType: m.ContentType,
				Timestamp:   m.CreatedAt.UnixMilli(),
			}); err != nil {
				h.node.Logger.Warn("send message during full state sync: %v", err)
			}
		}
	}
}

func (h *StreamHandler) sendSyncState(s network.Stream) {
	remotePeerID := h.remotePeerID(s)

	orgs, err := h.node.Store.ListOrganizations()
	if err == nil {
		for _, org := range orgs {
			if err := h.SendMessage(s, &Message{
				Type:      "sync_org",
				SenderID:  h.node.Host.ID().String(),
				OrgID:     org.OrgID,
				Content:   org.Name,
				Timestamp: org.CreatedAt.UnixMilli(),
			}); err != nil {
				h.node.Logger.Debug("send sync_org: %v", err)
			}
		}
	} else {
		h.node.Logger.Warn("list orgs for sync: %v", err)
	}

	channels, err := h.node.Store.ListAllChannels()
	if err == nil {
		for _, ch := range channels {
			if strings.HasPrefix(ch.ChannelID, "dm_") && !strings.Contains(ch.ChannelID, remotePeerID) {
				continue
			}
			if err := h.SendMessage(s, &Message{
				Type:        "sync_channel",
				SenderID:    h.node.Host.ID().String(),
				OrgID:       ch.OrgID,
				ChannelID:   ch.ChannelID,
				Content:     ch.Name,
				ChannelType: ch.ChannelType,
				Timestamp:   ch.CreatedAt.UnixMilli(),
			}); err != nil {
				h.node.Logger.Debug("send sync_channel: %v", err)
			}
		}
	} else {
		h.node.Logger.Warn("list channels for sync: %v", err)
	}

	allPeers, err := h.node.Store.ListPeers()
	if err == nil {
		for _, p := range allPeers {
			if p.DisplayName == "" || p.DisplayName == p.PeerID || p.DisplayName == "me" || strings.HasPrefix(p.DisplayName, "me_") {
				continue
			}
			if err := h.SendMessage(s, &Message{
				Type:      "sync_peer",
				SenderID:  p.PeerID,
				Content:   p.DisplayName,
				Timestamp: time.Now().UnixMilli(),
			}); err != nil {
				h.node.Logger.Debug("send sync_peer: %v", err)
			}
		}
	} else {
		h.node.Logger.Warn("list peers for sync: %v", err)
	}

	myName := h.node.Host.ID().String()
	identity, err := h.node.Store.GetIdentity()
	if err != nil {
		h.node.Logger.Warn("get identity for sync state: %v", err)
	}
	if identity != nil {
		myName = identity.DisplayName
	}
	if err := h.SendMessage(s, &Message{
		Type:      "sync_peer",
		SenderID:  h.node.Host.ID().String(),
		Content:   myName,
		Timestamp: time.Now().UnixMilli(),
	}); err != nil {
		h.node.Logger.Debug("send self sync_peer: %v", err)
	}
}

func (h *StreamHandler) peerCanAccess(peerID, channelID string) bool {
	if !strings.HasPrefix(channelID, "dm_") {
		return true
	}
	return strings.Contains(channelID, peerID)
}

func (h *StreamHandler) ensurePeerExists(peerID string) {
	if peerID == "" {
		return
	}
	existing, err := h.node.Store.GetPeer(peerID)
	if err != nil {
		h.node.Logger.Warn("get peer %s: %v", peerID, err)
	}
	if existing != nil {
		return
	}
	p := &storage.Peer{
		PeerID:      peerID,
		DisplayName: peerID,
		Status:      "unknown",
		FirstSeen:   time.Now().UTC(),
		LastSeen:    time.Now().UTC(),
	}
	if err := h.node.Store.SavePeer(p); err != nil {
		h.node.Logger.Warn("create stub peer: %v", err)
	}
}

func (h *StreamHandler) handleSyncMessage(msg *Message) {
	if msg.MessageID == "" || msg.ChannelID == "" {
		return
	}
	h.ensurePeerExists(msg.SenderID)

	existing, err := h.node.Store.GetChannel(msg.ChannelID)
	if err != nil {
		h.node.Logger.Warn("get channel %s: %v", msg.ChannelID, err)
	}
	if existing == nil {
		name := msg.ChannelID
		if strings.HasPrefix(msg.ChannelID, "dm_") {
			name = "DM"
		}
		ch := &storage.Channel{
			ChannelID:   msg.ChannelID,
			Name:        name,
			ChannelType: "text",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		if msg.ChannelType != "" {
			ch.ChannelType = msg.ChannelType
		}
		if strings.HasPrefix(msg.ChannelID, "dm_") {
			ch.ChannelType = "dm"
		}
		if err := h.node.Store.SaveChannel(ch); err != nil {
			h.node.Logger.Warn("create channel from message: %v", err)
		}
	}
	storeMsg := &storage.Message{
		MessageID:     msg.MessageID,
		ChannelID:     msg.ChannelID,
		SenderPeerID:  msg.SenderID,
		Content:       msg.Content,
		ContentType:   msg.ContentType,
		DeliveryState: "received",
		CreatedAt:     time.UnixMilli(msg.Timestamp).UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := h.node.Store.SaveMessage(storeMsg); err != nil {
		h.node.Logger.Warn("save synced message: %v", err)
	}
}

func (h *StreamHandler) handleSyncOrg(msg *Message) {
	if msg.OrgID == "" {
		return
	}
	existing, err := h.node.Store.GetOrganization(msg.OrgID)
	if err != nil {
		h.node.Logger.Warn("get organization %s: %v", msg.OrgID, err)
	}
	if existing != nil {
		return
	}
	org := &storage.Organization{
		OrgID:       msg.OrgID,
		Name:        msg.Content,
		OwnerPeerID: msg.SenderID,
		CreatedAt:   time.UnixMilli(msg.Timestamp).UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := h.node.Store.SaveOrganization(org); err != nil {
		h.node.Logger.Warn("save synced org: %v", err)
	}
}

func (h *StreamHandler) handleSyncChannel(msg *Message, s network.Stream) {
	if msg.ChannelID == "" {
		return
	}
	if strings.HasPrefix(msg.ChannelID, "dm_") {
		localPeerID := h.node.Host.ID().String()
		if !strings.Contains(msg.ChannelID, localPeerID) {
			return
		}
	}
	existing, err := h.node.Store.GetChannel(msg.ChannelID)
	if err != nil {
		h.node.Logger.Warn("get channel %s: %v", msg.ChannelID, err)
	}
	if existing != nil {
		return
	}
	ch := &storage.Channel{
		ChannelID:   msg.ChannelID,
		OrgID:       msg.OrgID,
		Name:        msg.Content,
		ChannelType: msg.ChannelType,
		CreatedAt:   time.UnixMilli(msg.Timestamp).UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if ch.ChannelType == "" {
		ch.ChannelType = "text"
	}
	if err := h.node.Store.SaveChannel(ch); err != nil {
		h.node.Logger.Warn("save synced channel: %v", err)
	}
}

func (h *StreamHandler) handleSyncPeer(msg *Message, remotePeerID string) {
	if msg.SenderID == "" || msg.Content == "" {
		return
	}
	if msg.Content == "me" || strings.HasPrefix(msg.Content, "me_") {
		return
	}
	// Sender authentication: only allow a peer to claim their own identity
	// or update an already-known peer
	if msg.SenderID != remotePeerID {
		known, err := h.node.Store.GetPeer(msg.SenderID)
		if err != nil {
			h.node.Logger.Warn("get peer %s: %v", msg.SenderID, err)
		}
		if known == nil {
			h.node.Logger.Debug("rejected sync_peer from %s claiming unknown peer %s", remotePeerID, msg.SenderID)
			return
		}
	}
	name := msg.Content
	h.dedupMu.Lock()
	existing, err := h.node.Store.GetPeerByDisplayName(name)
	if err != nil {
		h.node.Logger.Warn("get peer by display name %s: %v", name, err)
	}
	if existing != nil && existing.PeerID != msg.SenderID {
		suffix := 1
		for {
			candidate := fmt.Sprintf("%s_%d", name, suffix)
			dup, err := h.node.Store.GetPeerByDisplayName(candidate)
			if err != nil {
				h.node.Logger.Warn("get peer by display name %s: %v", candidate, err)
			}
			if dup == nil {
				name = candidate
				break
			}
			suffix++
		}
	}
	h.dedupMu.Unlock()
	if err := h.node.Store.SavePeer(&storage.Peer{
		PeerID:      msg.SenderID,
		DisplayName: name,
		Status:      "online",
		LastSeen:    time.Now().UTC(),
	}); err != nil {
		h.node.Logger.Warn("save synced peer: %v", err)
	}
}

func (h *StreamHandler) remotePeerID(s network.Stream) string {
	return s.Conn().RemotePeer().String()
}

func (h *StreamHandler) SendMessage(s network.Stream, msg *Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	sendMsg := *msg
	peerID := s.Conn().RemotePeer().String()

	if key, ok := h.node.GetSessionKey(peerID); ok && len(key) > 0 {
		c := crypto.NewCipher(key)
		encrypted, err := c.Encrypt([]byte(sendMsg.Content))
		if err == nil {
			sendMsg.Encrypted = true
			sendMsg.EncryptedData = encrypted
			sendMsg.Content = ""
		}
	}

	data, err := json.Marshal(sendMsg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')

	s.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if _, err := s.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (h *StreamHandler) DecryptMessage(msg *Message, peerID string) {
	if !msg.Encrypted || msg.EncryptedData == nil {
		return
	}
	key, ok := h.node.GetSessionKey(peerID)
	if !ok || len(key) == 0 {
		return
	}
	c := crypto.NewCipher(key)
	decrypted, err := c.Decrypt(msg.EncryptedData)
	if err != nil {
		h.node.Logger.Debug("decrypt failed from %s: %v", peerID, err)
		return
	}
	msg.Content = string(decrypted)
	msg.Encrypted = false
	msg.EncryptedData = nil
}

func (h *StreamHandler) handleKeyExchange(s network.Stream, peerID string, msg *Message) {
	peerPub, err := base64.StdEncoding.DecodeString(msg.Content)
	if err != nil || len(peerPub) != 32 {
		h.node.Logger.Debug("invalid key_exchange from %s", peerID)
		return
	}

	ephPriv, ephPub, err := crypto.GenerateEphemeralKeypair()
	if err != nil {
		h.node.Logger.Warn("generate ephemeral keypair: %v", err)
		return
	}

	pubB64 := base64.StdEncoding.EncodeToString(ephPub)
	h.SendMessage(s, &Message{
		Type:     "key_exchange_ack",
		SenderID: h.node.Host.ID().String(),
		Content:  pubB64,
	})

	shared, err := crypto.ComputeSharedSecret(ephPriv, peerPub)
	if err != nil {
		h.node.Logger.Warn("x25519 shared secret: %v", err)
		return
	}

	salt := append(peerPub, ephPub...)
	key, err := crypto.DeriveSessionKeys(shared, salt)
	if err != nil {
		h.node.Logger.Warn("derive session keys: %v", err)
		return
	}

	h.node.SetSessionKey(peerID, key)
	h.node.Logger.Debug("key exchange complete with %s", peerID)
}


