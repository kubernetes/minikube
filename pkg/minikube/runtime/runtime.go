// package runtime contains code specific to container runtimes
package runtime

import (
	"fmt"

	"github.com/golang/glog"
)

// CommandRunner is the subset of bootstrapper.CommandRunner this package consumes
type CommandRunner interface {
	Run(string) error
}

// Manager is a common interface for container runtimes
type Manager interface {
	// Name is a human readable name for a runtime
	Name() string
	// Enable idempotently enables this runtime on a host
	Enable(Runner) error
	// Disable idempotently disables this runtime on a host
	Disable(Runner) error
	// Load an image idempotently into the runtime on a host
	LoadImage(Runner, string) error
	// Active returns whether or not a runtime is active on a host
	Active(Runner) bool
	// KubeletOptions returns kubelet options for a runtime.
	KubeletOptions(map[string]string) map[string]string
	// SocketPath returns the path to the socket file for a given runtime
	SocketPath() string
}

// disableOthers disables all other runtimes except for me.
func disableOthers(me Manager, h Runner) error {
	// valid values returned by manager.Name()
	runtimes := []string{"containerd", "crio", "docker", "rkt"}
	for _, name := range runtimes {
		r, err := New(Config{Type: name})
		if err != nil {
			return fmt.Errorf("New(%s): %v", name, err)
		}

		// Don't disable myself.
		if r.Name() == me.Name() {
			continue
		}
		// runtime is already disabled, nothing to do.
		if !r.Active(h) {
			continue
		}
		err = r.Disable(h)
		if err != nil {
			glog.Warningf("disable failed: %v", err)
		}
		// Validate that the runtime really is offline - and that Active & Disable are properly written.
		if r.Active(h) {
			return fmt.Errorf("%s is still active after being disabled! 😱", r.Name())
		}
	}
	return nil
}

// Config is runtime configuration
type Config struct {
	// Type of runtime to create ("docker, "crio", etc)
	Type string
}

// New returns an appropriately configured runtime
func New(c Config) (Manager, error) {
	switch c.Type {
	case "", "docker":
		return &Docker{config: c}, nil
	case "crio", "cri-o":
		return &CRIO{config: c}, nil
	case "rkt":
		return &Rkt{config: c}, nil
	case "containerd":
		return &Containerd{config: c}, nil
	default:
		return nil, fmt.Errorf("unknown runtime type: %q", c.Type)
	}
}
