//go:build darwin

/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package krunkit

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/common"
	"k8s.io/minikube/pkg/drivers/common/virtiofs"
	"k8s.io/minikube/pkg/drivers/common/vmnet"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/firewall"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/process"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	driverName     = "krunkit"
	isoFileName    = "boot2docker.iso"
	pidFileName    = "krunkit.pid"
	sockFileName   = "krunkit.sock"
	logFileName    = "krunkit.log"
	serialFileName = "serial.log"
	logLevelInfo   = "3"
	defaultSSHUser = "docker"
)

// Driver is the machine driver for krunkit.
type Driver struct {
	*drivers.BaseDriver
	*common.CommonDriver
	Boot2DockerURL string
	DiskSize       int
	CPU            int
	Memory         int
	ExtraDisks     int
	VirtiofsMounts []*virtiofs.Mount
	MACAddress     string
	VmnetHelper    vmnet.Helper
}

// Ensure that Driver implements drivers.Driver interface
var _ drivers.Driver = &Driver{}

// NewDriver returns a new krunkit.Driver.
func NewDriver(hostName, storePath string, options *run.CommandOptions) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
		CommonDriver: &common.CommonDriver{
			CommandOptions: *options,
		},
	}
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) GetSSHKeyPath() string {
	return d.ResolveStorePath("id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

func (d *Driver) GetURL() (string, error) {
	if _, err := os.Stat(d.pidfilePath()); err != nil {
		return "", nil
	}
	ip, err := d.GetIP()
	if err != nil {
		log.Warnf("Failed to get IP: %v", err)
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	return d.IPAddress, nil
}

// GetState returns driver state. Since krunkit driver uses 2 processes
// (vmnet-helper, krunkit), this returns combined state of both processes.
func (d *Driver) GetState() (state.State, error) {
	if krunkitState, err := d.getKrunkitState(); err != nil {
		return state.Error, err
	} else if krunkitState == state.Running {
		return state.Running, nil
	}
	return d.VmnetHelper.GetState()
}

func (d *Driver) Create() error {
	var err error
	if d.SSHPort, err = d.GetSSHPort(); err != nil {
		return err
	}

	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Info("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Info("Creating disk image...")
	if err := d.generateDiskImage(d.DiskSize); err != nil {
		return err
	}

	if d.ExtraDisks > 0 {
		log.Info("Creating extra disk images...")
		for i := 0; i < d.ExtraDisks; i++ {
			path := common.ExtraDiskPath(d.BaseDriver, i)
			if err := common.CreateRawDisk(path, d.DiskSize); err != nil {
				return err
			}
		}
	}

	log.Info("Starting krunkit VM...")
	return d.Start()
}

func (d *Driver) Start() error {
	socketPath := d.VmnetHelper.SocketPath()
	if err := d.VmnetHelper.Start(socketPath); err != nil {
		return err
	}

	d.MACAddress = d.VmnetHelper.GetMACAddress()

	if err := d.startKrunkit(socketPath); err != nil {
		return err
	}

	if err := d.setupIP(d.MACAddress); err != nil {
		return err
	}

	if err := common.WaitForSSHAccess(d); err != nil {
		return err
	}

	if len(d.VirtiofsMounts) > 0 {
		log.Infof("Setup virtiofs mounts ...")
		if err := virtiofs.SetupMounts(d, d.VirtiofsMounts); err != nil {
			return err
		}
	}

	return nil
}

// startKrunkit starts the krunkit child process.
func (d *Driver) startKrunkit(socketPath string) error {
	var args = []string{
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--cpus", fmt.Sprintf("%d", d.CPU),
		"--restful-uri", d.restfulURI(),
		"--device", fmt.Sprintf("virtio-net,type=unixgram,path=%s,mac=%s,offloading=%t",
			socketPath, d.MACAddress, d.VmnetHelper.Offloading),
		"--device", fmt.Sprintf("virtio-serial,logFilePath=%s", d.serialPath()),
		"--krun-log-level", logLevelInfo,

		// The first device is the boot disk.
		"--device", fmt.Sprintf("virtio-blk,path=%s", d.isoPath()),
		"--device", fmt.Sprintf("virtio-blk,path=%s", d.diskPath()),
	}

	for i := 0; i < d.ExtraDisks; i++ {
		args = append(args,
			"--device", fmt.Sprintf("virtio-blk,path=%s", common.ExtraDiskPath(d.BaseDriver, i)))
	}

	for _, mount := range d.VirtiofsMounts {
		args = append(args,
			"--device", fmt.Sprintf("virtio-fs,sharedDir=%s,mountTag=%s", mount.HostPath, mount.Tag))
	}

	log.Debugf("executing: krunkit %s", strings.Join(args, " "))
	cmd := exec.Command(driverName, args...)

	// Create krunkit in a new process group, so minikube caller can use killpg
	// to terminate the entire process group without harming the krunkit process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logfile, err := d.openLogfile()
	if err != nil {
		return fmt.Errorf("failed to open krunkit logfile: %w", err)
	}
	defer logfile.Close()
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		return err
	}
	return process.WritePidfile(d.pidfilePath(), cmd.Process.Pid)
}

func (d *Driver) openLogfile() (*os.File, error) {
	logfile := d.ResolveStorePath(logFileName)
	return os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}

// TODO: duplicate from vfkit and hyperkit:
// https://github.com/kubernetes/minikube/issues/21093
func (d *Driver) setupIP(mac string) error {
	var err error
	getIP := func() error {
		d.IPAddress, err = common.GetIPAddressByMACAddress(mac)
		if err != nil {
			return errors.Wrap(err, "failed to get IP address")
		}
		return nil
	}
	// Implement a retry loop because IP address isn't added to dhcp leases file immediately
	for i := 0; i < 60; i++ {
		log.Debugf("Attempt %d", i)
		err = getIP()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if err == nil {
		log.Debugf("IP: %s", d.IPAddress)
		return nil
	}
	if !isBootpdError(err) {
		return errors.Wrap(err, "IP address never found in dhcp leases file")
	}
	if unblockErr := firewall.UnblockBootpd(&d.CommandOptions); unblockErr != nil {
		klog.Errorf("failed unblocking bootpd from firewall: %v", unblockErr)
		exit.Error(reason.IfBootpdFirewall, "ip not found", err)
	}
	out.Styled(style.Restarting, "Successfully unblocked bootpd process from firewall, retrying")
	return fmt.Errorf("ip not found: %v", err)
}

func isBootpdError(err error) bool {
	return strings.Contains(err.Error(), "could not find an IP address")
}

func (d *Driver) Stop() error {
	if err := d.stopKrunkit(); err != nil {
		return err
	}
	return d.VmnetHelper.Stop()
}

func (d *Driver) Kill() error {
	if err := d.killKrunkit(); err != nil {
		return err
	}
	return d.VmnetHelper.Kill()
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return errors.Wrap(err, "get state")
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return errors.Wrap(err, "kill")
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Running {
		if err := d.Stop(); err != nil {
			return err
		}
	}
	return d.Start()
}

func (d *Driver) StartDocker() error {
	return errors.New("hosts without a driver cannot start docker")
}

func (d *Driver) StopDocker() error {
	return errors.New("hosts without a driver cannot stop docker")
}

func (d *Driver) GetDockerConfigDir() string {
	return ""
}

func (d *Driver) getKrunkitState() (state.State, error) {
	pidfile := d.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return state.Error, err
		}
		return state.Stopped, nil
	}
	exists, err := process.Exists(pid, driverName)
	if err != nil {
		return state.Error, err
	}
	if !exists {
		// No process, stale pidfile.
		if err := os.Remove(pidfile); err != nil {
			log.Debugf("failed to remove %q: %s", pidfile, err)
		}
		return state.Stopped, nil
	}
	return state.Running, nil
}

func (d *Driver) stopKrunkit() error {
	// TODO: this stop request may be ignored by the guest:
	// https://github.com/kubernetes/minikube/issues/21092
	if err := d.setKrunkitState("Stop"); err != nil {
		// krunkit may be already stopped, shutting down, or not listening. It
		// does not support HardStop so the only way to recover is to terminate
		// the process.
		log.Debugf("Failed to set krunkit state to 'Stop': %s", err)
		pidfile := d.pidfilePath()
		pid, err := process.ReadPidfile(pidfile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			// No pidfile.
			return nil
		}
		log.Debugf("Terminate krunkit (pid=%d)", pid)
		if err := process.Terminate(pid, driverName); err != nil {
			if err != os.ErrProcessDone {
				return err
			}
			// No process, stale pidfile.
			log.Debugf("Remove krunkit pidfile %q", pidfile)
			if err := os.Remove(pidfile); err != nil {
				log.Debugf("failed to remove %q: %s", pidfile, err)
			}
		}
	}
	return nil
}

func (d *Driver) killKrunkit() error {
	// krunkit does not support HardStop like vfkit.
	pidfile := d.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// No pidfile.
		return nil
	}
	log.Debugf("Kill krunkit (pid=%d)", pid)
	if err := process.Kill(pid, driverName); err != nil {
		if err != os.ErrProcessDone {
			return err
		}
		log.Debugf("Remove krunkit pidfile %q", pidfile)
		// No process, stale pidfile.
		if err := os.Remove(pidfile); err != nil {
			log.Debugf("failed to remove %q: %s", pidfile, err)
		}
	}
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return d.ResolveStorePath("disk.img")
}

func (d *Driver) pidfilePath() string {
	return d.ResolveStorePath(pidFileName)
}

func (d *Driver) serialPath() string {
	return d.ResolveStorePath(serialFileName)
}

func (d *Driver) isoPath() string {
	return d.ResolveStorePath(isoFileName)
}

func (d *Driver) sockfilePath() string {
	return d.ResolveStorePath(sockFileName)
}

func (d *Driver) restfulURI() string {
	return fmt.Sprintf("unix://%s", d.sockfilePath())
}

func (d *Driver) vmStateURI() string {
	return "http://_/vm/state"
}

// generateDiskImage generates a boot2docker VM disk image.
// TODO: duplicate from vfkit and qemu: https://github.com/kubernetes/minikube/issues/21090
func (d *Driver) generateDiskImage(size int) error {
	log.Debugf("Creating %d MB hard disk image...", size)

	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	pubKey, err := os.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	rawFile := d.diskPath()
	if err := os.WriteFile(rawFile, buf.Bytes(), 0644); err != nil {
		return nil
	}
	if err := os.Truncate(rawFile, int64(size)*int64(1024*1024)); err != nil {
		return nil
	}
	log.Debugf("DONE writing to %s and %s", rawFile, d.diskPath())
	return nil
}

func httpUnixClient(path string) http.Client {
	return http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", path)
			},
		},
	}
}

type vmState struct {
	State string `json:"state"`
}

func (d *Driver) setKrunkitState(desired string) error {
	var vmstate vmState
	vmstate.State = desired
	log.Infof("Set krunkit state: %+v", vmstate)
	data, err := json.Marshal(&vmstate)
	if err != nil {
		return err
	}
	httpc := httpUnixClient(d.sockfilePath())
	_, err = httpc.Post(d.vmStateURI(), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	return nil
}
