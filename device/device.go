package device

import (
	"sync"

	"github.com/tarm/serial"
)

type Manager struct {
	devices map[string]serial.Config
	mu      sync.RWMutex
}

func (m *Manager) AddDevice(name string) error {
	cfg := serial.Config{Name: name}
	m.mu.Lock()
	m.devices[name] = cfg
	m.mu.Unlock()
	return nil
}

func (m *Manager) RemoveDevice(name string) error {
	m.mu.RLock()
	delete(m.devices, name)
	m.mu.RUnlock()
	return nil
}

type Conn struct {
	device serial.Config
	port   *serial.Port
	isOpen bool
}

func (c *Conn) Open() error {
	p, err := serial.OpenPort(&c.device)
	if err != nil {
		return nil
	}
	c.port = p
	c.isOpen = true
	return nil
}

// Close closes the port helt by *Conn.
func (c *Conn) Close() error {
	if c.isOpen {
		return c.port.Close()
	}
	return nil
}
