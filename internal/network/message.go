package network

import "time"

type Message struct {
	Type          string `json:"type"`
	SenderID      string `json:"sender_id"`
	ChannelID     string `json:"channel_id,omitempty"`
	OrgID         string `json:"org_id,omitempty"`
	MessageID     string `json:"message_id,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ChannelType   string `json:"channel_type,omitempty"`
	InviteID      string `json:"invite_id,omitempty"`
	TargetPeerID  string `json:"target_peer_id,omitempty"`
	MemberPeerID  string `json:"member_peer_id,omitempty"`
	MemberRole    string `json:"member_role,omitempty"`
	ReplyTo       string `json:"reply_to,omitempty"`
	Encrypted     bool   `json:"encrypted,omitempty"`
	EncryptedData []byte `json:"encrypted_data,omitempty"`
	Timestamp     int64  `json:"timestamp"`
}

func NewMessage(msgType, senderID string) *Message {
	return &Message{
		Type:      msgType,
		SenderID:  senderID,
		Timestamp: time.Now().UnixMilli(),
	}
}
