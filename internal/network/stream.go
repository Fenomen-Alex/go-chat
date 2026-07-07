package network

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go-chat/internal/storage"

	"github.com/libp2p/go-libp2p/core/network"
)

type StreamHandler struct {
	node *Node
}

func NewStreamHandler(node *Node) *StreamHandler {
	return &StreamHandler{node: node}
}

func (h *StreamHandler) Handle(s network.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer().String()
	r := bufio.NewReader(s)

	if h.node.Store != nil {
		_ = h.node.Store.SavePeer(&storage.Peer{
			PeerID:      peerID,
			DisplayName: peerID,
			Status:      "online",
			LastSeen:    time.Now().UTC(),
		})
	}

	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				h.node.Logger.Debug("stream read error from %s: %v", peerID, err)
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			h.node.Logger.Debug("invalid message from %s: %v", peerID, err)
			continue
		}

		h.handleMessage(&msg, s)
	}
}

func (h *StreamHandler) handleMessage(msg *Message, s network.Stream) {
	h.node.Logger.Info("received %s from %s", msg.Type, msg.SenderID)

	switch msg.Type {
	case "sync_request":
		h.sendFullState(s)
	case "sync_org":
		if h.node.Store != nil {
			h.handleSyncOrg(msg)
		}
	case "sync_channel":
		if h.node.Store != nil {
			h.handleSyncChannel(msg, s)
		}
	case "message":
		if h.node.Store != nil {
			h.handleSyncMessage(msg)
		}
	default:
		h.node.Logger.Debug("unknown message type: %s", msg.Type)
	}
}

func (h *StreamHandler) sendFullState(s network.Stream) {
	remotePeerID := h.remotePeerID(s)

	orgs, err := h.node.Store.ListOrganizations()
	if err == nil {
		for _, org := range orgs {
			_ = h.SendMessage(s, &Message{
				Type:      "sync_org",
				SenderID:  h.node.Host.ID().String(),
				OrgID:     org.OrgID,
				Content:   org.Name,
				Timestamp: org.CreatedAt.UnixMilli(),
			})
		}
	}

	channels, err := h.node.Store.ListAllChannels()
	if err == nil {
		for _, ch := range channels {
			if strings.HasPrefix(ch.ChannelID, "dm_") && !strings.Contains(ch.ChannelID, remotePeerID) {
				continue
			}
			_ = h.SendMessage(s, &Message{
				Type:        "sync_channel",
				SenderID:    h.node.Host.ID().String(),
				OrgID:       ch.OrgID,
				ChannelID:   ch.ChannelID,
				Content:     ch.Name,
				ChannelType: ch.ChannelType,
				Timestamp:   ch.CreatedAt.UnixMilli(),
			})
		}
	}

	allMsgs, err := h.node.Store.ListAllMessages(50)
	if err == nil {
		for _, m := range allMsgs {
			_ = h.SendMessage(s, &Message{
				Type:       "message",
				SenderID:   m.SenderPeerID,
				ChannelID:  m.ChannelID,
				MessageID:  m.MessageID,
				Content:    m.Content,
				Timestamp:  m.CreatedAt.UnixMilli(),
			})
		}
	}
}

func (h *StreamHandler) handleSyncMessage(msg *Message) {
	if msg.MessageID == "" || msg.ChannelID == "" {
		return
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
	existing, _ := h.node.Store.GetOrganization(msg.OrgID)
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
	existing, _ := h.node.Store.GetChannel(msg.ChannelID)
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

func (h *StreamHandler) remotePeerID(s network.Stream) string {
	return s.Conn().RemotePeer().String()
}

func (h *StreamHandler) SendMessage(s network.Stream, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')

	if _, err := s.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

func (h *StreamHandler) ReceiveMessages(s network.Stream) {
	r := bufio.NewReader(s)
	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			return
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		h.handleMessage(&msg, s)
	}
}
