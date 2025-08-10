//go:build darwin

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

package hyperkit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
	"github.com/johanneswuerbach/nfsexports"
	ps "github.com/mitchellh/go-ps"
	hyperkit "github.com/moby/hyperkit/go"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/common"
	"k8s.io/minikube/pkg/minikube/detect"
)

const (
	isoFilename     = "boot2docker.iso"
	pidFileName     = "hyperkit.pid"
	machineFileName = "hyperkit.json"
	permErr         = "%s needs to run with elevated permissions. " +
		"Please run the following command, then try again: " +
		"sudo chown root:wheel %s && sudo chmod u+s %s"
)

// Driver is the machine driver for Hyperkit
type Driver struct {
	*drivers.BaseDriver
	*common.CommonDriver
	Boot2DockerURL string
	DiskSize       int
	CPU            int
	Memory         int
	Cmdline        string
	NFSShares      []string
	NFSSharesRoot  string
	UUID           string
	VpnKitSock     string
	VSockPorts     []string
	ExtraDisks     int
}

// NewDriver creates a new driver for a host
func NewDriver(_, _ string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: "docker",
		},
		CommonDriver: &common.CommonDriver{},
	}
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *Driver) PreCreateCheck() error {
	return d.verifyRootPermissions()
}

// verifyRootPermissions is called before any step which needs root access
func (d *Driver) verifyRootPermissions() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	euid := syscall.Geteuid()
	log.Debugf("exe=%s uid=%d", exe, euid)
	if euid != 0 {
		return fmt.Errorf(permErr, filepath.Base(exe), exe, exe)
	}
	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	if err := d.verifyRootPermissions(); err != nil {
		return err
	}

	// TODO: handle different disk types.
	if err := common.MakeDiskImage(d.BaseDriver, d.Boot2DockerURL, d.DiskSize); err != nil {
		return errors.Wrap(err, "making disk image")
	}

	isoPath := d.ResolveStorePath(isoFilename)
	if err := d.extractKernel(isoPath); err != nil {
		return errors.Wrap(err, "extracting kernel")
	}

	return d.Start()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "hyperkit"
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.IPAddress, nil
}

// GetURL returns a Docker URL inside this host
// e.g. tcp://1.2.3.4:2376
// more info https://github.com/docker/machine/blob/b170508bf44c3405e079e26d5fdffe35a64c6972/libmachine/provision/utils.go#L159_L175
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

// Return the state of the hyperkit pid
func pidState(pid int) (state.State, error) {
	if pid == 0 {
		return state.Stopped, nil
	}
	p, err := ps.FindProcess(pid)
	if err != nil {
		return state.Error, err
	}
	if p == nil {
		log.Debugf("hyperkit pid %d missing from process table", pid)
		return state.Stopped, nil
	}
	// hyperkit or com.docker.hyper
	if !strings.Contains(p.Executable(), "hyper") {
		log.Debugf("pid %d is stale, and is being used by %s", pid, p.Executable())
		return state.Stopped, nil
	}
	return state.Running, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	if err := d.verifyRootPermissions(); err != nil {
		return state.Error, err
	}

	pid := d.getPid()
	log.Debugf("hyperkit pid from json: %d", pid)
	return pidState(pid)
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	if err := d.verifyRootPermissions(); err != nil {
		return err
	}
	return d.sendSignal(syscall.SIGKILL)
}

// Remove a host
func (d *Driver) Remove() error {
	if err := d.verifyRootPermissions(); err != nil {
		return err
	}

	s, err := d.GetState()
	if err != nil || s == state.Error {
		log.Debugf("Error checking machine status: %v, assuming it has been removed already", err)
	}
	if s == state.Running {
		if err := d.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	return common.Restart(d)
}

func (d *Driver) createHost() (*hyperkit.HyperKit, error) {
	stateDir := filepath.Join(d.StorePath, "machines", d.MachineName)
	h, err := hyperkit.New("", d.VpnKitSock, stateDir)
	if err != nil {
		return nil, errors.Wrap(err, "new-ing Hyperkit")
	}

	// TODO: handle the rest of our settings.
	h.Kernel = d.ResolveStorePath("bzimage")
	h.Initrd = d.ResolveStorePath("initrd")
	h.VMNet = true
	h.ISOImages = []string{d.ResolveStorePath(isoFilename)}
	h.Console = hyperkit.ConsoleFile
	h.CPUs = d.CPU
	h.Memory = d.Memory
	h.UUID = d.UUID

	if vsockPorts, err := d.extractVSockPorts(); err != nil {
		return nil, err
	} else if len(vsockPorts) >= 1 {
		h.VSock = true
		h.VSockPorts = vsockPorts
	}

	h.Disks = []hyperkit.Disk{
		&hyperkit.RawDisk{
			Path: common.GetDiskPath(d.BaseDriver),
			Size: d.DiskSize,
			Trim: true,
		},
	}
	for i := 0; i < d.ExtraDisks; i++ {
		h.Disks = append(h.Disks, &hyperkit.RawDisk{
			Path: common.ExtraDiskPath(d.BaseDriver, i),
			Size: d.DiskSize,
			Trim: true,
		})
	}

	return h, nil
}

// Start a host
func (d *Driver) Start() error {
	if err := d.verifyRootPermissions(); err != nil {
		return err
	}

	if err := d.recoverFromUncleanShutdown(); err != nil {
		return err
	}

	h, err := d.createHost()
	if err != nil {
		return err
	}

	log.Debugf("Using UUID %s", h.UUID)
	mac, err := GetMACAddressFromUUID(h.UUID)
	if err != nil {
		return errors.Wrap(err, "getting MAC address from UUID")
	}

	log.Debugf("Generated MAC %s", mac)

	log.Debugf("Starting with cmdline: %s", d.Cmdline)
	_, err = h.Start(d.Cmdline)
	if err != nil {
		return errors.Wrapf(err, "starting with cmd line: %s", d.Cmdline)
	}

	if err := d.setupIP(mac); err != nil {
		return err
	}

	return d.setupNFSMounts()
}

func (d *Driver) setupIP(mac string) error {
	getIP := func() error {
		st, err := d.GetState()
		if err != nil {
			return errors.Wrap(err, "get state")
		}
		if st == state.Error || st == state.Stopped {
			return fmt.Errorf("hyperkit crashed! command line:\n  hyperkit %s", d.Cmdline)
		}

		d.IPAddress, err = common.GetIPAddressByMACAddress(mac)
		if err != nil {
			return &tempError{err}
		}
		return nil
	}

	var err error

	// Implement a retry loop without calling any minikube code
	multiplier := 1
	if detect.NestedVM() {
		multiplier = 3 // will help with running in Free github action Macos VMs (takes 112+ retries on average)
	}
	for i := 0; i < 60*multiplier; i++ {
		log.Debugf("Attempt %d", i)
		err = getIP()
		if err == nil {
			break
		}
		if _, ok := err.(*tempError); !ok {
			return err
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("IP address never found in dhcp leases file %v", err)
	}
	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func (d *Driver) setupNFSMounts() error {
	var err error

	if len(d.NFSShares) > 0 {
		log.Info("Setting up NFS mounts")
		// takes some time here for ssh / nfsd to work properly
		time.Sleep(time.Second * 30)
		err = d.setupNFSShare()
		if err != nil {
			// TODO(tstromberg): Check that logging an and error and return it is appropriate. Seems weird.
			log.Errorf("NFS setup failed: %v", err)
			return err
		}
	}

	return nil
}

type tempError struct {
	Err error
}

func (t tempError) Error() string {
	return "Temporary error: " + t.Err.Error()
}

// recoverFromUncleanShutdown searches for an existing hyperkit.pid file in
// the machine directory. If it can't find it, a clean shutdown is assumed.
// If it finds the pid file, it checks for a running hyperkit process with that pid
// as the existence of a file might not indicate an unclean shutdown but an actual running
// hyperkit server. This is an error situation - we shouldn't start minikube as there is likely
// an instance running already. If the PID in the pidfile does not belong to a running hyperkit
// process, we can safely delete it, and there is a good chance the machine will recover when restarted.
func (d *Driver) recoverFromUncleanShutdown() error {
	stateDir := filepath.Join(d.StorePath, "machines", d.MachineName)
	pidFile := filepath.Join(stateDir, pidFileName)

	if _, err := os.Stat(pidFile); err != nil {
		if os.IsNotExist(err) {
			log.Debugf("clean start, hyperkit pid file doesn't exist: %s", pidFile)
			return nil
		}
		return errors.Wrap(err, "stat")
	}

	log.Warnf("minikube might have been shutdown in an unclean way, the hyperkit pid file still exists: %s", pidFile)
	bs, err := os.ReadFile(pidFile)
	if err != nil {
		return errors.Wrapf(err, "reading pidfile %s", pidFile)
	}
	content := strings.TrimSpace(string(bs))
	pid, err := strconv.Atoi(content)
	if err != nil {
		return errors.Wrapf(err, "parsing pidfile %s", pidFile)
	}

	st, err := pidState(pid)
	if err != nil {
		return errors.Wrap(err, "pidState")
	}

	log.Debugf("pid %d is in state %q", pid, st)
	if st == state.Running {
		return nil
	}
	log.Debugf("Removing stale pid file %s...", pidFile)
	if err := os.Remove(pidFile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("removing pidFile %s", pidFile))
	}
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	if err := d.verifyRootPermissions(); err != nil {
		return err
	}
	d.cleanupNfsExports()
	err := d.sendSignal(syscall.SIGTERM)
	if err != nil {
		return errors.Wrap(err, "hyperkit sigterm failed")
	}

	// wait 5s for graceful shutdown
	for i := 0; i < 5; i++ {
		log.Debug("waiting for graceful shutdown")
		time.Sleep(time.Second * 1)
		s, err := d.GetState()
		if err != nil {
			return errors.Wrap(err, "hyperkit waiting graceful shutdown failed")
		}
		if s == state.Stopped {
			return nil
		}
	}

	log.Debug("sending sigkill")
	return d.Kill()
}

func (d *Driver) extractKernel(isoPath string) error {
	for _, f := range []struct {
		pathInIso string
		destPath  string
	}{
		{"/boot/bzimage", "bzimage"},
		{"/boot/initrd", "initrd"},
	} {
		fullDestPath := d.ResolveStorePath(f.destPath)
		if err := common.ExtractFile(isoPath, f.pathInIso, fullDestPath); err != nil {
			return err
		}
	}
	return nil
}

// InvalidPortNumberError implements the Error interface.
// It is used when a VSockPorts port number cannot be recognised as an integer.
type InvalidPortNumberError string

// Error returns an Error for InvalidPortNumberError
func (port InvalidPortNumberError) Error() string {
	return fmt.Sprintf("vsock port '%s' is not an integer", string(port))
}

func (d *Driver) extractVSockPorts() ([]int, error) {
	vsockPorts := make([]int, 0, len(d.VSockPorts))

	for _, port := range d.VSockPorts {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, InvalidPortNumberError(port)
		}
		vsockPorts = append(vsockPorts, p)
	}

	return vsockPorts, nil
}

func (d *Driver) setupNFSShare() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	hostIP, err := GetNetAddr()
	if err != nil {
		return err
	}

	mountCommands := "#/bin/bash\\n"
	log.Info(d.IPAddress)

	for _, share := range d.NFSShares {
		if !path.IsAbs(share) {
			share = d.ResolveStorePath(share)
		}
		nfsConfig := fmt.Sprintf("%s %s -alldirs -mapall=%s", share, d.IPAddress, u.Username)

		if _, err := nfsexports.Add("", d.nfsExportIdentifier(share), nfsConfig); err != nil {
			if strings.Contains(err.Error(), "conflicts with existing export") {
				log.Info("Conflicting NFS Share not setup and ignored:", err)
				continue
			}
			return err
		}

		root := d.NFSSharesRoot
		mountCommands += fmt.Sprintf("sudo mkdir -p %s/%s\\n", root, share)
		mountCommands += fmt.Sprintf("sudo mount -t nfs -o noacl,async %s:%s %s/%s\\n", hostIP, share, root, share)
	}

	if err := nfsexports.ReloadDaemon(); err != nil {
		return err
	}

	writeScriptCmd := fmt.Sprintf("echo -e \"%s\" | sh", mountCommands)

	if _, err := drivers.RunSSHCommandFromDriver(d, writeScriptCmd); err != nil {
		return err
	}

	return nil
}

// p is path
func (d *Driver) nfsExportIdentifier(p string) string {
	return fmt.Sprintf("minikube-hyperkit %s-%s", d.MachineName, p)
}

func (d *Driver) sendSignal(s os.Signal) error {
	pid := d.getPid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return proc.Signal(s)
}

func (d *Driver) getPid() int {
	pidPath := d.ResolveStorePath(machineFileName)

	f, err := os.Open(pidPath)
	if err != nil {
		log.Warnf("Error reading pid file: %v", err)
		return 0
	}
	dec := json.NewDecoder(f)

	var config struct {
		Pid int `json:"pid"`
	}

	if err := dec.Decode(&config); err != nil {
		log.Warnf("Error decoding pid file: %v", err)
		return 0
	}

	return config.Pid
}

func (d *Driver) cleanupNfsExports() {
	if len(d.NFSShares) > 0 {
		log.Infof("You must be root to remove NFS shared folders. Please type root password.")
		for _, share := range d.NFSShares {
			if _, err := nfsexports.Remove("", d.nfsExportIdentifier(share)); err != nil {
				log.Errorf("failed removing nfs share (%s): %v", share, err)
			}
		}

		if err := nfsexports.ReloadDaemon(); err != nil {
			log.Errorf("failed to reload the nfs daemon: %v", err)
		}
	}
}
