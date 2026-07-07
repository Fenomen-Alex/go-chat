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

func (m *Manager) GetChannel(channelID string) (*storage.Channel, error) {
	return m.store.GetChannel(channelID)
}

func (m *Manager) ListChannels(orgID string) ([]*storage.Channel, error) {
	return m.store.ListChannels(orgID)
}

func (m *Manager) ArchiveChannel(channelID string) error {
	return m.store.ArchiveChannel(channelID)
}
