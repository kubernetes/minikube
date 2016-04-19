package drivers

import (
	"fmt"

	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type DriverNotSupported struct {
	*BaseDriver
	Name string
}

type NotSupported struct {
	DriverName string
}

func (e NotSupported) Error() string {
	return fmt.Sprintf("Driver %q not supported on this platform.", e.DriverName)
}

// NewDriverNotSupported creates a placeholder Driver that replaces
// a driver that is not supported on a given platform. eg fusion on linux.
func NewDriverNotSupported(driverName, hostName, storePath string) Driver {
	return &DriverNotSupported{
		BaseDriver: &BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		Name: driverName,
	}
}

func (d *DriverNotSupported) DriverName() string {
	return d.Name
}

func (d *DriverNotSupported) PreCreateCheck() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetCreateFlags() []mcnflag.Flag {
	return nil
}

func (d *DriverNotSupported) SetConfigFromFlags(flags DriverOptions) error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetURL() (string, error) {
	return "", NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetSSHHostname() (string, error) {
	return "", NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) GetState() (state.State, error) {
	return state.Error, NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Create() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Remove() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Start() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Stop() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Restart() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Kill() error {
	return NotSupported{d.DriverName()}
}

func (d *DriverNotSupported) Upgrade() error {
	return NotSupported{d.DriverName()}
}
