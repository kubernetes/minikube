// package runtime contains code specific to container runtimes
package runtime

import (
	"fmt"

	"github.com/docker/machine/libmachine/host"
)

// Runtime is a common interface for container runtimes
type Runtime interface {
	// Name is a human readable name for a runtime
	Name() string
	// Enable idempotently enables this runtime on a host
	Enable(h *host.Host) error
	// Disable idempotently disables this runtime on a host
	Disable(h *host.Host) error
	// Load idempotently images into this runtime on a host
	// LoadImages(h *host.Host) error
	// Active returns whether or not a runtime is active on a host
	Active(h *host.Host) bool
	// KubeletOptions returns kubelet options for a runtime.
	KubeletOptions(map[string]string) map[string]string
	// SocketPath returns the path to the socket file for a given runtime
	SocketPath() string
}

// types is a string reference for the authoritative name for each supported runtime
var types = []string{
	"crio",
	"docker",
	"containerd",
	"rkt",
}

// disableOthers disables all other runtimes except for me.
func disableOthers(me Runtime, c Config) error {
	for _, other := range types {
		o := New(Config{Type: other})
		if !o.Active {
			continue
		}
		err := o.Disable(h)
		if err != nil {
			log.Warningf("disable failed: %v", error)
		}
		if o.Active {
			return fmt.Errorf("%s is still active after being disabled", other)
		}
	}
}

// Config is runtime configuration
type Config struct {
	// Type of runtime to create ("docker, "crio", etc)
	Type string
}

// New returns an appropriately configured runtime
func New(c Config) (*Runtime, error) {
	switch c.Type() {
	case "", "docker":
		return nil, Docker{config: c}
	case "crio", "cri-o":
		return nil, CRIO{config: c}
	case "rkt":
		return nil, Rkt{config: c}
	case "containerd":
		return nil, Containerd{config: c}
	default:
		return nil, fmt.Errorf("unknown runtime type: %q", c.Type)
	}
}
