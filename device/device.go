package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

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
	modems  map[string]*Modem
	mu      sync.RWMutex
	monitor *udev.Monitor
	done    chan struct{}
	stop    chan struct{}
}

// New returns a new Manager instance
func New() *Manager {
	return &Manager{
		devices: make(map[string]serial.Config),
		modems:  make(map[string]*Modem),
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
				dpath := filepath.Join("/dev", filepath.Base(d.Devpath()))
				switch d.Action() {
				case "add":
					fmt.Println(" add ", dpath)
					m.AddDevice(d)
					fmt.Println(" done adding ", dpath)
				case "remove":
					m.RemoveDevice(dpath)
				}
			case quit := <-m.stop:
				m.done <- quit
				break stop
			}
		}
	}()

}

// AddDevice adds device name to the manager
//
// WARNING: The way modems are picked is a hack. It asserts that the modem with
// the lowest tty number is the control modem( Which I'm not so sure is always
// correct).
//
// TODO: comeup with a proper way to identify modems
func (m *Manager) AddDevice(d *udev.Device) error {
	err := m.addDevice(d)
	if err != nil {
		return err
	}
	return m.Symlink()
}
func (m *Manager) addDevice(d *udev.Device) error {
	name := filepath.Join("/dev", filepath.Base(d.Devpath()))
	cfg := serial.Config{Name: name, Baud: 9600, ReadTimeout: 10 * time.Second}
	conn := &Conn{device: cfg}
	if strings.Contains(name, "ttyUSB") {
		fmt.Println("checking modem")
		modem, err := newModem(conn)
		if err != nil {
			return err
		}
		if mm, ok := m.getModem(modem.IMEI); ok {
			n1, err := getttyNum(mm.Path)
			if err != nil {
				return err
			}
			n2, err := getttyNum(modem.Path)
			if err != nil {
				return err
			}
			if n1 > n2 {
				m.setModem(modem)
			}
			return nil
		}
		m.setModem(modem)
	}
	return nil
}

func (m *Manager) Symlink() error {
	for _, v := range m.modems {
		err := v.Symlink()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) setModem(mod *Modem) {
	m.mu.Lock()
	m.modems[mod.IMEI] = mod
	m.mu.Unlock()
}

func (m *Manager) getModem(imei string) (*Modem, bool) {
	m.mu.RLock()
	mod, ok := m.modems[imei]
	m.mu.RUnlock()
	return mod, ok
}

func getttyNum(tty string) (int, error) {
	b := filepath.Base(tty)
	b = strings.TrimPrefix(b, "ttyUSB")
	return strconv.Atoi(b)
}

// List serves the list of current devices. The list wont cover all devices ,
// only the significant ones( modems for now)
func (m *Manager) List(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	json.NewEncoder(w).Encode(m.modems)
	m.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
}

// RunComand runs commands to the exposed devices over serial ports
func (m *Manager) RunCommand(w http.ResponseWriter, r *http.Request) {
}

// RemoveDevice removes device name from the manager
func (m *Manager) RemoveDevice(name string) error {
	return nil
}

type Modem struct {
	IMEI         string `json:"imei"`
	IMSI         string `json:"imsi"`
	Manufacturer string `json:"manufacturer"`
	Path         string `json:"tty"`
	conn         *Conn
}

// Symlink adds symlink to the  modem. The symlink links the tty to the IMEI
// number of the modem.
func (m *Modem) Symlink() error {
	return os.Symlink(m.Path, fmt.Sprintf("/dev/%s", m.IMEI))
}

func newModem(c *Conn) (*Modem, error) {
	m := &Modem{}
	ich := time.After(10 * time.Second)
STOP:
	for {
		select {
		case <-ich:
			break STOP
		default:
			imei, err := c.Run(modemCommands.IMEI)
			if err != nil {
				continue
			}
			i, err := cleanResult(imei)
			if err != nil {
				continue
			}
			im := string(i)
			if !isNumber(im) {
				continue
			}
			m.IMEI = im
			break STOP
		}
	}
	if m.IMEI == "" {
		return nil, errors.New("no imei")
	}
	// we make sure we obtain the sim card information.
	var wg sync.WaitGroup
	done := time.After(20 * time.Second)
	wg.Add(1)
	go func() {
		defer wg.Done()
	END:
		for {
			select {
			case <-done:
				fmt.Println("Timed out")
				break END
			default:
				imsi, err := c.Run("AT+CIMI")
				if err != nil {
					continue
				}
				s, err := cleanResult(imsi)
				if err != nil {
					continue
				}
				ss := string(s)
				if !isNumber(ss) {
					continue
				}
				if m.IMEI == ss {
					continue
				}
				fmt.Println(string(s))
				m.IMSI = ss
				break END
			}
		}
	}()
	wg.Wait()
	if m.IMSI == "" {
		return nil, errors.New(" can't find IMSI")
	}
	m.conn = c
	m.Path = c.device.Name
	return m, nil
}

func isNumber(src string) bool {
	for _, v := range src {
		if !unicode.IsDigit(v) {
			return false
		}
	}
	return true
}

func (m *Manager) reload() {
	for _, v := range m.devices {
		conn := &Conn{device: v}
		modem, err := newModem(conn)
		if err != nil {
			continue
		}
		fmt.Println(*modem)
		m.modems[modem.IMEI] = modem
	}
}

func cleanResult(src []byte) ([]byte, error) {
	i := bytes.Index(src, []byte("OK"))
	if i == -1 {
		return nil, errors.New("not okay")
	}
	ns := bytes.TrimSpace(src[:i])
	ch, _ := utf8.DecodeRune(ns)
	if unicode.IsLetter(ch) {
		at := bytes.Index(ns, []byte("AT"))
		if at != -1 {
			i := bytes.IndexRune(ns[at:], '\r')
			if i > 0 {
				return bytes.TrimSpace(ns[at+i:]), nil
			}
		}
	}
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
		err := c.Open()
		if err != nil {
			return nil, err
		}
	}
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
		return nil, errors.New("command " + string(cmd) + " xeite without OK" + " got " + string(buf))
	}
	_ = c.port.Flush()
	_ = c.port.Close()
	c.isOpen = false
	return buf, nil
}

// Run helper for Exec that adds \r to the command
func (c *Conn) Run(cmd string) ([]byte, error) {
	return c.Exec(fmt.Sprintf("%s \r", cmd))
}
