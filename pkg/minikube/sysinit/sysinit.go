package sysinit

import (
	"os/exec"
	"runtime"

	"github.com/golang/glog"
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
}

// New returns an appropriately configured service manager
func New(r Runner) Manager {
	_, file, no, _ := runtime.Caller(1)

	if r == nil {
		glog.Warningf("manager from %s:%d: nil runner! (systemd)", file, no)
		return nil
	}

	var systemd bool
	if cachedSystemdCheck != nil {
		systemd = *cachedSystemdCheck
	} else {
		glog.Errorf("uncached systemd check from %s:%d", file, no)
		systemd = usesSystemd(r)
		cachedSystemdCheck = &systemd
	}

	if systemd {
		return &Systemd{r: r}
	}
	return &SysV{r: r}
}
