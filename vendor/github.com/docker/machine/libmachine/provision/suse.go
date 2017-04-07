package provision

import (
	"fmt"
	"strings"

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
	Register("openSUSE", &RegisteredProvisioner{
		New: NewOpenSUSEProvisioner,
	})
	Register("SUSE Linux Enterprise Desktop", &RegisteredProvisioner{
		New: NewSLEDProvisioner,
	})
	Register("SUSE Linux Enterprise Server", &RegisteredProvisioner{
		New: NewSLESProvisioner,
	})
}

func NewSLEDProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		NewSystemdProvisioner("sled", d),
	}
}

func NewSLESProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		NewSystemdProvisioner("sles", d),
	}
}

func NewOpenSUSEProvisioner(d drivers.Driver) Provisioner {
	return &SUSEProvisioner{
		NewSystemdProvisioner("openSUSE", d),
	}
}

type SUSEProvisioner struct {
	SystemdProvisioner
}

func (provisioner *SUSEProvisioner) CompatibleWithHost() bool {
	return strings.ToLower(provisioner.OsReleaseInfo.ID) == strings.ToLower(provisioner.OsReleaseID)
}

func (provisioner *SUSEProvisioner) String() string {
	return "openSUSE"
}

func (provisioner *SUSEProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install:
		packageAction = "in"
	case pkgaction.Remove:
		packageAction = "rm"
	case pkgaction.Upgrade:
		packageAction = "up"
	}

	command := fmt.Sprintf("sudo -E zypper -n %s %s", packageAction, name)

	log.Debugf("zypper: action=%s name=%s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *SUSEProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *SUSEProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	// figure out the filesystem used by /var/lib
	fs, err := provisioner.SSHCommand("stat -f -c %T /var/lib/")
	if err != nil {
		return err
	}
	graphDriver := "overlay"
	if strings.Contains(fs, "btrfs") {
		graphDriver = "btrfs"
	}

	storageDriver, err := decideStorageDriver(provisioner, graphDriver, engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	log.Debug("Setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if strings.ToLower(provisioner.OsReleaseInfo.ID) != "opensuse" {
		// This is a SLE machine, enable the containers module to have access
		// to the docker packages
		if _, err := provisioner.SSHCommand("sudo -E SUSEConnect -p sle-module-containers/12/$(uname -m) -r ''"); err != nil {
			return fmt.Errorf(
				"Error while adding the 'containers' module, make sure this machine is registered either against SUSE Customer Center (SCC) or to a local Subscription Management Tool (SMT): %v",
				err)
		}
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

	// create symlinks for containerd, containerd-shim and runc.
	// We have to do that because machine overrides the openSUSE systemd
	// unit of docker
	if _, err := provisioner.SSHCommand("sudo -E ln -s /usr/sbin/runc /usr/sbin/docker-runc"); err != nil {
		return err
	}
	if _, err := provisioner.SSHCommand("sudo -E ln -s /usr/sbin/containerd /usr/sbin/docker-containerd"); err != nil {
		return err
	}
	if _, err := provisioner.SSHCommand("sudo -E ln -s /usr/sbin/containerd-shim /usr/sbin/docker-containerd-shim"); err != nil {
		return err
	}

	// Is yast2 firewall installed?
	if _, installed := provisioner.SSHCommand("rpm -q yast2-firewall"); installed == nil {
		// Open the firewall port required by docker
		if _, err := provisioner.SSHCommand("sudo -E /sbin/yast2 firewall services add ipprotocol=tcp tcpport=2376 zone=EXT"); err != nil {
			return err
		}
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
	if err := provisioner.Service("docker", serviceaction.Enable); err != nil {
		return err
	}

	return nil
}
