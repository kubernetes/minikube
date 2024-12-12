//go:build linux

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kvm

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/util/retry"
	"libvirt.org/go/libvirt"
)

// Driver is the machine driver for KVM
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver

	// How much memory, in MB, to allocate to the VM
	Memory int

	// How many cpus to allocate to the VM
	CPU int

	// The name of the default network
	Network string

	// The name of the private network
	PrivateNetwork string

	// The size of the disk to be created for the VM, in MB
	DiskSize int

	// The path of the disk .img
	DiskPath string

	// A file or network URI to fetch the minikube ISO
	Boot2DockerURL string

	// The location of the iso to boot from
	ISO string

	// The randomly generated MAC Address
	// If empty, a random MAC will be generated.
	MAC string

	// The randomly generated MAC Address for the NIC attached to the private network
	// If empty, a random MAC will be generated.
	PrivateMAC string

	// Whether to passthrough GPU devices from the host to the VM.
	GPU bool

	// Whether to hide the KVM hypervisor signature from the guest
	Hidden bool

	// XML that needs to be added to passthrough GPU devices.
	DevicesXML string

	// QEMU Connection URI
	ConnectionURI string

	// NUMA node count default value is 1
	NUMANodeCount int

	// NUMA XML
	NUMANodeXML string

	// Extra Disks
	ExtraDisks int

	// Extra Disks XML
	ExtraDisksXML []string
}

const (
	qemusystem                = "qemu:///system"
	defaultPrivateNetworkName = "minikube-net"
	defaultNetworkName        = "default"
)

// NewDriver creates a new driver for a host
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
			SSHUser:     "docker",
		},
		CommonDriver:   &pkgdrivers.CommonDriver{},
		PrivateNetwork: defaultPrivateNetworkName,
		Network:        defaultNetworkName,
		ConnectionURI:  qemusystem,
	}
}

// GetURL returns a Docker URL inside this host
// e.g. tcp://1.2.3.4:2376
// more info https://github.com/docker/machine/blob/b170508bf44c3405e079e26d5fdffe35a64c6972/libmachine/provision/utils.go#L159_L175
func (d *Driver) GetURL() (string, error) {
	if err := d.PreCommandCheck(); err != nil {
		return "", errors.Wrap(err, "prechecking")
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", errors.Wrap(err, "getting domain IP")
	}

	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

// PreCommandCheck checks the connection before issuing a command
func (d *Driver) PreCommandCheck() error {
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("unable to close libvirt connection: %v", err)
		}
	}()

	libVersion, err := conn.GetLibVersion()
	if err != nil {
		return errors.Wrap(err, "getting libvirt version")
	}

	log.Debugf("using libvirt version %d", libVersion)

	return nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	dom, conn, err := d.getDomain()
	if err != nil {
		return state.None, errors.Wrap(err, "getting domain")
	}
	defer func() {
		if err := closeDomain(dom, conn); err != nil {
			log.Errorf("unable to close domain: %v", err)
		}
	}()

	lvs, _, err := dom.GetState() // state, reason, error
	if err != nil {
		return state.None, errors.Wrap(err, "getting domain state")
	}

	return machineState(lvs), nil
}

// machineState converts libvirt state to libmachine state
func machineState(lvs libvirt.DomainState) state.State {
	// Possible States (ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainState):
	// - VIR_DOMAIN_NOSTATE no state
	// - VIR_DOMAIN_RUNNING the domain is running
	// - VIR_DOMAIN_BLOCKED the domain is blocked on resource
	// - VIR_DOMAIN_PAUSED the domain is paused by user
	// - VIR_DOMAIN_SHUTDOWN the domain is being shut down
	// - VIR_DOMAIN_SHUTOFF the domain is shut off
	// - VIR_DOMAIN_CRASHED the domain is crashed
	// - VIR_DOMAIN_PMSUSPENDED the domain is suspended by guest power management
	// - VIR_DOMAIN_LAST this enum value will increase over time as new events are added to the libvirt API. It reflects the last state supported by this version of the libvirt API.

	switch lvs {
	case libvirt.DOMAIN_RUNNING:
		return state.Running
	case libvirt.DOMAIN_BLOCKED, libvirt.DOMAIN_CRASHED:
		return state.Error
	case libvirt.DOMAIN_PAUSED:
		return state.Paused
	case libvirt.DOMAIN_SHUTDOWN:
		return state.Stopping
	case libvirt.DOMAIN_SHUTOFF:
		return state.Stopped
	case libvirt.DOMAIN_PMSUSPENDED:
		return state.Saved
	case libvirt.DOMAIN_NOSTATE:
		return state.None
	default:
		return state.None
	}
}

// GetIP returns an IP or hostname that this host is available at
func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", errors.Wrap(err, "getting domain state")
	}

	if s != state.Running {
		return "", errors.New("domain is not running")
	}

	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return "", errors.Wrap(err, "getting libvirt connection")
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("unable to close libvirt connection: %v", err)
		}
	}()

	return ipFromXML(conn, d.MachineName, d.PrivateNetwork)
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "kvm2"
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "getting domain state")
	}

	if s == state.Stopped {
		return nil
	}

	log.Info("killing domain...")

	dom, conn, err := d.getDomain()
	if err != nil {
		return errors.Wrap(err, "getting domain")
	}
	defer func() {
		if err := closeDomain(dom, conn); err != nil {
			log.Errorf("unable to close domain: %v", err)
		}
	}()

	// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainDestroy
	//   "virDomainDestroy first requests that a guest terminate (e.g. SIGTERM), then waits for it to comply.
	//    After a reasonable timeout, if the guest still exists, virDomainDestroy will forcefully terminate the guest (e.g. SIGKILL)
	//    if necessary (which may produce undesirable results, for example unflushed disk cache in the guest).
	//    To avoid this possibility, it's recommended to instead call virDomainDestroyFlags, sending the VIR_DOMAIN_DESTROY_GRACEFUL flag."
	// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainDestroyFlags
	// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainDestroyFlagsValues
	// we're using default virDomainDestroyFlags - ie, VIR_DOMAIN_DESTROY_DEFAULT (0), the "Default behavior - could lead to data loss!!"
	return dom.Destroy()
}

// Restart a host
func (d *Driver) Restart() error {
	log.Info("restarting domain...")
	return pkgdrivers.Restart(d)
}

// Start a host
func (d *Driver) Start() error {
	log.Info("starting domain...")

	// this call ensures that all networks are active
	log.Info("ensuring networks are active...")
	if err := d.ensureNetwork(); err != nil {
		return errors.Wrap(err, "ensuring active networks")
	}

	log.Info("getting domain XML...")
	dom, conn, err := d.getDomain()
	if err != nil {
		return errors.Wrap(err, "getting domain XML")
	}
	defer func() {
		if err := closeDomain(dom, conn); err != nil {
			log.Errorf("unable to close domain: %v", err)
		}
	}()

	log.Info("creating domain...")
	if err := dom.Create(); err != nil {
		return errors.Wrap(err, "creating domain")
	}

	log.Info("waiting for IP...")
	if err := d.waitForStaticIP(conn); err != nil {
		return errors.Wrap(err, "waiting for IP")
	}

	log.Info("waiting for SSH...")
	if err := drivers.WaitForSSH(d); err != nil {
		return errors.Wrap(err, "waiting for SSH")
	}

	return nil
}

// waitForStaticIP waits for IP address of domain that has been created & starting and then makes that IP static.
func (d *Driver) waitForStaticIP(conn *libvirt.Connect) error {
	query := func() error {
		sip, err := ipFromAPI(conn, d.MachineName, d.PrivateNetwork)
		if err != nil {
			return errors.Wrap(err, "getting domain IP, will retry")
		}

		if sip == "" {
			return errors.New("waiting for domain to come up")
		}

		log.Infof("found domain IP: %s", sip)
		d.IPAddress = sip

		return nil
	}
	if err := retry.Local(query, 1*time.Minute); err != nil {
		return fmt.Errorf("domain %s didn't return IP after 1 minute", d.MachineName)
	}

	log.Info("reserving static IP address...")
	if err := addStaticIP(conn, d.PrivateNetwork, d.MachineName, d.PrivateMAC, d.IPAddress); err != nil {
		log.Warnf("failed reserving static IP address %s for domain %s, will continue anyway: %v", d.IPAddress, d.MachineName, err)
	} else {
		log.Infof("reserved static IP address %s for domain %s", d.IPAddress, d.MachineName)
	}

	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("creating KVM machine...")

	log.Info("creating network...")
	if err := d.createNetwork(); err != nil {
		return errors.Wrap(err, "creating network")
	}

	if d.GPU {
		log.Info("getting devices XML...")
		xml, err := getDevicesXML()
		if err != nil {
			return errors.Wrap(err, "getting devices XML")
		}
		d.DevicesXML = xml
	}

	if d.NUMANodeCount > 1 {
		numaXML, err := numaXML(d.CPU, d.Memory, d.NUMANodeCount)
		if err != nil {
			return errors.Wrap(err, "creating NUMA XML")
		}
		d.NUMANodeXML = numaXML
	}

	store := d.ResolveStorePath(".")
	log.Infof("setting up store path in %s ...", store)
	// 0755 because it must be accessible by libvirt/qemu across a variety of configs
	if err := os.MkdirAll(store, 0755); err != nil {
		return errors.Wrap(err, "creating store")
	}

	log.Infof("building disk image from %s", d.Boot2DockerURL)
	if err := pkgdrivers.MakeDiskImage(d.BaseDriver, d.Boot2DockerURL, d.DiskSize); err != nil {
		return errors.Wrap(err, "creating disk")
	}

	if d.ExtraDisks > 20 {
		// Limiting the number of disks to 20 arbitrarily. If more disks are
		// needed, the logical name generation has to changed to create them if
		// the form hdaa, hdab, etc
		return errors.New("cannot create more than 20 extra disks")
	}
	for i := 0; i < d.ExtraDisks; i++ {
		diskpath := pkgdrivers.ExtraDiskPath(d.BaseDriver, i)
		if err := pkgdrivers.CreateRawDisk(diskpath, d.DiskSize); err != nil {
			return errors.Wrap(err, "creating extra disks")
		}
		// Starting the logical names for the extra disks from hdd as the cdrom device is set to hdc.
		// TODO: Enhance the domain template to use variable for the logical name of the main disk and the cdrom disk.
		extraDisksXML, err := getExtraDiskXML(diskpath, fmt.Sprintf("hd%v", string(rune('d'+i))))
		if err != nil {
			return errors.Wrap(err, "creating extraDisk XML")
		}
		d.ExtraDisksXML = append(d.ExtraDisksXML, extraDisksXML)
	}

	if err := ensureDirPermissions(store); err != nil {
		log.Errorf("unable to ensure permissions on %s: %v", store, err)
	}

	log.Info("creating domain...")

	dom, err := d.createDomain()
	if err != nil {
		return errors.Wrap(err, "creating domain")
	}
	defer func() {
		if err := dom.Free(); err != nil {
			log.Errorf("unable to free domain: %v", err)
		}
	}()

	if err := d.Start(); err != nil {
		return errors.Wrap(err, "starting domain")
	}

	log.Infof("KVM machine creation complete")
	return nil
}

// ensureDirPermissions ensures that libvirt has access to access the image store directory
func ensureDirPermissions(store string) error {
	// traverse upwards from /home/user/.minikube/machines to ensure
	// that libvirt/qemu has execute access
	for dir := store; dir != "/"; dir = filepath.Dir(dir) {
		log.Debugf("checking permissions on dir: %s", dir)

		s, err := os.Stat(dir)
		if err != nil {
			return err
		}

		owner := int(s.Sys().(*syscall.Stat_t).Uid)
		if owner != os.Geteuid() {
			log.Debugf("skipping %s - not owner", dir)
			continue
		}

		mode := s.Mode()
		if mode&0011 != 1 {
			log.Infof("setting executable bit set on %s (perms=%s)", dir, mode)
			mode |= 0011
			if err := os.Chmod(dir, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop a host gracefully or forcefully otherwise.
func (d *Driver) Stop() error {
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "getting domain state")
	}

	if s == state.Stopped {
		return nil
	}

	log.Info("stopping domain...")

	dom, conn, err := d.getDomain()
	if err != nil {
		return errors.Wrap(err, "getting domain")
	}
	defer func() {
		if err := closeDomain(dom, conn); err != nil {
			log.Errorf("unable to close domain: %v", err)
		}
	}()

	log.Info("gracefully shutting down domain...")

	// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainShutdownFlags
	// note: "The order in which the hypervisor tries each shutdown method is undefined, and a hypervisor is not required to support all methods."
	// so we skip "VIR_DOMAIN_SHUTDOWN_DEFAULT" - the "hypervisor choice" and use bitwise-OR of available virDomainShutdownFlagValues
	// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainShutdownFlagValues
	// ref: https://bugzilla.redhat.com/show_bug.cgi?id=1744156
	//   "libvirt's QEMU driver, used to manage KVM guests, only supports the 'agent' and 'acpi' reboot/shutdown modes because those are the ones QEMU itself supports"
	// note: we don't install/use agent
	if err := dom.ShutdownFlags(libvirt.DOMAIN_SHUTDOWN_ACPI_POWER_BTN); err != nil {
		log.Warnf("setting virDomainShutdownFlags failed, will continue anyway with defaults: %v", err)
		if err := dom.ShutdownFlags(libvirt.DOMAIN_SHUTDOWN_DEFAULT); err != nil {
			log.Errorf("resetting virDomainShutdownFlags to defaults failed, will continue anyway: %v", err)
		}
	}

	if err := dom.Shutdown(); err != nil {
		return errors.Wrap(err, "gracefully shutting down domain")
	}

	if s, err = d.waitForStopState(90, "graceful shutdown"); err == nil {
		log.Info("domain gracefully shutdown")
		return nil
	}

	// could not get domain state
	if s == state.None {
		return err
	}

	// at this point shutdown failed, so we try with a little bit of force
	log.Warn("waiting for domain graceful shutdown failed, will try to force-stop")
	if err := d.Kill(); err != nil {
		log.Warnf("force-stopping domain request failed: %v", err)
	}

	if s, err := d.waitForStopState(30, "force-stop"); err != nil {
		return fmt.Errorf("unable to stop domain %s, current state is %q", d.MachineName, s.String())
	}

	log.Info("domain force-stopped")

	return nil
}

// waitForStopState waits maxsec for the domain to reach a stopped state.
func (d *Driver) waitForStopState(maxsec int, method string) (state.State, error) {
	var s state.State
	var err error
	for i := 0; i < maxsec; i++ {
		if s, err = d.GetState(); err != nil {
			return s, errors.Wrap(err, "getting domain state")
		}

		if s == state.Stopped {
			return state.Stopped, nil
		}

		log.Infof("waiting for domain %s %d/%d", method, i, maxsec)
		time.Sleep(1 * time.Second)
	}

	return s, fmt.Errorf("timed out waiting for domain %s, current state is %q", method, s)
}

// Remove a host
func (d *Driver) Remove() error {
	log.Info("removing KVM machine...")

	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("unable to close libvirt connection: %v", err)
		}
	}()

	// Tear down network if it exists and is not in use by another minikube instance
	log.Info("deleting networks...")
	if err := d.deleteNetwork(); err != nil {
		log.Errorf("deleting networks failed, will continue anyway: %v", err)
	} else {
		log.Info("successfully deleted networks")
	}

	// Tear down the domain now
	log.Info("checking if the domain needs to be deleted")
	dom, err := conn.LookupDomainByName(d.MachineName)
	if err != nil {
		log.Warnf("domain %s does not exist, nothing to clean up...", d.MachineName)
		return nil
	}

	log.Infof("domain %s exists, removing...", d.MachineName)
	if err := d.destroyRunningDomain(dom); err != nil {
		return errors.Wrap(err, "destroying running domain")
	}

	if err := d.undefineDomain(conn, dom); err != nil {
		return errors.Wrap(err, "undefining domain")
	}

	log.Info("removing static IP address...")
	if err := delStaticIP(conn, d.PrivateNetwork, "", "", d.IPAddress); err != nil {
		log.Warnf("failed removing static IP address %s for domain %s, will continue anyway: %v", d.IPAddress, d.MachineName, err)
	} else {
		log.Infof("removed static IP address %s for domain %s", d.IPAddress, d.MachineName)
	}

	log.Infof("KVM machine removal complete")

	return nil
}

func (d *Driver) destroyRunningDomain(dom *libvirt.Domain) error {
	lvs, _, err := dom.GetState()
	if err != nil {
		return errors.Wrap(err, "getting domain state")
	}

	// if the domain is not running, we don't destroy it
	if machineState(lvs) != state.Running {
		log.Warnf("domain %s already destroyed, skipping...", d.MachineName)
		return nil
	}

	return dom.Destroy()
}

func (d *Driver) undefineDomain(conn *libvirt.Connect, dom *libvirt.Domain) error {
	definedDomains, err := conn.ListDefinedDomains()
	if err != nil {
		return errors.Wrap(err, "listing domains")
	}

	var found bool
	for _, domain := range definedDomains {
		if domain == d.MachineName {
			found = true
			break
		}
	}

	if !found {
		log.Warnf("domain %s not defined, skipping undefine...", d.MachineName)
		return nil
	}

	return dom.UndefineFlags(libvirt.DOMAIN_UNDEFINE_KEEP_NVRAM)
}

// lvErr will return libvirt Error struct containing specific libvirt error code, domain, message and level
func lvErr(err error) libvirt.Error {
	if err != nil {
		if lverr, ok := err.(libvirt.Error); ok {
			return lverr
		}
		return libvirt.Error{Code: libvirt.ERR_INTERNAL_ERROR, Message: "internal error"}
	}
	return libvirt.Error{Code: libvirt.ERR_OK, Message: ""}
}
