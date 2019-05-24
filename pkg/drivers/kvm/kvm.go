// +build linux

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

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/pkg/errors"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
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
		Boot2DockerURL: constants.DefaultISOURL,
		CPU:            constants.DefaultCPUS,
		DiskSize:       util.CalculateDiskSizeInMB(constants.DefaultDiskSize),
		Memory:         constants.DefaultMemory,
		PrivateNetwork: defaultPrivateNetworkName,
		Network:        defaultNetworkName,
		DiskPath:       filepath.Join(constants.GetMinipath(), "machines", config.GetMachineName(), fmt.Sprintf("%s.rawdisk", config.GetMachineName())),
		ISO:            filepath.Join(constants.GetMinipath(), "machines", config.GetMachineName(), "boot2docker.iso"),
	}
}

// PreCommandCheck checks the connection before issuing a command
func (d *Driver) PreCommandCheck() error {
	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "Error connecting to libvirt socket.  Have you added yourself to the libvirtd group?")
	}
	libVersion, err := conn.GetLibVersion()
	if err != nil {
		return errors.Wrap(err, "getting libvirt version")
	}
	log.Debugf("Using libvirt version %d", libVersion)

	return nil
}

// GetURL returns a Docker compatible host URL for connecting to this host
func (d *Driver) GetURL() (string, error) {
	if err := d.PreCommandCheck(); err != nil {
		return "", errors.Wrap(err, "getting URL, precheck failed")
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", errors.Wrap(err, "getting URL, could not get IP")
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (st state.State, err error) {
	dom, conn, err := d.getDomain()
	if err != nil {
		return state.None, errors.Wrap(err, "getting connection")
	}
	defer func() {
		if ferr := closeDomain(dom, conn); ferr != nil {
			err = ferr
		}
	}()

	lvs, _, err := dom.GetState() // state, reason, error
	if err != nil {
		return state.None, errors.Wrap(err, "getting domain state")
	}
	st = machineState(lvs)
	return // st, err
}

// machineState converts libvirt state to libmachine state
func machineState(lvs libvirt.DomainState) state.State {
	// Possible States:
	//
	// VIR_DOMAIN_NOSTATE no state
	// VIR_DOMAIN_RUNNING the domain is running
	// VIR_DOMAIN_BLOCKED the domain is blocked on resource
	// VIR_DOMAIN_PAUSED the domain is paused by user
	// VIR_DOMAIN_SHUTDOWN the domain is being shut down
	// VIR_DOMAIN_SHUTOFF the domain is shut off
	// VIR_DOMAIN_CRASHED the domain is crashed
	// VIR_DOMAIN_PMSUSPENDED the domain is suspended by guest power management
	// VIR_DOMAIN_LAST this enum value will increase over time as new events are added to the libvirt API. It reflects the last state supported by this version of the libvirt API.

	switch lvs {
	// DOMAIN_SHUTDOWN technically means the VM is still running, but in the
	// process of being shutdown, so we return state.Running
	case libvirt.DOMAIN_RUNNING, libvirt.DOMAIN_SHUTDOWN:
		return state.Running
	case libvirt.DOMAIN_BLOCKED, libvirt.DOMAIN_CRASHED:
		return state.Error
	case libvirt.DOMAIN_PAUSED:
		return state.Paused
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
		return "", errors.Wrap(err, "machine in unknown state")
	}
	if s != state.Running {
		return "", errors.New("host is not running")
	}
	ip, err := d.lookupIP()
	if err != nil {
		return "", errors.Wrap(err, "getting IP")
	}

	return ip, nil
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
func (d *Driver) Kill() (err error) {
	dom, conn, err := d.getDomain()
	if err != nil {
		return errors.Wrap(err, "getting connection")
	}
	defer func() {
		if ferr := closeDomain(dom, conn); ferr != nil {
			err = ferr
		}
	}()
	return dom.Destroy()
}

// Restart a host
func (d *Driver) Restart() error {
	return pkgdrivers.Restart(d)
}

// Start a host
func (d *Driver) Start() (err error) {
	// if somebody/something deleted the network in the meantime,
	// we might need to recreate it. It's (nearly) a noop if the network exists.
	log.Info("Creating network...")
	err = d.createNetwork()
	if err != nil {
		return errors.Wrap(err, "creating network")
	}

	// this call ensures that all networks are active
	log.Info("Ensuring networks are active...")
	err = d.ensureNetwork()
	if err != nil {
		return errors.Wrap(err, "ensuring active networks")
	}

	log.Info("Getting domain xml...")
	dom, conn, err := d.getDomain()
	if err != nil {
		return errors.Wrap(err, "getting connection")
	}
	defer func() {
		if ferr := closeDomain(dom, conn); ferr != nil {
			err = ferr
		}
	}()

	log.Info("Creating domain...")
	if err := dom.Create(); err != nil {
		return errors.Wrap(err, "Error creating VM")
	}

	log.Info("Waiting to get IP...")
	for i := 0; i <= 40; i++ {
		ip, err := d.GetIP()
		if err != nil {
			return errors.Wrap(err, "getting ip during machine start")
		}
		if ip == "" {
			log.Debugf("Waiting for machine to come up %d/%d", i, 40)
			time.Sleep(3 * time.Second)
			continue
		}

		if ip != "" {
			log.Infof("Found IP for machine: %s", ip)
			d.IPAddress = ip
			break
		}
	}

	if d.IPAddress == "" {
		return errors.New("Machine didn't return an IP after 120 seconds")
	}

	log.Info("Waiting for SSH to be available...")
	if err := drivers.WaitForSSH(d); err != nil {
		d.IPAddress = ""
		return errors.Wrap(err, "SSH not available after waiting")
	}

	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() (err error) {
	log.Info("Creating KVM machine...")
	defer log.Infof("KVM machine creation complete!")

	err = d.createNetwork()
	if err != nil {
		return errors.Wrap(err, "creating network")
	}
	if d.GPU {
		log.Info("Creating devices...")
		d.DevicesXML, err = getDevicesXML()
		if err != nil {
			return errors.Wrap(err, "creating devices")
		}
	}

	store := d.ResolveStorePath(".")
	log.Infof("Setting up store path in %s ...", store)
	// 0755 because it must be accessible by libvirt/qemu across a variety of configs
	if err := os.MkdirAll(store, 0755); err != nil {
		return errors.Wrap(err, "creating store")
	}

	log.Infof("Building disk image from %s", d.Boot2DockerURL)
	if err = pkgdrivers.MakeDiskImage(d.BaseDriver, d.Boot2DockerURL, d.DiskSize); err != nil {
		return errors.Wrap(err, "Error creating disk")
	}

	if err := ensureDirPermissions(store); err != nil {
		log.Errorf("unable to ensure permissions on %s: %v", store, err)
	}

	log.Info("Creating domain...")
	dom, err := d.createDomain()
	if err != nil {
		return errors.Wrap(err, "creating domain")
	}
	defer func() {
		if ferr := dom.Free(); ferr != nil {
			err = ferr
		}
	}()
	return d.Start()
}

// ensureDirPermissions ensures that libvirt has access to access the image store directory
func ensureDirPermissions(store string) error {
	// traverse upwards from /home/user/.minikube/machines to ensure
	// that libvirt/qemu has execute access
	for dir := store; dir != "/"; dir = filepath.Dir(dir) {
		log.Debugf("Checking permissions on dir: %s", dir)
		s, err := os.Stat(dir)
		if err != nil {
			return err
		}
		owner := int(s.Sys().(*syscall.Stat_t).Uid)
		if owner != os.Geteuid() {
			log.Debugf("Skipping %s - not owner", dir)
			continue
		}
		mode := s.Mode()
		if mode&0011 != 1 {
			log.Infof("Setting executable bit set on %s (perms=%s)", dir, mode)
			mode |= 0011
			if err := os.Chmod(dir, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() (err error) {
	d.IPAddress = ""
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "getting state of VM")
	}

	if s != state.Stopped {
		dom, conn, err := d.getDomain()
		defer func() {
			if ferr := closeDomain(dom, conn); ferr != nil {
				err = ferr
			}
		}()
		if err != nil {
			return errors.Wrap(err, "getting connection")
		}

		err = dom.Shutdown()
		if err != nil {
			return errors.Wrap(err, "stopping vm")
		}

		for i := 0; i < 60; i++ {
			s, err := d.GetState()
			if err != nil {
				return errors.Wrap(err, "Error getting state of VM")
			}
			if s == state.Stopped {
				return nil
			}
			log.Infof("Waiting for machine to stop %d/%d", i, 60)
			time.Sleep(1 * time.Second)
		}

	}

	return fmt.Errorf("unable to stop vm, current state %q", s.String())
}

// Remove a host
func (d *Driver) Remove() error {
	log.Debug("Removing machine...")
	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "getting connection")
	}
	defer conn.Close()

	// Tear down network if it exists and is not in use by another minikube instance
	log.Debug("Trying to delete the networks (if possible)")
	if err := d.deleteNetwork(); err != nil {
		log.Warnf("Deleting of networks failed: %v", err)
	} else {
		log.Info("Successfully deleted networks")
	}

	// Tear down the domain now
	log.Debug("Checking if the domain needs to be deleted")
	dom, err := conn.LookupDomainByName(d.MachineName)
	if err != nil {
		log.Warnf("Domain %s does not exist, nothing to clean up...", d.MachineName)
		return nil
	}

	log.Infof("Domain %s exists, removing...", d.MachineName)
	if err := d.destroyRunningDomain(dom); err != nil {
		return errors.Wrap(err, "destroying running domain")
	}

	return dom.Undefine()
}

func (d *Driver) destroyRunningDomain(dom *libvirt.Domain) error {
	state, reason, err := dom.GetState()
	if err != nil {
		return errors.Wrap(err, "getting domain state")
	}

	if state == libvirt.DOMAIN_SHUTOFF && reason == int(libvirt.DOMAIN_SHUTOFF_DESTROYED) {
		log.Warnf("Domain %s already destroyed, skipping...", d.MachineName)
		return nil
	}

	return dom.Destroy()
}
