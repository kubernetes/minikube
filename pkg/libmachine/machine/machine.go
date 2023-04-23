package machine

import (
	"regexp"

	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcndockerclient"
	"k8s.io/minikube/pkg/libmachine/mcnerror"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/versioncmp"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/cert"
	"k8s.io/minikube/pkg/libmachine/cruntime"
	"k8s.io/minikube/pkg/libmachine/detector"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/provision"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/runner"
	"k8s.io/minikube/pkg/minikube/command"
)

// NOTE:
// at the moment, each time we want to use a lowe-level feature of the lib,
// we're reading os-release of the linux machine, parsing, and then
// for for each provisioner we have, we initialize it and try to see if fits

var (
	validHostNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*$`)
)

// alias machine=host
// HOst e' la struttura da cui partiamo..
// e qua che sta il riferimento al driver
type Machine struct {
	// Driver is the driver in use,
	// it can be swapped if the circumstances require it.
	// It can always be brought back to the desired one, which
	// is referenced by the DriverName field
	Driver drivers.Driver

	// Provisioner is the provisioner in use
	Provisioner provision.Provisioner

	// Runner is a mechanism used to run commands inside the machine
	Runner runner.Runner

	//
	//
	//
	// The part that directly depends on the configuration.

	// ConfigVersion is used to match an existing machine against the version
	// that the current release of minikube supports
	ConfigVersion int

	// MachineOptions
	MachineOptions *Options

	// Name of the machine,
	// needed in order to control multiple machines
	Name string

	// RawDriver represents the configuration of the machine driver.
	//
	// The idea is that we need the machine to resume its state between
	// minikube command runs, and to do that we rely on the driver.
	// Ideally the driver should be able to get initialized and resume the
	// machine state once configured.
	// We achieve persistence of the machine by saving the RawDriver to a file and
	// reading it back to the driver once another command is issued.
	RawDriver []byte `json:"-"`
}

type Options struct {
	Driver         string
	Runtime        string
	Memory         int
	Disk           int
	RuntimeOptions *cruntime.Options
	AuthOptions    *auth.Options
}

type Metadata struct {
	ConfigVersion  int
	DriverName     string
	MachineOptions Options
}

// TODO:
func (m *Machine) GetRunner() (command.Runner, error) { return nil, nil }

func (m *Machine) Provision() error {
	provisioner, err := detector.DetectProvisioner(m.Driver)
	if err != nil {
		return err
	}

	return provisioner.Provision(*m.MachineOptions.AuthOptions, *m.MachineOptions.RuntimeOptions)
}

func ValidateMachineName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Machine) RunCommand(command string) (string, error) {
	return h.Driver.RunCommand(command)
}

func (h *Machine) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		return mcnerror.ErrHostAlreadyInState{
			Name:  h.Name,
			State: desiredState,
		}
	}

	if err := action(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
}

func (m *Machine) WaitForCRuntime() error { return nil }

func (m *Machine) WaitForDocker() error {
	provisioner, err := detector.DetectProvisioner(m.Driver)
	if err != nil {
		return err
	}

	return provision.WaitForDocker(provisioner, cruntime.DefaultPort)
}

func (m *Machine) Start() error {
	log.Infof("Starting %q...", m.Name)
	if err := m.runActionForState(m.Driver.Start, state.Running); err != nil {
		return err
	}

	log.Infof("Machine %q was started.", m.Name)

	return m.WaitForDocker()
}

func (m *Machine) Stop() error {
	log.Infof("Stopping %q...", m.Name)
	if err := m.runActionForState(m.Driver.Stop, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was stopped.", m.Name)
	return nil
}

func (m *Machine) Kill() error {
	log.Infof("Killing %q...", m.Name)
	if err := m.runActionForState(m.Driver.Kill, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was killed.", m.Name)
	return nil
}

func (m *Machine) Restart() error {
	log.Infof("Restarting %q...", m.Name)
	if drivers.MachineInState(m.Driver, state.Stopped)() {
		if err := m.Start(); err != nil {
			return err
		}
	} else if drivers.MachineInState(m.Driver, state.Running)() {
		if err := m.Driver.Restart(); err != nil {
			return err
		}
		if err := mcnutils.WaitFor(drivers.MachineInState(m.Driver, state.Running)); err != nil {
			return err
		}
	}

	return m.WaitForDocker()
}

func (m *Machine) DockerVersion() (string, error) {
	url, err := m.Driver.GetURL()
	if err != nil {
		return "", err
	}

	dockerMachine := &mcndockerclient.RemoteDocker{
		HostURL:    url,
		AuthOption: m.AuthOptions(),
	}
	dockerVersion, err := mcndockerclient.DockerVersion(dockerMachine)
	if err != nil {
		return "", err
	}

	return dockerVersion, nil
}

func (m *Machine) Upgrade() error {
	machineState, err := m.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		log.Info("Starting machine so machine can be upgraded...")
		if err := m.Start(); err != nil {
			return err
		}
	}

	p, err := detector.DetectProvisioner(m.Driver)
	if err != nil {
		return err
	}

	dockerVersion, err := m.DockerVersion()
	if err != nil {
		return err
	}

	// If we're upgrading from a pre-CE (e.g., 1.13.1) release to a CE
	// release (e.g., 17.03.0-ce), we should simply uninstall and
	// re-install from scratch, since the official package names will
	// change from 'docker-engine' to 'docker-ce'.
	if versioncmp.LessThanOrEqualTo(dockerVersion, detector.LastReleaseBeforeCEVersioning) &&
		// RancherOS and boot2docker, being 'static ISO builds', have
		// an upgrade process which simply grabs the latest if it's
		// different, and so do not need to jump through this hoop to
		// upgrade safely.
		p.GetProvisionerName() != "rancheros" &&
		p.GetProvisionerName() != "boot2docker" {

		// Name of package 'docker-engine' will fall through in this
		// case, so that we execute, e.g.,
		//
		// 'sudo apt-get purge -y docker-engine'
		if err := p.PackageAction("docker-engine", pkgaction.Purge); err != nil {
			return err
		}

		// Then we kick off the normal provisioning process which will
		// go off and install Docker (get.docker.com script should work
		// fine to install Docker from scratch after removing the old
		// packages, and images/containers etc. should be preserved in
		// /var/lib/docker)
		return m.Provision()
	}

	log.Info("Upgrading docker...")
	if err := p.PackageAction("docker", pkgaction.Upgrade); err != nil {
		return err
	}

	log.Info("Restarting docker...")
	return p.ServiceAction("docker", serviceaction.Restart)
}

func (m *Machine) URL() (string, error) {
	return m.Driver.GetURL()
}

func (m *Machine) AuthOptions() *auth.Options {
	if m.MachineOptions == nil {
		return nil
	}
	return m.MachineOptions.AuthOptions
}

func (m *Machine) ConfigureAuth() error {
	provisioner, err := detector.DetectProvisioner(m.Driver)
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	return provisioner.Provision(*m.MachineOptions.AuthOptions, *m.MachineOptions.RuntimeOptions)
}

func (m *Machine) ConfigureAllAuth() error {
	log.Info("Regenerating local certificates")
	if err := cert.BootstrapCertificates(m.AuthOptions()); err != nil {
		return err
	}
	return m.ConfigureAuth()
}
