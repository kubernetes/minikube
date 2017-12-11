package provision

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	versionsURL  = "http://releases.rancher.com/os/versions.yml"
	isoURL       = "https://github.com/rancherio/os/releases/download/%s/machine-rancheros.iso"
	hostnameTmpl = `sudo mkdir -p /var/lib/rancher/conf/cloud-config.d/  
sudo tee /var/lib/rancher/conf/cloud-config.d/machine-hostname.yml << EOF
#cloud-config

hostname: %s
EOF
`
)

func init() {
	Register("RancherOS", &RegisteredProvisioner{
		New: NewRancherProvisioner,
	})
}

func NewRancherProvisioner(d drivers.Driver) Provisioner {
	return &RancherProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/var/lib/rancher/conf",
			DaemonOptionsFile: "/var/lib/rancher/conf/docker",
			OsReleaseID:       "rancheros",
			Driver:            d,
		},
	}
}

type RancherProvisioner struct {
	GenericProvisioner
}

func (provisioner *RancherProvisioner) String() string {
	return "rancheros"
}

func (provisioner *RancherProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	command := fmt.Sprintf("sudo system-docker %s %s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *RancherProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	if name == "docker" && action == pkgaction.Upgrade {
		return provisioner.upgrade()
	}

	switch action {
	case pkgaction.Install:
		packageAction = "enabled"
	case pkgaction.Remove:
		packageAction = "disable"
	case pkgaction.Upgrade:
		// TODO: support upgrade
		packageAction = "upgrade"
	}

	command := fmt.Sprintf("sudo rancherctl service %s %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *RancherProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	log.Debugf("Running RancherOS provisioner on %s", provisioner.Driver.GetMachineName())

	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "overlay"
	} else if provisioner.EngineOptions.StorageDriver != "overlay" {
		return fmt.Errorf("Unsupported storage driver: %s", provisioner.EngineOptions.StorageDriver)
	}

	log.Debugf("Setting hostname %s", provisioner.Driver.GetMachineName())
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		log.Debugf("Installing package %s", pkg)
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	if engineOptions.InstallURL == drivers.DefaultEngineInstallURL {
		log.Debugf("Skipping docker engine default: %s", engineOptions.InstallURL)
	} else {
		log.Debugf("Selecting docker engine: %s", engineOptions.InstallURL)
		if err := selectDocker(provisioner, engineOptions.InstallURL); err != nil {
			return err
		}
	}

	log.Debugf("Preparing certificates")
	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debugf("Setting up certificates")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debugf("Configuring swarm")
	err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions)
	return err
}

func (provisioner *RancherProvisioner) SetHostname(hostname string) error {
	// /etc/hosts is bind mounted from Docker, this is hack to that the generic provisioner doesn't try to mv /etc/hosts
	if _, err := provisioner.SSHCommand("sed /127.0.1.1/d /etc/hosts > /tmp/hosts && cat /tmp/hosts | sudo tee /etc/hosts"); err != nil {
		return err
	}

	if err := provisioner.GenericProvisioner.SetHostname(hostname); err != nil {
		return err
	}

	if _, err := provisioner.SSHCommand(fmt.Sprintf(hostnameTmpl, hostname)); err != nil {
		return err
	}

	return nil
}

func (provisioner *RancherProvisioner) upgrade() error {
	switch provisioner.Driver.DriverName() {
	case "virtualbox":
		return provisioner.upgradeIso()
	default:
		log.Infof("Running upgrade")
		if _, err := provisioner.SSHCommand("sudo rancherctl os upgrade -f --no-reboot"); err != nil {
			return err
		}

		log.Infof("Upgrade succeeded, rebooting")
		// ignore errors here because the SSH connection will close
		provisioner.SSHCommand("sudo reboot")

		return nil
	}
}

func (provisioner *RancherProvisioner) upgradeIso() error {
	// Largely copied from Boot2Docker provisioner, we should find a way to share this code
	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err
	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %s...", machineName)

	// TODO: Ideally, we should not read from mcndirs directory at all.
	// The driver should be able to communicate how and where to place the
	// relevant files.
	b2dutils := mcnutils.NewB2dUtils(mcndirs.GetBaseDir())

	url, err := provisioner.getLatestISOURL()
	if err != nil {
		return err
	}

	if err := b2dutils.DownloadISOFromURL(url); err != nil {
		return err
	}

	// Copy the latest version of boot2docker ISO to the machine's directory
	if err := b2dutils.CopyIsoToMachineDir("", machineName); err != nil {
		return err
	}

	log.Infof("Starting machine back up...")

	if err := provisioner.Driver.Start(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Running))
}

func (provisioner *RancherProvisioner) getLatestISOURL() (string, error) {
	log.Debugf("Reading %s", versionsURL)
	resp, err := http.Get(versionsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Don't want to pull in yaml parser, we'll do this manually
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "current: ") {
			log.Debugf("Found %s", line)
			return fmt.Sprintf(isoURL, strings.Split(line, ":")[2]), err
		}
	}

	return "", fmt.Errorf("Failed to find current version")
}

func selectDocker(p Provisioner, baseURL string) error {
	// TODO: detect if its a cloud-init, or a ros setting - and use that..
	if output, err := p.SSHCommand(fmt.Sprintf("wget -O- %s | sh -", baseURL)); err != nil {
		return fmt.Errorf("error selecting docker: (%s) %s", err, output)
	}

	return nil
}
