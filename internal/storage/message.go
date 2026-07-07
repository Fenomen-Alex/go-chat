package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type Message struct {
	ID           int64     `json:"id"`
	MessageID    string    `json:"message_id"`
	ChannelID    string    `json:"channel_id"`
	SenderPeerID string    `json:"sender_peer_id"`
	Content      string    `json:"content"`
	ContentType  string    `json:"content_type"`
	Encrypted    bool      `json:"encrypted"`
	ReplyTo      string    `json:"reply_to"`
	Edited       bool      `json:"edited"`
	Deleted      bool      `json:"deleted"`
	Pinned       bool      `json:"pinned"`
	DeliveryState string   `json:"delivery_state"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Store) SaveMessage(msg *Message) error {
	_, err := s.db.Exec(`INSERT INTO messages (message_id, channel_id, sender_peer_id, content, content_type, encrypted, reply_to, edited, deleted, pinned, delivery_state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(message_id) DO UPDATE SET
			content=excluded.content,
			edited=excluded.edited,
			deleted=excluded.deleted,
			pinned=excluded.pinned,
			delivery_state=excluded.delivery_state,
			updated_at=excluded.updated_at`,
		msg.MessageID, msg.ChannelID, msg.SenderPeerID, msg.Content, msg.ContentType, msg.Encrypted,
		msg.ReplyTo, msg.Edited, msg.Deleted, msg.Pinned, msg.DeliveryState, msg.CreatedAt, msg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}

func (s *Store) GetMessage(messageID string) (*Message, error) {
	msg := &Message{}
	err := s.db.QueryRow(`SELECT id, message_id, channel_id, sender_peer_id, content, content_type, encrypted, reply_to, edited, deleted, pinned, delivery_state, created_at, updated_at
		FROM messages WHERE message_id=?`, messageID).Scan(
		&msg.ID, &msg.MessageID, &msg.ChannelID, &msg.SenderPeerID, &msg.Content, &msg.ContentType,
		&msg.Encrypted, &msg.ReplyTo, &msg.Edited, &msg.Deleted, &msg.Pinned, &msg.DeliveryState,
		&msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get message: %w", err)
	}
	return msg, nil
}

func (s *Store) ListMessages(channelID string, limit, offset int) ([]*Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id, message_id, channel_id, sender_peer_id, content, content_type, encrypted, reply_to, edited, deleted, pinned, delivery_state, created_at, updated_at
		FROM messages WHERE channel_id=? AND deleted=0 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		channelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(&msg.ID, &msg.MessageID, &msg.ChannelID, &msg.SenderPeerID, &msg.Content, &msg.ContentType,
			&msg.Encrypted, &msg.ReplyTo, &msg.Edited, &msg.Deleted, &msg.Pinned, &msg.DeliveryState,
			&msg.CreatedAt, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (s *Store) ListAllMessages(limit int) ([]*Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id, message_id, channel_id, sender_peer_id, content, content_type, encrypted, reply_to, edited, deleted, pinned, delivery_state, created_at, updated_at
		FROM messages WHERE deleted=0 ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list all messages: %w", err)
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(&msg.ID, &msg.MessageID, &msg.ChannelID, &msg.SenderPeerID, &msg.Content, &msg.ContentType,
			&msg.Encrypted, &msg.ReplyTo, &msg.Edited, &msg.Deleted, &msg.Pinned, &msg.DeliveryState,
			&msg.CreatedAt, &msg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (s *Store) DeleteMessage(messageID string) error {
	_, err := s.db.Exec(`UPDATE messages SET deleted=1, updated_at=? WHERE message_id=?`,
		time.Now().UTC(), messageID)
	return err
}

func (s *Store) PinMessage(messageID string, pinned bool) error {
	_, err := s.db.Exec(`UPDATE messages SET pinned=?, updated_at=? WHERE message_id=?`,
		pinned, time.Now().UTC(), messageID)
	return err
}
