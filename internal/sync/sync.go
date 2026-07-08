package sync

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Sync() error {
	return nil // TODO(milestone-2): implement sync reconciliation
}

func (m *Manager) Reconcile() error {
	return nil // TODO(milestone-2): implement conflict resolution
}
