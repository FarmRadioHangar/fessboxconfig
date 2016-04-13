package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/jochenvg/go-udev"
	"github.com/tarm/serial"
)

var modemCommands = struct {
	IMEI, IMSI string
}{
	"AT+GSN", "AT+CIMI",
}

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
	monitor *udev.Monitor
	done    chan struct{}
	stop    chan struct{}
}

// New returns a new Manager instance
func New() *Manager {
	return &Manager{
		devices: make(map[string]serial.Config),
		done:    make(chan struct{}),
		stop:    make(chan struct{}),
	}
}

// Init initializes the manager. This involves creating a new goroutine to watch
// over the changes detected by udev for any device interaction with the system.
//
// The only interesting device actions are add and reomove for adding and
// removing devices respctively.
func (m *Manager) Init() {
	u := udev.Udev{}
	monitor := u.NewMonitorFromNetlink("udev")
	monitor.FilterAddMatchTag("systemd")
	devCh, err := monitor.DeviceChan(m.done)
	if err != nil {
		panic(err)
	}
	m.monitor = monitor
	go func() {
	stop:
		for {
			select {
			case d := <-devCh:
				switch d.Action() {
				case "add":
					dpath := filepath.Join("/dev", filepath.Base(d.Devpath()))
					m.AddDevice(dpath)
					fmt.Printf(" new device added  %s\n", dpath)
					m.reload()
				case "remove":
					dpath := filepath.Join("/dev", filepath.Base(d.Devpath()))
					fmt.Printf(" %s was removed\n", dpath)
					m.RemoveDevice(dpath)
					m.reload()
				default:
					fmt.Println(d.Action())
				}
			case quit := <-m.stop:
				m.done <- quit
				break stop
			}
		}
	}()
}

// AddDevice adds device name to the manager
func (m *Manager) AddDevice(name string) error {
	cfg := serial.Config{Name: name, Baud: 9600, ReadTimeout: time.Second}
	m.mu.Lock()
	m.devices[name] = cfg
	m.mu.Unlock()
	return nil
}

// List serves the list of current devices. The list wont cover all devices ,
// only the significant ones( modems for now)
func (m *Manager) List(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	for i := 0; i < len(m.conn); i++ {
		c := m.conn[i]
		data[c.device.Name] = c.imei
	}
	json.NewEncoder(w).Encode(data)
	w.Header().Set("Content-Type", "application/json")
}

// RunComand runs commands to the exposed devices over serial ports
func (m *Manager) RunCommand(w http.ResponseWriter, r *http.Request) {
}

// RemoveDevice removes device name from the manager
func (m *Manager) RemoveDevice(name string) error {
	m.mu.RLock()
	delete(m.devices, name)
	m.mu.RUnlock()
	return nil
}

// Exec executes command over serial port for devices which have open ports
func (m *Manager) Exec(name string, cmds string, isIMEI bool) ([]byte, error) {
	for i := 0; i < len(m.conn); i++ {
		c := m.conn[i]
		if isIMEI {
			if c.imei != name {
				continue
			}
		}
		if c.device.Name != name {
			continue
		}
		return c.Run(cmds)
	}
	return nil, errors.New("no device found")
}

// close all ports that are open for the devices
func (m *Manager) releaseAllPorts() {
	for _, c := range m.conn {
		err := c.Close()
		if err != nil {
			log.Printf("[ERR] closing port %s %v\n", c.device.Name, err)
		}
	}
}

func (m *Manager) reload() {
	m.releaseAllPorts()
	var conns []*Conn
	for _, v := range m.devices {
		conn := &Conn{device: v}
		imei, err := conn.Run(modemCommands.IMEI)
		if err != nil {
			_ = conn.Close()
			continue
		}
		i, err := cleanIMEI(imei)
		if err != nil {
			_ = conn.Close()
			continue
		}
		conn.imei = string(i)
		conns = append(conns, conn)
	}
	m.conn = conns
}

func cleanIMEI(src []byte) ([]byte, error) {
	i := bytes.Index(src, []byte("OK"))
	if i == -1 {
		return nil, errors.New("not okay")
	}
	ns := bytes.TrimSpace(src[:i])
	return ns, nil
}

//Close shuts down the device manager. This makes sure the udev monitor is
//closed and all goroutines are properly exited.
func (m *Manager) Close() {
	m.stop <- struct{}{}
}

// Conn is a device serial connection
type Conn struct {
	device serial.Config
	imei   string
	port   *serial.Port
	isOpen bool
}

// Open opens a serial port to the undelying device
func (c *Conn) Open() error {
	p, err := serial.OpenPort(&c.device)
	if err != nil {
		return err
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

// Write wites b to the serieal port
func (c *Conn) Write(b []byte) (int, error) {
	return c.port.Write(b)
}

// Read reads from serial port
func (c *Conn) Read(b []byte) (int, error) {
	return c.port.Read(b)
}

// Exec sends the command over serial port and rrturns the response. If the port
// is closed it is opened  before sending the command.
func (c *Conn) Exec(cmd string) ([]byte, error) {
	if !c.isOpen {
		fmt.Println("Opening port")
		err := c.Open()
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = c.port.Flush() }()
	_, err := c.Write([]byte(cmd))
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 128)
	_, err = c.Read(buf)
	if err != nil {
		return nil, err
	}
	if !bytes.Contains(buf, []byte("OK")) {
		return nil, errors.New(" not Okay")
	}
	return buf, nil
}

// Run helper for Exec that adds \r to the command
func (c *Conn) Run(cmd string) ([]byte, error) {
	return c.Exec(fmt.Sprintf("%s \r", cmd))
}
