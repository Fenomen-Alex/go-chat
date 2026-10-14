package gui

import (
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"go-chat/internal/app"
	"go-chat/internal/config"
	"go-chat/internal/crypto"
	"go-chat/internal/storage"
)

// Used to satisfy staticcheck that config is used
var _ = config.Config{}

// Service provides a GUI-facing API to the go-chat application.
// It exposes only the necessary functionality for the GUI and validates all inputs.
type Service struct {
	app    *app.App
	cfgDir string
}

// NewService creates a new GUI service from an existing app instance.
// It resolves paths in the config to ensure data is stored in the proper user directories.
func NewService(a *app.App, configPath string) *Service {
	s := &Service{app: a}

	// Resolve paths so that when launched from a shortcut, data goes to the right place
	if configPath == "" {
		s.cfgDir = getAppConfigDir()
	} else {
		s.cfgDir = getConfigDir(configPath)
	}
	resolvePaths(a.Config, configPath)

	return s
}

// Identity returns the current user's identity (without private key).
func (s *Service) Identity() (*IdentityView, error) {
	id := s.app.Identity()
	if id == nil {
		return nil, errors.New("identity not found")
	}

	fp := crypto.Fingerprint(id.PublicKey)
	isDefault := strings.HasPrefix(id.DisplayName, "user-")

	return &IdentityView{
		DisplayName: id.DisplayName,
		PeerID:      id.PeerID,
		PublicKey:   hex.EncodeToString(id.PublicKey),
		AvatarColor: id.AvatarColor,
		Fingerprint: fp,
		CreatedAt:   id.CreatedAt.UTC().Format(time.RFC3339),
		IsDefault:   isDefault,
	}, nil
}

// SetDisplayName updates the user's display name.
func (s *Service) SetDisplayName(input *SetDisplayNameInput) error {
	if err := ValidateDisplayName(input.Name); err != nil {
		return err
	}
	s.app.SetDisplayName(input.Name)
	return nil
}

// Peers returns a list of known peers.
func (s *Service) Peers() ([]*PeerView, error) {
	peers, err := s.app.ListPeers()
	if err != nil {
		return nil, err
	}

	peerViews := make([]*PeerView, 0, len(peers))
	for _, p := range peers {
		peerViews = append(peerViews, &PeerView{
			PeerID:     p.PeerID,
			DisplayName: p.DisplayName,
			Status:     p.Status,
			Trusted:    p.Trusted,
			Blocked:    p.Blocked,
			Favorite:   p.Favorite,
			AvatarColor: p.AvatarColor,
			Fingerprint: hex.EncodeToString(p.PublicKey),
			FirstSeen:  p.FirstSeen.UTC().Format(time.RFC3339),
			LastSeen:   p.LastSeen.UTC().Format(time.RFC3339),
		})
	}
	return peerViews, nil
}

// ConnectedPeers returns currently connected peers with diagnostics.
func (s *Service) ConnectedPeers() (*DiagnosticsView, error) {
	var peerViews []*PeerView
	for _, p := range s.app.Node.Host.Network().Peers() {
		peer, err := s.app.Store.GetPeer(p.String())
		if err != nil || peer == nil {
			// Create a minimal peer view for connected peers not in storage
			peerViews = append(peerViews, &PeerView{
				PeerID:   p.String(),
				Status:   "online",
			})
			continue
		}
		peerViews = append(peerViews, &PeerView{
			PeerID:     peer.PeerID,
			DisplayName: peer.DisplayName,
			Status:     peer.Status,
			Trusted:    peer.Trusted,
			Blocked:    peer.Blocked,
			Favorite:   peer.Favorite,
			AvatarColor: peer.AvatarColor,
			Fingerprint: hex.EncodeToString(peer.PublicKey),
			FirstSeen:  peer.FirstSeen.UTC().Format(time.RFC3339),
			LastSeen:   peer.LastSeen.UTC().Format(time.RFC3339),
		})
	}

	return &DiagnosticsView{
		PeerID:          s.app.PeerID(),
		ListeningAddrs:  s.app.AllAddrs(),
		ConnectedPeers:  peerViews,
		ConnectedCount:  s.app.ConnectedCount(),
		IsConnected:     s.app.Node != nil && s.app.Node.Host.Network().Peers() != nil && len(s.app.Node.Host.Network().Peers()) > 0,
	}, nil
}

// Connect attempts to connect to a peer at the given multiaddr.
func (s *Service) Connect(input *ConnectInput) error {
	if err := ValidateMultiaddr(input.Addr); err != nil {
		return err
	}
	return s.app.Connect(input.Addr)
}

// DisconnectAll disconnects from all peers.
func (s *Service) DisconnectAll() {
	s.app.DisconnectAll()
}

// SaveConnection saves a connection for quick reconnection.
func (s *Service) SaveConnection(input *SaveConnectionInput) error {
	if err := ValidateMultiaddr(input.Addr); err != nil {
		return err
	}
	s.app.SaveConnection(input.Addr)
	return nil
}

// ListConnections returns saved connections.
func (s *Service) ListConnections() ([]*storage.Connection, error) {
	return s.app.ListConnections()
}

// Organizations returns a list of organizations.
func (s *Service) Organizations() ([]*OrganizationView, error) {
	orgs, err := s.app.ListOrgs()
	if err != nil {
		return nil, err
	}

	orgViews := make([]*OrganizationView, 0, len(orgs))
	for _, o := range orgs {
		orgViews = append(orgViews, &OrganizationView{
			OrgID:        o.OrgID,
			Name:         o.Name,
			Description:  o.Description,
			Icon:         o.Icon,
			OwnerPeerID:  o.OwnerPeerID,
			Visibility:   o.Visibility,
			MaxMembers:   o.MaxMembers,
			HistoryRet:   o.HistoryRetention,
			AttachmentLimit: o.AttachmentLimit,
			Archived:     o.Archived,
			CreatedAt:    o.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    o.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return orgViews, nil
}

// CreateOrg creates a new organization.
func (s *Service) CreateOrg(input *CreateOrgInput) (*OrganizationView, error) {
	if err := ValidateOrganizationName(input.Name); err != nil {
		return nil, err
	}
	if err := ValidateOrganizationName(input.Description); err != nil {
		return nil, err
	}

	org, err := s.app.CreateOrg(input.Name, input.Description)
	if err != nil {
		return nil, err
	}

	return &OrganizationView{
		OrgID:        org.OrgID,
		Name:         org.Name,
		Description:  org.Description,
		Icon:         org.Icon,
		OwnerPeerID:  org.OwnerPeerID,
		Visibility:   org.Visibility,
		MaxMembers:   org.MaxMembers,
		HistoryRet:   org.HistoryRetention,
		AttachmentLimit: org.AttachmentLimit,
		Archived:     org.Archived,
		CreatedAt:    org.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    org.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// Channels returns a list of public and private channels the user has access to.
func (s *Service) Channels() ([]*ChannelView, error) {
	channels, err := s.app.ListChannels()
	if err != nil {
		return nil, err
	}

	// Get unread counts for each channel
	channelViews := make([]*ChannelView, 0, len(channels))
	for _, ch := range channels {
		unreadCount := 0
		// For simplicity, we'll just check if there are any messages we haven't seen
		// A real implementation would track read state per user per channel
		msgCount, err := s.app.CountChannelMessages(ch.ChannelID)
		if err == nil {
			// In a real app, we'd compare against last read timestamp
			// For now, we'll just show if there are messages
			if msgCount > 0 {
				unreadCount = msgCount // placeholder
			}
		}

		channelViews = append(channelViews, &ChannelView{
			ChannelID:   ch.ChannelID,
			Name:        ch.Name,
			ChannelType: ch.ChannelType,
			OrgID:       ch.OrgID,
			Description: ch.Description,
			UnreadCount: unreadCount,
			Archived:    ch.Archived,
		})
	}
	return channelViews, nil
}

// CreateChannel creates a new public channel.
func (s *Service) CreateChannel(input *CreateChannelInput) (*ChannelView, error) {
	if err := ValidateChannelName(input.Name); err != nil {
		return nil, err
	}
	if input.ChannelType != "text" && input.ChannelType != "private" {
		return nil, &ValidationError{Field: "channel_type", Message: "channel type must be 'text' or 'private'"}
	}
	if err := ValidateChannelName(input.Description); err != nil {
		return nil, err
	}

	var ch *storage.Channel
	var err error
	if input.ChannelType == "private" {
		ch, err = s.app.CreatePrivateChannel(input.Name, input.Description)
	} else {
		ch, err = s.app.CreateChannel(input.Name, input.ChannelType)
	}
	if err != nil {
		return nil, err
	}

	return &ChannelView{
		ChannelID:   ch.ChannelID,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		OrgID:       ch.OrgID,
		Description: ch.Description,
		Archived:    ch.Archived,
	}, nil
}

// InviteToChannel invites a peer to a private channel.
func (s *Service) InviteToChannel(input *InviteToChannelInput) (*InviteView, error) {
	if input.ChannelID == "" {
		return nil, &ValidationError{Field: "channel_id", Message: "channel ID cannot be empty"}
	}
	if input.PeerID == "" {
		return nil, &ValidationError{Field: "peer_id", Message: "peer ID cannot be empty"}
	}

	inviteID, err := s.app.InviteToChannel(input.ChannelID, input.PeerID)
	if err != nil {
		return nil, err
	}

	inv, err := s.app.Store.GetInvite(inviteID)
	if err != nil || inv == nil {
		return nil, errors.New("failed to create invite")
	}

	return &InviteView{
		InviteID:   inv.InviteID,
		SenderPeerID: inv.SenderPeerID,
		TargetPeerID: inv.TargetPeerID,
		OrgID:      inv.OrgID,
		ChannelID:  inv.ChannelID,
		InviteType: inv.InviteType,
		Message:    inv.Message,
		OneTime:    inv.OneTime,
		MaxUses:    inv.MaxUses,
		UseCount:   inv.UseCount,
		ExpiresAt:  func() string {
			if inv.ExpiresAt == nil {
				return ""
			}
			return inv.ExpiresAt.UTC().Format(time.RFC3339)
		}(),
		CreatedAt:  inv.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ListPendingInvites returns pending invites for the current user.
func (s *Service) ListPendingInvites() ([]*InviteView, error) {
	invites, err := s.app.ListPendingInvites()
	if err != nil {
		return nil, err
	}

	invViews := make([]*InviteView, 0, len(invites))
	for _, inv := range invites {
		invViews = append(invViews, &InviteView{
			InviteID:   inv.InviteID,
			SenderPeerID: inv.SenderPeerID,
			TargetPeerID: inv.TargetPeerID,
			OrgID:      inv.OrgID,
			ChannelID:  inv.ChannelID,
			InviteType: inv.InviteType,
			Message:    inv.Message,
			OneTime:    inv.OneTime,
			MaxUses:    inv.MaxUses,
			UseCount:   inv.UseCount,
			ExpiresAt:  func() string {
				if inv.ExpiresAt == nil {
					return ""
				}
				return inv.ExpiresAt.UTC().Format(time.RFC3339)
			}(),
			CreatedAt:  inv.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return invViews, nil
}

// AcceptInvite accepts a channel invite.
func (s *Service) AcceptInvite(input *AcceptInviteInput) (*ChannelView, error) {
	if input.InviteID == "" {
		return nil, &ValidationError{Field: "invite_id", Message: "invite ID cannot be empty"}
	}

	channelID, err := s.app.AcceptInvite(input.InviteID)
	if err != nil {
		return nil, err
	}

	ch, err := s.app.Store.GetChannel(channelID)
	if err != nil || ch == nil {
		return nil, errors.New("failed to get channel after accepting invite")
	}

	return &ChannelView{
		ChannelID:   ch.ChannelID,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		OrgID:       ch.OrgID,
		Description: ch.Description,
		Archived:    ch.Archived,
	}, nil
}

// DMChannels returns direct message channels.
func (s *Service) DMChannels() ([]*ChannelView, error) {
	dms, err := s.app.ListDMChannels()
	if err != nil {
		return nil, err
	}

	dmViews := make([]*ChannelView, 0, len(dms))
	for _, dm := range dms {
		dmViews = append(dmViews, &ChannelView{
			ChannelID:   dm.ChannelID,
			Name:        dm.Name,
			ChannelType: dm.ChannelType,
			OrgID:       dm.OrgID,
			Description: dm.Description,
			Archived:    dm.Archived,
		})
	}
	return dmViews, nil
}

// OpenDM opens or creates a direct message channel with a peer.
func (s *Service) OpenDM(peerID string) (*ChannelView, error) {
	if peerID == "" {
		return nil, &ValidationError{Field: "peer_id", Message: "peer ID cannot be empty"}
	}

	channelID, err := s.app.OpenDM(peerID)
	if err != nil {
		return nil, err
	}

	ch, err := s.app.Store.GetChannel(channelID)
	if err != nil || ch == nil {
		return nil, errors.New("failed to get or create DM channel")
	}

	return &ChannelView{
		ChannelID:   ch.ChannelID,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		OrgID:       ch.OrgID,
		Description: ch.Description,
		Archived:    ch.Archived,
	}, nil
}

// Messages returns messages for a channel.
func (s *Service) Messages(input *ListMessagesInput) ([]*MessageView, error) {
	if input.ChannelID == "" {
		return nil, &ValidationError{Field: "channel_id", Message: "channel ID cannot be empty"}
	}
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	messages, err := s.app.ListMessages(input.ChannelID, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	msgViews := make([]*MessageView, 0, len(messages))
	for _, msg := range messages {
		senderName := s.app.GetPeerDisplayName(msg.SenderPeerID)
		isSelf := msg.SenderPeerID == s.app.PeerID()

		msgViews = append(msgViews, &MessageView{
			MessageID:   msg.MessageID,
			ChannelID:   msg.ChannelID,
			SenderPeerID: msg.SenderPeerID,
			SenderName:  senderName,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Timestamp:   msg.CreatedAt.UTC().Format(time.RFC3339),
			IsSelf:      isSelf,
			DeliveryState: msg.DeliveryState,
		})
	}
	return msgViews, nil
}

// SendMessage sends a message to a channel.
func (s *Service) SendMessage(channelID, content, contentType string) error {
	if channelID == "" {
		return &ValidationError{Field: "channel_id", Message: "channel ID cannot be empty"}
	}
	if content == "" {
		return &ValidationError{Field: "content", Message: "message content cannot be empty"}
	}
	if len(content) > 4000 {
		return &ValidationError{Field: "content", Message: "message too long (max 4000 characters)"}
	}
	if contentType == "" {
		contentType = "text"
	}
	return s.app.SendMessage(channelID, content, contentType)
}

// Close shuts down the application cleanly.
func (s *Service) Close() error {
	return s.app.Close()
}

// Helper functions

func getConfigDir(configPath string) string {
	if configPath == "" {
		return getAppConfigDir()
	}
	return filepath.Dir(configPath)
}