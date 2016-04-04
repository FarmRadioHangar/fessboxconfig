package device

import "github.com/tarm/serial"

type Manager struct {
	devices map[string]serial.Config
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
