//go:build !linux

package sandbox

func (m *Manager) initializePlatformNetworking() error {
	return nil
}
