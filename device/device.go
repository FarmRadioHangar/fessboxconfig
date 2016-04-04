package device

import (
	"sync"

	"github.com/tarm/serial"
)

// Manager manages devices that are plugged into the system. It supports auto
// detection of devices.
//
// Serial ports are opened each for a device, and a clean API for communicating
// is provided via Read, Write and Flush methods.
//
// The devices are monitored via udev, and any changes that requires reloading
// of the  ports are handled by reloading the ports to the devices.
//
// This is safe to use concurrently in multiple goroutines
type Manager struct {
	devices map[string]serial.Config
	conn    []*Conn
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
