package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type Channel struct {
	ID              int64     `json:"id"`
	ChannelID       string    `json:"channel_id"`
	OrgID           string    `json:"org_id"`
	Name            string    `json:"name"`
	Topic           string    `json:"topic"`
	Description     string    `json:"description"`
	ChannelType     string    `json:"channel_type"`
	Category        string    `json:"category"`
	ParentCategory  string    `json:"parent_category"`
	ReadOnly        bool      `json:"read_only"`
	Archived        bool      `json:"archived"`
	Muted           bool      `json:"muted"`
	Favorite        bool      `json:"favorite"`
	Pinned          bool      `json:"pinned"`
	SlowMode        int       `json:"slow_mode"`
	Retention       int       `json:"retention"`
	AttachmentLimit int64     `json:"attachment_limit"`
	EmojiReactions  bool      `json:"emoji_reactions"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *Store) SaveChannel(ch *Channel) error {
	_, err := s.db.Exec(`INSERT INTO channels (channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(channel_id) DO UPDATE SET
			name=excluded.name, topic=excluded.topic, description=excluded.description,
			channel_type=excluded.channel_type, category=excluded.category,
			read_only=excluded.read_only, archived=excluded.archived, muted=excluded.muted,
			favorite=excluded.favorite, pinned=excluded.pinned,
			slow_mode=excluded.slow_mode, retention=excluded.retention,
			attachment_limit=excluded.attachment_limit, emoji_reactions=excluded.emoji_reactions,
			updated_at=excluded.updated_at`,
		ch.ChannelID, ch.OrgID, ch.Name, ch.Topic, ch.Description, ch.ChannelType, ch.Category,
		ch.ParentCategory, ch.ReadOnly, ch.Archived, ch.Muted, ch.Favorite, ch.Pinned,
		ch.SlowMode, ch.Retention, ch.AttachmentLimit, ch.EmojiReactions, ch.CreatedAt, ch.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save channel: %w", err)
	}
	return nil
}

func (s *Store) GetChannelByName(name string) (*Channel, error) {
	ch := &Channel{}
	err := s.db.QueryRow(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels WHERE name=? LIMIT 1`, name).Scan(
		&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
		&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
		&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel by name: %w", err)
	}
	return ch, nil
}

func (s *Store) GetChannel(channelID string) (*Channel, error) {
	ch := &Channel{}
	err := s.db.QueryRow(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels WHERE channel_id=?`, channelID).Scan(
		&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
		&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
		&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel: %w", err)
	}
	return ch, nil
}

func (s *Store) ListChannels(orgID string) ([]*Channel, error) {
	rows, err := s.db.Query(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels WHERE org_id=? ORDER BY name`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
			&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
			&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (s *Store) ListAllChannels() ([]*Channel, error) {
	rows, err := s.db.Query(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list all channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
			&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
			&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (s *Store) ArchiveChannel(channelID string) error {
	_, err := s.db.Exec(`UPDATE channels SET archived=1, updated_at=? WHERE channel_id=?`,
		time.Now().UTC(), channelID)
	return err
}

func (s *Store) ListDMChannels() ([]*Channel, error) {
	rows, err := s.db.Query(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels WHERE channel_type='dm' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list dm channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
			&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
			&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (s *Store) ListPrivateChannels() ([]*Channel, error) {
	rows, err := s.db.Query(`SELECT id, channel_id, org_id, name, topic, description, channel_type, category, parent_category, read_only, archived, muted, favorite, pinned, slow_mode, retention, attachment_limit, emoji_reactions, created_at, updated_at
		FROM channels WHERE channel_type='private' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list private channels: %w", err)
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.ChannelID, &ch.OrgID, &ch.Name, &ch.Topic, &ch.Description, &ch.ChannelType,
			&ch.Category, &ch.ParentCategory, &ch.ReadOnly, &ch.Archived, &ch.Muted, &ch.Favorite, &ch.Pinned,
			&ch.SlowMode, &ch.Retention, &ch.AttachmentLimit, &ch.EmojiReactions, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
