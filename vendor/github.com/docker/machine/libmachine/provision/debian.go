package provision

import (
	"fmt"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("Debian", &RegisteredProvisioner{
		New: NewDebianProvisioner,
	})
}

func NewDebianProvisioner(d drivers.Driver) Provisioner {
	return &DebianProvisioner{
		NewSystemdProvisioner("debian", d),
	}
}

type DebianProvisioner struct {
	SystemdProvisioner
}

func (provisioner *DebianProvisioner) String() string {
	return "debian"
}

func (provisioner *DebianProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	updateMetadata := true

	switch action {
	case pkgaction.Install, pkgaction.Upgrade:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
		updateMetadata = false
	case pkgaction.Purge:
		packageAction = "purge"
		updateMetadata = false
	}

	switch name {
	case "docker":
		name = "docker-engine"
	}

	if updateMetadata {
		if err := waitForLockAptGetUpdate(provisioner); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageAction, name)

	log.Debugf("package: action=%s name=%s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *DebianProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *DebianProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "aufs", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	// HACK: since debian does not come with sudo by default we install
	log.Debug("installing sudo")
	if _, err := provisioner.SSHCommand("if ! type sudo; then apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y sudo; fi"); err != nil {
		return err
	}

	log.Debug("setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	log.Debug("installing base packages")
	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	log.Debug("installing docker")
	if err := installDockerGeneric(provisioner, engineOptions.InstallURL); err != nil {
		return err
	}

	log.Debug("waiting for docker daemon")
	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debug("configuring auth")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debug("configuring swarm")
	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	// enable in systemd
	log.Debug("enabling docker in systemd")
	err = provisioner.Service("docker", serviceaction.Enable)
	return err
}
