package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-chat/internal/config"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(cfg config.DatabaseConfig) (*Store, error) {
	dir := filepath.Dir(cfg.Path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", cfg.Path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS identities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			display_name TEXT NOT NULL DEFAULT '',
			peer_id TEXT NOT NULL UNIQUE,
			private_key BLOB NOT NULL,
			public_key BLOB NOT NULL,
			avatar_color TEXT NOT NULL DEFAULT '#5865F2',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS peers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			peer_id TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL DEFAULT '',
			public_key BLOB,
			fingerprint TEXT NOT NULL DEFAULT '',
			avatar_color TEXT NOT NULL DEFAULT '#5865F2',
			status TEXT NOT NULL DEFAULT 'offline',
			trusted INTEGER NOT NULL DEFAULT 0,
			blocked INTEGER NOT NULL DEFAULT 0,
			favorite INTEGER NOT NULL DEFAULT 0,
			bio TEXT NOT NULL DEFAULT '',
			timezone TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			first_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			org_id TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			owner_peer_id TEXT NOT NULL,
			visibility TEXT NOT NULL DEFAULT 'private',
			max_members INTEGER NOT NULL DEFAULT 0,
			history_retention INTEGER NOT NULL DEFAULT 0,
			attachment_limit INTEGER NOT NULL DEFAULT 0,
			archived INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id TEXT NOT NULL UNIQUE,
			org_id TEXT,
			name TEXT NOT NULL,
			topic TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			channel_type TEXT NOT NULL DEFAULT 'text',
			category TEXT NOT NULL DEFAULT '',
			parent_category TEXT NOT NULL DEFAULT '',
			read_only INTEGER NOT NULL DEFAULT 0,
			archived INTEGER NOT NULL DEFAULT 0,
			muted INTEGER NOT NULL DEFAULT 0,
			favorite INTEGER NOT NULL DEFAULT 0,
			pinned INTEGER NOT NULL DEFAULT 0,
			slow_mode INTEGER NOT NULL DEFAULT 0,
			retention INTEGER NOT NULL DEFAULT 0,
			attachment_limit INTEGER NOT NULL DEFAULT 0,
			emoji_reactions INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS memberships (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			peer_id TEXT NOT NULL,
			org_id TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(peer_id, org_id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id TEXT NOT NULL UNIQUE,
			channel_id TEXT NOT NULL,
			sender_peer_id TEXT NOT NULL,
			content TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'text',
			encrypted INTEGER NOT NULL DEFAULT 1,
			reply_to TEXT NOT NULL DEFAULT '',
			edited INTEGER NOT NULL DEFAULT 0,
			deleted INTEGER NOT NULL DEFAULT 0,
			pinned INTEGER NOT NULL DEFAULT 0,
			delivery_state TEXT NOT NULL DEFAULT 'sent',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS attachments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			attachment_id TEXT NOT NULL UNIQUE,
			message_id TEXT,
			channel_id TEXT,
			sender_peer_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			filepath TEXT NOT NULL,
			mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			size INTEGER NOT NULL DEFAULT 0,
			hash TEXT NOT NULL DEFAULT '',
			chunk_size INTEGER NOT NULL DEFAULT 0,
			total_chunks INTEGER NOT NULL DEFAULT 0,
			received_chunks INTEGER NOT NULL DEFAULT 0,
			transfer_state TEXT NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS invites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			invite_id TEXT NOT NULL UNIQUE,
			sender_peer_id TEXT NOT NULL,
			target_peer_id TEXT NOT NULL,
			org_id TEXT NOT NULL DEFAULT '',
			channel_id TEXT NOT NULL DEFAULT '',
			invite_type TEXT NOT NULL DEFAULT 'peer',
			message TEXT NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL DEFAULT '',
			one_time INTEGER NOT NULL DEFAULT 1,
			max_uses INTEGER NOT NULL DEFAULT 0,
			use_count INTEGER NOT NULL DEFAULT 0,
			expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS reactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id TEXT NOT NULL,
			peer_id TEXT NOT NULL,
			emoji TEXT NOT NULL,
			custom_emoji TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(message_id, peer_id, emoji)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL UNIQUE,
			peer_id TEXT NOT NULL,
			enc_key BLOB NOT NULL,
			auth_key BLOB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL,
			nickname TEXT NOT NULL DEFAULT '',
			last_connected_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_peer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_peers_peer_id ON peers(peer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_memberships_peer ON memberships(peer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_memberships_org ON memberships(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_message ON reactions(message_id)`,
	}

	for _, q := range tables {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	return nil
}

func (s *Store) Now() time.Time {
	return time.Now().UTC()
}
