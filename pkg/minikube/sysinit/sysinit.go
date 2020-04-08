package sysinit

import (
	"os/exec"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
)

var cachedSystemdCheck *bool

// Runner is the subset of command.Runner this package consumes
type Runner interface {
	RunCmd(cmd *exec.Cmd) (*command.RunResult, error)
}

// Manager is a common interface for init systems
type Manager interface {
	// Name returns the name of the init manager
	Name() string

	// Active returns if a service is active
	Active(string) bool

	// Start starts a service idempotently
	Start(string) error

	// Restart restarts a service
	Restart(string) error

	// Stop stops a service
	Stop(string) error

	// ForceStop stops a service with prejudice
	ForceStop(string) error

	// GenerateInitShim generates any additional init files required for this service
	GenerateInitShim(svc string, binary string, unit string) ([]assets.CopyableFile, error)
}

// New returns an appropriately configured service manager
func New(r Runner) Manager {
	// If we are not provided a runner, we can't do anything anyways
	if r == nil {
		return nil
	}

	var systemd bool

	// Caching the result is important, as this manager may be created in many places,
	// and ssh calls are expensive on some drivers, such as Docker.
	if cachedSystemdCheck != nil {
		systemd = *cachedSystemdCheck
	} else {
		systemd = usesSystemd(r)
		cachedSystemdCheck = &systemd
	}

	if systemd {
		return &Systemd{r: r}
	}
	return &OpenRC{r: r}
}
