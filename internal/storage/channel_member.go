package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type ChannelMember struct {
	ID        int64     `json:"id"`
	ChannelID string    `json:"channel_id"`
	PeerID    string    `json:"peer_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

func (s *Store) AddChannelMember(channelID, peerID, role string) error {
	if role == "" {
		role = "member"
	}
	_, err := s.db.Exec(`INSERT INTO channel_members (channel_id, peer_id, role, joined_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(channel_id, peer_id) DO UPDATE SET role=excluded.role`,
		channelID, peerID, role, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("add channel member: %w", err)
	}
	return nil
}

func (s *Store) RemoveChannelMember(channelID, peerID string) error {
	_, err := s.db.Exec(`DELETE FROM channel_members WHERE channel_id=? AND peer_id=?`,
		channelID, peerID)
	return err
}

func (s *Store) IsChannelMember(channelID, peerID string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM channel_members WHERE channel_id=? AND peer_id=?`,
		channelID, peerID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check channel member: %w", err)
	}
	return count > 0, nil
}

func (s *Store) ListChannelMembers(channelID string) ([]*ChannelMember, error) {
	rows, err := s.db.Query(`SELECT id, channel_id, peer_id, role, joined_at
		FROM channel_members WHERE channel_id=? ORDER BY joined_at`, channelID)
	if err != nil {
		return nil, fmt.Errorf("list channel members: %w", err)
	}
	defer rows.Close()

	var members []*ChannelMember
	for rows.Next() {
		m := &ChannelMember{}
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.PeerID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan channel member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *Store) ListPeerChannels(peerID string) ([]*Channel, error) {
	rows, err := s.db.Query(`SELECT c.id, c.channel_id, c.org_id, c.name, c.topic, c.description,
		c.channel_type, c.category, c.parent_category, c.read_only, c.archived, c.muted, c.favorite,
		c.pinned, c.slow_mode, c.retention, c.attachment_limit, c.emoji_reactions, c.created_at, c.updated_at
		FROM channels c
		INNER JOIN channel_members cm ON c.channel_id = cm.channel_id
		WHERE cm.peer_id = ?
		ORDER BY c.name`, peerID)
	if err != nil {
		return nil, fmt.Errorf("list peer channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description,
			&ch.ChannelType, &ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted,
			&ch.Favorite, &ch.Pinned, &ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions,
			&ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

type Invite struct {
	ID           int64      `json:"id"`
	InviteID     string     `json:"invite_id"`
	SenderPeerID string     `json:"sender_peer_id"`
	TargetPeerID string     `json:"target_peer_id"`
	OrgID        string     `json:"org_id"`
	ChannelID    string     `json:"channel_id"`
	InviteType   string     `json:"invite_type"`
	Message      string     `json:"message"`
	OneTime      bool       `json:"one_time"`
	MaxUses      int        `json:"max_uses"`
	UseCount     int        `json:"use_count"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (s *Store) SaveInvite(inv *Invite) error {
	_, err := s.db.Exec(`INSERT INTO invites (invite_id, sender_peer_id, target_peer_id, org_id, channel_id, invite_type, message, one_time, max_uses, use_count, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(invite_id) DO UPDATE SET
			use_count=excluded.use_count`,
		inv.InviteID, inv.SenderPeerID, inv.TargetPeerID, inv.OrgID, inv.ChannelID,
		inv.InviteType, inv.Message, inv.OneTime, inv.MaxUses, inv.UseCount,
		inv.ExpiresAt, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("save invite: %w", err)
	}
	return nil
}

func (s *Store) GetInvite(inviteID string) (*Invite, error) {
	inv := &Invite{}
	err := s.db.QueryRow(`SELECT id, invite_id, sender_peer_id, target_peer_id, org_id, channel_id, invite_type, message, one_time, max_uses, use_count, expires_at, created_at
		FROM invites WHERE invite_id=?`, inviteID).Scan(
		&inv.ID, &inv.InviteID, &inv.SenderPeerID, &inv.TargetPeerID, &inv.OrgID, &inv.ChannelID,
		&inv.InviteType, &inv.Message, &inv.OneTime, &inv.MaxUses, &inv.UseCount, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get invite: %w", err)
	}
	return inv, nil
}

func (s *Store) ListPendingInvites(targetPeerID string) ([]*Invite, error) {
	rows, err := s.db.Query(`SELECT id, invite_id, sender_peer_id, target_peer_id, org_id, channel_id, invite_type, message, one_time, max_uses, use_count, expires_at, created_at
		FROM invites WHERE target_peer_id=? ORDER BY created_at`, targetPeerID)
	if err != nil {
		return nil, fmt.Errorf("list pending invites: %w", err)
	}
	defer rows.Close()

	var invites []*Invite
	for rows.Next() {
		inv := &Invite{}
		if err := rows.Scan(&inv.ID, &inv.InviteID, &inv.SenderPeerID, &inv.TargetPeerID, &inv.OrgID, &inv.ChannelID,
			&inv.InviteType, &inv.Message, &inv.OneTime, &inv.MaxUses, &inv.UseCount, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan invite: %w", err)
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (s *Store) DeleteInvite(inviteID string) error {
	_, err := s.db.Exec(`DELETE FROM invites WHERE invite_id=?`, inviteID)
	return err
}
