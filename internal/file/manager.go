package file

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Transfer() error {
	return nil // TODO(milestone-2): implement file transfer
}

func (m *Manager) Verify() error {
	return nil // TODO(milestone-2): implement file integrity verification
}
