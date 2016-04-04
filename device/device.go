package device

import "github.com/tarm/serial"

type Manager struct {
	devices map[string]serial.Config
}
