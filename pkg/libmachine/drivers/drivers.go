package drivers

import (
	"errors"

	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/state"
)

// Driver defines how a host is created and controlled.
// Different types of driver represent different ways hosts
// can be created (e.g. different hypervisors, cloud providers, ..).
// As the Driver is responsible of maintaining the machine state between minikube commands,
// it should not rely on initialized objects as struct fields,
// as at each run, the Driver gets initialized at its base value, and its
// fileds populated with the machine's RawDriver field.
type Driver interface {
	// GetRunner is responsible for instantiating the ssh session if that is the case,
	// or more generically provide a way to run commands inside the machine
	GetRunner() runner.Runner
	// DriverName returns the name of the driver
	DriverName() string

	// NOTE:
	// those are used in order to configure the daemon authentication
	// ...
	// GetIP returns an IP or hostname that this machine is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)
	// GetMachineName returns the name of the machine
	GetMachineName() string
	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetState returns the state that the host is in (running, stopped, etc)
	GetMachineState() (state.State, error)
	// Kill stops a host forcefully
	KillMachine() error
	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error
	// Create a linux machine using the driver's config
	CreateMachine() error
	// Remove a machine
	RemoveMachine() error
	// Restart a host. This may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	RestartMachine() error
	// Start a machine
	StartMachine() error
	// Stop a machine gracefully
	StopMachine() error
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
		currentState, err := d.GetMachineState()
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
	s, err := d.GetMachineState()
	if err != nil {
		return err
	}

	if s != state.Running {
		return ErrHostIsNotRunning
	}

	return nil
}
