package gui

import (
	"time"
)

// Used to satisfy staticcheck that time is used
var _ = time.Second

// IdentityView represents a user's identity for the GUI (no private key).
type IdentityView struct {
	DisplayName string `json:"display_name"`
	PeerID      string `json:"peer_id"`
	PublicKey   string `json:"public_key"` // hex encoded
	AvatarColor string `json:"avatar_color"`
	Fingerprint string `json:"fingerprint"`
	CreatedAt   string `json:"created_at"`
	IsDefault   bool   `json:"is_default"`
}

// PeerView represents a peer for the GUI.
type PeerView struct {
	PeerID       string `json:"peer_id"`
	DisplayName  string `json:"display_name"`
	Status       string `json:"status"` // online, offline, etc.
	Trusted      bool   `json:"trusted"`
	Blocked      bool   `json:"blocked"`
	Favorite     bool   `json:"favorite"`
	AvatarColor  string `json:"avatar_color"`
	Fingerprint  string `json:"fingerprint"`
	FirstSeen    string `json:"first_seen"`
	LastSeen     string `json:"last_seen"`
}

// ChannelView represents a channel for the GUI.
type ChannelView struct {
	ChannelID    string `json:"channel_id"`
	Name         string `json:"name"`
	ChannelType  string `json:"channel_type"` // text, dm, private
	OrgID        string `json:"org_id"`
	Description  string `json:"description"`
	UnreadCount  int    `json:"unread_count"`
	Archived     bool   `json:"archived"`
}

// MessageView represents a message for the GUI.
type MessageView struct {
	MessageID    string `json:"message_id"`
	ChannelID    string `json:"channel_id"`
	SenderPeerID string `json:"sender_peer_id"`
	SenderName   string `json:"sender_name"`
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	Timestamp    string `json:"timestamp"`
	IsSelf       bool   `json:"is_self"`
	DeliveryState string `json:"delivery_state"`
}

// DiagnosticsView represents network diagnostics for the GUI.
type DiagnosticsView struct {
	PeerID          string      `json:"peer_id"`
	ListeningAddrs  []string    `json:"listening_addrs"`
	ConnectedPeers  []*PeerView `json:"connected_peers"`
	ConnectedCount  int         `json:"connected_count"`
	IsConnected     bool        `json:"is_connected"`
}

// OrganizationView represents an organization for the GUI.
type OrganizationView struct {
	OrgID        string `json:"org_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	OwnerPeerID  string `json:"owner_peer_id"`
	Visibility   string `json:"visibility"`
	MaxMembers   int    `json:"max_members"`
	HistoryRet int    `json:"history_retention"`
	AttachmentLimit int64 `json:"attachment_limit"`
	Archived     bool   `json:"archived"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ChannelMemberView represents a channel member for the GUI.
type ChannelMemberView struct {
	PeerID     string `json:"peer_id"`
	Role       string `json:"role"`
	JoinedAt   string `json:"joined_at"`
}

// InviteView represents an invite for the GUI.
type InviteView struct {
	InviteID   string `json:"invite_id"`
	SenderPeerID string `json:"sender_peer_id"`
	TargetPeerID string `json:"target_peer_id"`
	OrgID      string `json:"org_id"`
	ChannelID  string `json:"channel_id"`
	InviteType string `json:"invite_type"`
	Message    string `json:"message"`
	OneTime    bool   `json:"one_time"`
	MaxUses    int    `json:"max_uses"`
	UseCount   int    `json:"use_count"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// CreateOrgInput represents input for creating an organization.
type CreateOrgInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateChannelInput represents input for creating a channel.
type CreateChannelInput struct {
	Name        string `json:"name"`
	ChannelType string `json:"channel_type"` // text, private
	Description string `json:"description"`
}

// CreatePrivateChannelInput represents input for creating a private channel.
type CreatePrivateChannelInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InviteToChannelInput represents input for inviting to a channel.
type InviteToChannelInput struct {
	ChannelID string `json:"channel_id"`
	PeerID    string `json:"peer_id"`
}

// AcceptInviteInput represents input for accepting an invite.
type AcceptInviteInput struct {
	InviteID string `json:"invite_id"`
}

// SetDisplayNameInput represents input for setting display name.
type SetDisplayNameInput struct {
	Name string `json:"name"`
}

// ConnectInput represents input for connecting to a peer.
type ConnectInput struct {
	Addr string `json:"addr"`
}

// SaveConnectionInput represents input for saving a connection.
type SaveConnectionInput struct {
	Addr     string `json:"addr"`
	Nickname string `json:"nickname,omitempty"`
}

// ListMessagesInput represents input for listing messages.
type ListMessagesInput struct {
	ChannelID string `json:"channel_id"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}