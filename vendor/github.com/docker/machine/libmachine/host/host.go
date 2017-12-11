package host

import (
	"regexp"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/versioncmp"
)

var (
	validHostNamePattern                  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*$`)
	stdSSHClientCreator  SSHClientCreator = &StandardSSHClientCreator{}
)

type SSHClientCreator interface {
	CreateSSHClient(d drivers.Driver) (ssh.Client, error)
}

type StandardSSHClientCreator struct {
	drivers.Driver
}

func SetSSHClientCreator(creator SSHClientCreator) {
	stdSSHClientCreator = creator
}

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *Options
	Name          string
	RawDriver     []byte `json:"-"`
}

type Options struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.Options
	SwarmOptions  *swarm.Options
	AuthOptions   *auth.Options
}

type Metadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   Options
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	return stdSSHClientCreator.CreateSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateSSHClient(d drivers.Driver) (ssh.Client, error) {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
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

func (h *Host) WaitForDocker() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provision.WaitForDocker(provisioner, engine.DefaultPort)
}

func (h *Host) Start() error {
	log.Infof("Starting %q...", h.Name)
	if err := h.runActionForState(h.Driver.Start, state.Running); err != nil {
		return err
	}

	log.Infof("Machine %q was started.", h.Name)

	return h.WaitForDocker()
}

func (h *Host) Stop() error {
	log.Infof("Stopping %q...", h.Name)
	if err := h.runActionForState(h.Driver.Stop, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was stopped.", h.Name)
	return nil
}

func (h *Host) Kill() error {
	log.Infof("Killing %q...", h.Name)
	if err := h.runActionForState(h.Driver.Kill, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was killed.", h.Name)
	return nil
}

func (h *Host) Restart() error {
	log.Infof("Restarting %q...", h.Name)
	if drivers.MachineInState(h.Driver, state.Stopped)() {
		if err := h.Start(); err != nil {
			return err
		}
	} else if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Driver.Restart(); err != nil {
			return err
		}
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return err
		}
	}

	return h.WaitForDocker()
}

func (h *Host) DockerVersion() (string, error) {
	url, err := h.Driver.GetURL()
	if err != nil {
		return "", err
	}

	dockerHost := &mcndockerclient.RemoteDocker{
		HostURL:    url,
		AuthOption: h.AuthOptions(),
	}
	dockerVersion, err := mcndockerclient.DockerVersion(dockerHost)
	if err != nil {
		return "", err
	}

	return dockerVersion, nil
}

func (h *Host) Upgrade() error {
	machineState, err := h.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		log.Info("Starting machine so machine can be upgraded...")
		if err := h.Start(); err != nil {
			return err
		}
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	dockerVersion, err := h.DockerVersion()
	if err != nil {
		return err
	}

	// If we're upgrading from a pre-CE (e.g., 1.13.1) release to a CE
	// release (e.g., 17.03.0-ce), we should simply uninstall and
	// re-install from scratch, since the official package names will
	// change from 'docker-engine' to 'docker-ce'.
	if versioncmp.LessThanOrEqualTo(dockerVersion, provision.LastReleaseBeforeCEVersioning) &&
		// RancherOS and boot2docker, being 'static ISO builds', have
		// an upgrade process which simply grabs the latest if it's
		// different, and so do not need to jump through this hoop to
		// upgrade safely.
		provisioner.String() != "rancheros" &&
		provisioner.String() != "boot2docker" {

		// Name of package 'docker-engine' will fall through in this
		// case, so that we execute, e.g.,
		//
		// 'sudo apt-get purge -y docker-engine'
		if err := provisioner.Package("docker-engine", pkgaction.Purge); err != nil {
			return err
		}

		// Then we kick off the normal provisioning process which will
		// go off and install Docker (get.docker.com script should work
		// fine to install Docker from scratch after removing the old
		// packages, and images/containers etc. should be preserved in
		// /var/lib/docker)
		return h.Provision()
	}

	log.Info("Upgrading docker...")
	if err := provisioner.Package("docker", pkgaction.Upgrade); err != nil {
		return err
	}

	log.Info("Restarting docker...")
	return provisioner.Service("docker", serviceaction.Restart)
}

func (h *Host) URL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) AuthOptions() *auth.Options {
	if h.HostOptions == nil {
		return nil
	}
	return h.HostOptions.AuthOptions
}

func (h *Host) ConfigureAuth() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	return provisioner.Provision(swarm.Options{}, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

func (h *Host) Provision() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}
