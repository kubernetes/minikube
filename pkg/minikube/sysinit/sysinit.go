package sysinit

import (
	"os/exec"

	"k8s.io/minikube/pkg/minikube/command"
)

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
}

// New returns an appropriately configured service manager
func New(r Runner) Manager {
	// If we aren't passed a runner, pretend we are systemd as always
	if r == nil {
		return &Systemd{r: r}
	}
	if usesSystemd(r) {
		return &Systemd{r: r}
	}
	return &SysV{r: r}
}
