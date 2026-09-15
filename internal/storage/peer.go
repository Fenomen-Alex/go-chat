package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type Peer struct {
	ID          int64     `json:"id"`
	PeerID      string    `json:"peer_id"`
	DisplayName string    `json:"display_name"`
	PublicKey   []byte    `json:"public_key"`
	Fingerprint string    `json:"fingerprint"`
	AvatarColor string    `json:"avatar_color"`
	Status      string    `json:"status"`
	Trusted     bool      `json:"trusted"`
	Blocked     bool      `json:"blocked"`
	Favorite    bool      `json:"favorite"`
	Bio         string    `json:"bio"`
	Timezone    string    `json:"timezone"`
	Notes       string    `json:"notes"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Store) SavePeer(p *Peer) error {
	_, err := s.db.Exec(`INSERT INTO peers (peer_id, display_name, public_key, fingerprint, avatar_color, status, trusted, blocked, favorite, bio, timezone, notes, first_seen, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(peer_id) DO UPDATE SET
			display_name=excluded.display_name,
			public_key=excluded.public_key,
			fingerprint=excluded.fingerprint,
			avatar_color=excluded.avatar_color,
			status=excluded.status,
			trusted=excluded.trusted,
			blocked=excluded.blocked,
			favorite=excluded.favorite,
			bio=excluded.bio,
			timezone=excluded.timezone,
			notes=excluded.notes,
			last_seen=excluded.last_seen,
			updated_at=excluded.updated_at`,
		p.PeerID, p.DisplayName, p.PublicKey, p.Fingerprint, p.AvatarColor, p.Status, p.Trusted, p.Blocked, p.Favorite,
		p.Bio, p.Timezone, p.Notes, p.FirstSeen, p.LastSeen, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save peer: %w", err)
	}
	return nil
}

func (s *Store) GetPeer(peerID string) (*Peer, error) {
	p := &Peer{}
	err := s.db.QueryRow(`SELECT id, peer_id, display_name, public_key, fingerprint, avatar_color, status, trusted, blocked, favorite, bio, timezone, notes, first_seen, last_seen, created_at, updated_at
		FROM peers WHERE peer_id=?`, peerID).Scan(
		&p.ID, &p.PeerID, &p.DisplayName, &p.PublicKey, &p.Fingerprint, &p.AvatarColor, &p.Status,
		&p.Trusted, &p.Blocked, &p.Favorite, &p.Bio, &p.Timezone, &p.Notes,
		&p.FirstSeen, &p.LastSeen, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get peer: %w", err)
	}
	return p, nil
}

func (s *Store) GetPeerByDisplayName(name string) (*Peer, error) {
	p := &Peer{}
	err := s.db.QueryRow(`SELECT id, peer_id, display_name, public_key, fingerprint, avatar_color, status, trusted, blocked, favorite, bio, timezone, notes, first_seen, last_seen, created_at, updated_at
		FROM peers WHERE display_name=? LIMIT 1`, name).Scan(
		&p.ID, &p.PeerID, &p.DisplayName, &p.PublicKey, &p.Fingerprint, &p.AvatarColor, &p.Status,
		&p.Trusted, &p.Blocked, &p.Favorite, &p.Bio, &p.Timezone, &p.Notes,
		&p.FirstSeen, &p.LastSeen, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get peer by display name: %w", err)
	}
	return p, nil
}

func (s *Store) ListPeers() ([]*Peer, error) {
	rows, err := s.db.Query(`SELECT id, peer_id, display_name, public_key, fingerprint, avatar_color, status, trusted, blocked, favorite, bio, timezone, notes, first_seen, last_seen, created_at, updated_at FROM peers ORDER BY display_name`)
	if err != nil {
		return nil, fmt.Errorf("list peers: %w", err)
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(&p.ID, &p.PeerID, &p.DisplayName, &p.PublicKey, &p.Fingerprint, &p.AvatarColor, &p.Status,
			&p.Trusted, &p.Blocked, &p.Favorite, &p.Bio, &p.Timezone, &p.Notes,
			&p.FirstSeen, &p.LastSeen, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan peer: %w", err)
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func (s *Store) UpdatePeerStatus(peerID, status string) error {
	_, err := s.db.Exec(`UPDATE peers SET status=?, last_seen=?, updated_at=? WHERE peer_id=?`,
		status, time.Now().UTC(), time.Now().UTC(), peerID)
	return err
}
