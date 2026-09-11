package network

import (
	"crypto/ed25519"
	"encoding/json"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func SignMessage(privateKey []byte, msg *Message) {
	canonical := struct {
		Type         string `json:"type"`
		SenderID     string `json:"sender_id"`
		ChannelID    string `json:"channel_id,omitempty"`
		OrgID        string `json:"org_id,omitempty"`
		MessageID    string `json:"message_id,omitempty"`
		Content      string `json:"content,omitempty"`
		ContentType  string `json:"content_type,omitempty"`
		ChannelType  string `json:"channel_type,omitempty"`
		InviteID     string `json:"invite_id,omitempty"`
		TargetPeerID string `json:"target_peer_id,omitempty"`
		MemberPeerID string `json:"member_peer_id,omitempty"`
		MemberRole   string `json:"member_role,omitempty"`
		ReplyTo      string `json:"reply_to,omitempty"`
		Timestamp    int64  `json:"timestamp"`
	}{
		Type:         msg.Type,
		SenderID:     msg.SenderID,
		ChannelID:    msg.ChannelID,
		OrgID:        msg.OrgID,
		MessageID:    msg.MessageID,
		Content:      msg.Content,
		ContentType:  msg.ContentType,
		ChannelType:  msg.ChannelType,
		InviteID:     msg.InviteID,
		TargetPeerID: msg.TargetPeerID,
		MemberPeerID: msg.MemberPeerID,
		MemberRole:   msg.MemberRole,
		ReplyTo:      msg.ReplyTo,
		Timestamp:    msg.Timestamp,
	}
	data, _ := json.Marshal(canonical)
	msg.Signature = ed25519.Sign(privateKey, data)
}

func VerifyMessageSignature(publicKey []byte, msg *Message) bool {
	if len(msg.Signature) == 0 {
		return false
	}
	canonical := struct {
		Type         string `json:"type"`
		SenderID     string `json:"sender_id"`
		ChannelID    string `json:"channel_id,omitempty"`
		OrgID        string `json:"org_id,omitempty"`
		MessageID    string `json:"message_id,omitempty"`
		Content      string `json:"content,omitempty"`
		ContentType  string `json:"content_type,omitempty"`
		ChannelType  string `json:"channel_type,omitempty"`
		InviteID     string `json:"invite_id,omitempty"`
		TargetPeerID string `json:"target_peer_id,omitempty"`
		MemberPeerID string `json:"member_peer_id,omitempty"`
		MemberRole   string `json:"member_role,omitempty"`
		ReplyTo      string `json:"reply_to,omitempty"`
		Timestamp    int64  `json:"timestamp"`
	}{
		Type:         msg.Type,
		SenderID:     msg.SenderID,
		ChannelID:    msg.ChannelID,
		OrgID:        msg.OrgID,
		MessageID:    msg.MessageID,
		Content:      msg.Content,
		ContentType:  msg.ContentType,
		ChannelType:  msg.ChannelType,
		InviteID:     msg.InviteID,
		TargetPeerID: msg.TargetPeerID,
		MemberPeerID: msg.MemberPeerID,
		MemberRole:   msg.MemberRole,
		ReplyTo:      msg.ReplyTo,
		Timestamp:    msg.Timestamp,
	}
	data, _ := json.Marshal(canonical)
	return ed25519.Verify(publicKey, data, msg.Signature)
}

func GetPublicKeyFromPeerstore(host host.Host, peerIDStr string) ([]byte, bool) {
	pID, err := peer.Decode(peerIDStr)
	if err != nil {
		return nil, false
	}
	if pubKey := host.Peerstore().PubKey(pID); pubKey != nil {
		if raw, ok := rawEd25519PublicKey(pubKey); ok {
			return raw, true
		}
	}
	for _, c := range host.Network().ConnsToPeer(pID) {
		if pubKey := c.RemotePublicKey(); pubKey != nil {
			if raw, ok := rawEd25519PublicKey(pubKey); ok {
				return raw, true
			}
		}
	}
	return nil, false
}

func rawEd25519PublicKey(pubKey libp2pcrypto.PubKey) ([]byte, bool) {
	if ed25519Pub, ok := pubKey.(*libp2pcrypto.Ed25519PublicKey); ok {
		raw, err := ed25519Pub.Raw()
		if err != nil {
			return nil, false
		}
		return raw, true
	}
	return nil, false
}
