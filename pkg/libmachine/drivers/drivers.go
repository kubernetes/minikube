package drivers

import (
	"errors"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

// Driver defines how a host is created and controlled. Different types of
// driver represent different ways hosts can be created (e.g. different
// hypervisors, different cloud providers)
type Driver interface {
	// Create a host using the driver's config
	Create() error

	// DriverName returns the name of the driver
	DriverName() string

	// GetCreateFlags returns the mcnflag.Flag slice representing the flags
	// that can be set, their descriptions and defaults.
	GetCreateFlags() []mcnflag.Flag

	// GetIP returns an IP or hostname that this host is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)

	// GetMachineName returns the name of the machine
	GetMachineName() string

	// GetSSHHostname returns hostname for use with ssh
	GetSSHHostname() (string, error)

	// GetSSHKeyPath returns key path for use with ssh
	GetSSHKeyPath() string

	// GetSSHPort returns port for use with ssh
	GetSSHPort() (int, error)

	// GetSSHUsername returns username for use with ssh
	GetSSHUsername() string

	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetState returns the state that the host is in (running, stopped, etc)
	GetState() (state.State, error)

	// Kill stops a host forcefully
	Kill() error

	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error

	// Remove a host
	Remove() error

	// Restart a host. This may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	Restart() error

	// SetConfigFromFlags configures the driver with the object that was returned
	// by RegisterCreateFlags
	SetConfigFromFlags(opts DriverOptions) error

	// Start a host
	Start() error

	// Stop a host gracefully
	Stop() error
}

var ErrHostIsNotRunning = errors.New("Host is not running")

type DriverOptions interface {
	String(key string) string
	StringSlice(key string) []string
	Int(key string) int
	Bool(key string) bool
}

func MachineInState(d Driver, desiredState state.State) func() bool {
	return func() bool {
		currentState, err := d.GetState()
		if err != nil {
			log.Debugf("Error getting machine state: %s", err)
		}
		if currentState == desiredState {
			return true
		}
		return false
	}
}

// MustBeRunning will return an error if the machine is not in a running state.
func MustBeRunning(d Driver) error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s != state.Running {
		return ErrHostIsNotRunning
	}

	return nil
}
