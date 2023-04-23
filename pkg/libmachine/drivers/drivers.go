package drivers

import (
	"errors"

	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/minikube/command"
)

// Driver defines how a host is created and controlled.
// Different types of driver represent different ways hosts
// can be created (e.g. different hypervisors, cloud providers, ..)
// The Driver is responsible of maintaining the machine state between minikube calls..
// The Driver should not rely on initialized objects as struct fields,
// as at each run, the Driver gets initialized at its base value, and its
// fileds populated with the machine's RawDriver field.
type Driver interface {
	// RunCommand runs a command inside the linux machine
	RunCommand(string) (string, error)
	GetRunner() (command.Runner, error)

	// DriverName returns the name of the driver
	DriverName() string

	// GetIP returns an IP or hostname that this machine is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)
	// GetMachineName returns the name of the machine
	GetMachineName() string
	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetSSHHostname returns hostname for use with ssh
	// GetSSHHostname() (string, error)

	// GetSSHKeyPath returns key path for use with ssh
	// GetSSHKeyPath() string

	// GetSSHPort returns port for use with ssh
	// GetSSHPort() (int, error)

	// GetSSHUsername returns username for use with ssh
	// TODO:
	// ho visto che questa la usano anche per prendere il nome
	// dello user a cui associare i permessi dei file nella machine..
	// mi sembra... Insomma viene usata..
	// veidamo di trovare un rimpiazzo.
	// Tipo GetUsername() e basta.. e gli diamo in ritorno lo user 1000
	// GetSSHUsername() string

	// GetState returns the state that the host is in (running, stopped, etc)
	GetState() (state.State, error)
	// Kill stops a host forcefully
	Kill() error
	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error
	// Create a linux machine using the driver's config
	Create() error
	// Remove a host
	Remove() error
	// Restart a host. This may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	Restart() error
	// Start a machine
	Start() error
	// Stop a machine gracefully
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
