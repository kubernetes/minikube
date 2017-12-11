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
	Register("Arch", &RegisteredProvisioner{
		New: NewArchProvisioner,
	})
}

func NewArchProvisioner(d drivers.Driver) Provisioner {
	return &ArchProvisioner{
		NewSystemdProvisioner("arch", d),
	}
}

type ArchProvisioner struct {
	SystemdProvisioner
}

func (provisioner *ArchProvisioner) String() string {
	return "arch"
}

func (provisioner *ArchProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID || provisioner.OsReleaseInfo.IDLike == provisioner.OsReleaseID
}

func (provisioner *ArchProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	updateMetadata := true

	switch action {
	case pkgaction.Install, pkgaction.Upgrade:
		packageAction = "S"
	case pkgaction.Remove:
		packageAction = "R"
		updateMetadata = false
	}

	switch name {
	case "docker-engine":
		name = "docker"
	case "docker":
		name = "docker"
	}

	pacmanOpts := "-" + packageAction
	if updateMetadata {
		pacmanOpts = pacmanOpts + "y"
	}

	pacmanOpts = pacmanOpts + " --noconfirm --noprogressbar"

	command := fmt.Sprintf("sudo -E pacman %s %s", pacmanOpts, name)

	log.Debugf("package: action=%s name=%s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *ArchProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *ArchProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "overlay", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	// HACK: since Arch does not come with sudo by default we install
	log.Debug("Installing sudo")
	if _, err := provisioner.SSHCommand("if ! type sudo; then pacman -Sy --noconfirm --noprogressbar sudo; fi"); err != nil {
		return err
	}

	log.Debug("Setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	log.Debug("Installing base packages")
	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	log.Debug("Installing docker")
	if err := provisioner.Package("docker", pkgaction.Install); err != nil {
		return err
	}

	log.Debug("Starting systemd docker service")
	if err := provisioner.Service("docker", serviceaction.Start); err != nil {
		return err
	}

	log.Debug("Waiting for docker daemon")
	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debug("Configuring auth")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debug("Configuring swarm")
	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	// enable in systemd
	log.Debug("Enabling docker in systemd")
	err = provisioner.Service("docker", serviceaction.Enable)
	return err
}
