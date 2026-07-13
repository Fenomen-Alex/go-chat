package channel

import (
	"fmt"
	"time"

	"go-chat/internal/storage"
)

type Manager struct {
	store *storage.Store
}

func NewManager(store *storage.Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) CreateChannel(orgID, name, channelType, category, description string) (*storage.Channel, error) {
	existing, err := m.store.GetChannelByName(name)
	if err != nil {
		return nil, fmt.Errorf("check existing channel: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("channel '%s' already exists", name)
	}

	channelID := fmt.Sprintf("ch_%d", time.Now().UnixNano())

	ch := &storage.Channel{
		ChannelID:   channelID,
		OrgID:       orgID,
		Name:        name,
		Description: description,
		ChannelType: channelType,
		Category:    category,
	}

	if err := m.store.SaveChannel(ch); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	return ch, nil
}

func (m *Manager) CreatePrivateChannel(name, description, creatorPeerID string) (*storage.Channel, error) {
	existing, err := m.store.GetChannelByName(name)
	if err != nil {
		return nil, fmt.Errorf("check existing channel: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("channel '%s' already exists", name)
	}

	channelID := fmt.Sprintf("ch_%d", time.Now().UnixNano())

	ch := &storage.Channel{
		ChannelID:   channelID,
		Name:        name,
		Description: description,
		ChannelType: "private",
	}

	if err := m.store.SaveChannel(ch); err != nil {
		return nil, fmt.Errorf("save channel: %w", err)
	}

	if err := m.store.AddChannelMember(channelID, creatorPeerID, "owner"); err != nil {
		return nil, fmt.Errorf("add creator as owner: %w", err)
	}

	return ch, nil
}

func (m *Manager) AddMember(channelID, peerID, role string) error {
	return m.store.AddChannelMember(channelID, peerID, role)
}

func (m *Manager) RemoveMember(channelID, peerID string) error {
	return m.store.RemoveChannelMember(channelID, peerID)
}

func (m *Manager) IsMember(channelID, peerID string) (bool, error) {
	return m.store.IsChannelMember(channelID, peerID)
}

func (m *Manager) ListMembers(channelID string) ([]*storage.ChannelMember, error) {
	return m.store.ListChannelMembers(channelID)
}

func (m *Manager) GetChannel(channelID string) (*storage.Channel, error) {
	return m.store.GetChannel(channelID)
}

func (m *Manager) ListChannels(orgID string) ([]*storage.Channel, error) {
	return m.store.ListChannels(orgID)
}

func (m *Manager) ListDMChannels() ([]*storage.Channel, error) {
	return m.store.ListDMChannels()
}

func (m *Manager) ArchiveChannel(channelID string) error {
	return m.store.ArchiveChannel(channelID)
}
