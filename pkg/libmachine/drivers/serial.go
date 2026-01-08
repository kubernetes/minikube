package drivers

import (
	"sync"

	"encoding/json"

	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

var stdLock = &sync.Mutex{}

// SerialDriver is a wrapper struct which is used to ensure that RPC calls
// to a driver only occur one at a time.
// Some providers, e.g. virtualbox, should not run driver operations at the
// same time as other driver instances of the same type. Otherwise, we scrape
// up against VirtualBox's own locking mechanisms.
//
// It would be preferable to simply have a lock around, say, the VBoxManage
// command, but with our current one-server-process-per-machine model it is
// impossible to dictate this locking on the server side.
type SerialDriver struct {
	Driver
	sync.Locker
}

func NewSerialDriver(innerDriver Driver) Driver {
	return newSerialDriverWithLock(innerDriver, stdLock)
}

func newSerialDriverWithLock(innerDriver Driver, lock sync.Locker) Driver {
	return &SerialDriver{
		Driver: innerDriver,
		Locker: lock,
	}
}

// Create a host using the driver's config
func (d *SerialDriver) Create() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Create()
}

// DriverName returns the name of the driver as it is registered
func (d *SerialDriver) DriverName() string {
	d.Lock()
	defer d.Unlock()
	return d.Driver.DriverName()
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *SerialDriver) GetCreateFlags() []mcnflag.Flag {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetCreateFlags()
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *SerialDriver) GetIP() (string, error) {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetIP()
}

// GetMachineName returns the name of the machine
func (d *SerialDriver) GetMachineName() string {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetMachineName()
}

// GetSSHHostname returns hostname for use with ssh
func (d *SerialDriver) GetSSHHostname() (string, error) {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetSSHHostname()
}

// GetSSHKeyPath returns key path for use with ssh
func (d *SerialDriver) GetSSHKeyPath() string {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetSSHKeyPath()
}

// GetSSHPort returns port for use with ssh
func (d *SerialDriver) GetSSHPort() (int, error) {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetSSHPort()
}

// GetSSHUsername returns username for use with ssh
func (d *SerialDriver) GetSSHUsername() string {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetSSHUsername()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *SerialDriver) GetURL() (string, error) {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetURL()
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *SerialDriver) GetState() (state.State, error) {
	d.Lock()
	defer d.Unlock()
	return d.Driver.GetState()
}

// Kill stops a host forcefully
func (d *SerialDriver) Kill() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Kill()
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *SerialDriver) PreCreateCheck() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.PreCreateCheck()
}

// Remove a host
func (d *SerialDriver) Remove() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Remove()
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *SerialDriver) Restart() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Restart()
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *SerialDriver) SetConfigFromFlags(opts DriverOptions) error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.SetConfigFromFlags(opts)
}

// Start a host
func (d *SerialDriver) Start() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Start()
}

// Stop a host gracefully
func (d *SerialDriver) Stop() error {
	d.Lock()
	defer d.Unlock()
	return d.Driver.Stop()
}

func (d *SerialDriver) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Driver)
}
