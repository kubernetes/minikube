//go:build darwin

/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package vfkit

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
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"errors"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/common"
	"k8s.io/minikube/pkg/drivers/common/virtiofs"
	"k8s.io/minikube/pkg/drivers/common/vmnet"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/firewall"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/process"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	isoFilename    = "boot2docker.iso"
	pidFileName    = "vfkit.pid"
	sockFilename   = "vfkit.sock"
	logFileName    = "vfkit.log"
	serialFileName = "serial.log"
	defaultSSHUser = "docker"

	// Rosetta mount in the guest.
	rosettaMountTag   = "minikube-rosetta"
	rosettaMountPoint = "/mnt/minikube-rosetta"
)

// Driver is the machine driver for vfkit (Virtualization.framework)
type Driver struct {
	*drivers.BaseDriver
	*common.CommonDriver
	Boot2DockerURL string
	DiskSize       int
	CPU            int
	Memory         int
	ExtraDisks     int
	VirtiofsMounts []*virtiofs.Mount
	Network        string        // "", "nat", "vmnet-shared"
	MACAddress     string        // For network=nat, network=""
	VmnetHelper    *vmnet.Helper // For network=vmnet-shared
	Rosetta        bool          // Enable rosetta support
}

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
	return "vfkit"
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

func (d *Driver) getVfkitState() (state.State, error) {
	pidfile := d.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return state.Error, err
		}
		return state.Stopped, nil
	}
	exists, err := process.Exists(pid, "vfkit")
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

func (d *Driver) getVmnetHelperState() (state.State, error) {
	if d.VmnetHelper == nil {
		return state.Stopped, nil
	}
	return d.VmnetHelper.GetState()
}

// GetState returns driver state. Since vfkit driver may use 2 processes
// (vmnet-helper, vfkit), this returns combined state of both processes.
func (d *Driver) GetState() (state.State, error) {
	if vfkitState, err := d.getVfkitState(); err != nil {
		return state.Error, err
	} else if vfkitState == state.Running {
		return state.Running, nil
	}
	return d.getVmnetHelperState()
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

	if err := d.extractKernel(); err != nil {
		return err
	}

	log.Info("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	log.Info("Creating Disk image...")
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

	log.Info("Starting vfkit VM...")
	return d.Start()
}

func (d *Driver) extractKernel() error {
	log.Info("Extracting bzimage and initrd...")
	if err := common.ExtractFile(d.isoPath(), "/boot/bzimage", d.kernelPath()); err != nil {
		return err
	}
	return common.ExtractFile(d.isoPath(), "/boot/initrd", d.initrdPath())
}

func (d *Driver) Start() error {
	var socketPath string

	if d.VmnetHelper != nil {
		socketPath = d.VmnetHelper.SocketPath()
		if err := d.VmnetHelper.Start(socketPath); err != nil {
			return err
		}

		d.MACAddress = d.VmnetHelper.GetMACAddress()
	}

	if err := d.startVfkit(socketPath); err != nil {
		return err
	}

	if err := d.setupIP(d.MACAddress); err != nil {
		return err
	}

	if err := common.WaitForSSHAccess(d); err != nil {
		return err
	}

	if d.Rosetta {
		if err := d.setupRosetta(); err != nil {
			return err
		}
	}

	if len(d.VirtiofsMounts) > 0 {
		log.Infof("Setup virtiofs mounts ...")
		if err := virtiofs.SetupMounts(d, d.VirtiofsMounts); err != nil {
			return err
		}
	}

	return nil
}

// startVfkit starts the vfkit child process. If socketPath is not empty, vfkit
// is connected to the vmnet network via the socket instead of "nat" network.
func (d *Driver) startVfkit(socketPath string) error {
	var startCmd []string

	startCmd = append(startCmd,
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--cpus", fmt.Sprintf("%d", d.CPU),
		"--restful-uri", fmt.Sprintf("unix://%s", d.sockfilePath()),
		"--log-level", "debug")

	// On arm64 console= is required get boot messages in serial.log. On x86_64
	// serial log is always empty.
	var cmdline string
	switch runtime.GOARCH {
	case "arm64":
		cmdline = "console=hvc0"
	case "amd64":
		cmdline = "console=ttyS0"
	}

	// TODO: Switch to --bootloader efi when x86_64 iso changed to EFI.
	startCmd = append(startCmd,
		"--bootloader", fmt.Sprintf("linux,kernel=%s,initrd=%s,cmdline=\"%s\"",
			d.kernelPath(), d.initrdPath(), cmdline))

	if socketPath != "" {
		// The guest will be able to access other guests in the vmnet network.
		startCmd = append(startCmd,
			"--device", fmt.Sprintf("virtio-net,unixSocketPath=%s,mac=%s", socketPath, d.MACAddress))
	} else {
		// The guest will not be able to access other guests.
		startCmd = append(startCmd,
			"--device", fmt.Sprintf("virtio-net,nat,mac=%s", d.MACAddress))
	}

	startCmd = append(startCmd,
		"--device", "virtio-rng")

	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-blk,path=%s", d.isoPath()))

	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-blk,path=%s", d.diskPath()))

	for i := 0; i < d.ExtraDisks; i++ {
		startCmd = append(startCmd,
			"--device", fmt.Sprintf("virtio-blk,path=%s", common.ExtraDiskPath(d.BaseDriver, i)))
	}

	serialPath := d.ResolveStorePath(serialFileName)
	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-serial,logFilePath=%s", serialPath))

	for _, mount := range d.VirtiofsMounts {
		startCmd = append(startCmd,
			"--device", fmt.Sprintf("virtio-fs,sharedDir=%s,mountTag=%s", mount.HostPath, mount.Tag))

	}

	if d.Rosetta {
		startCmd = append(startCmd, d.rosettaOptions()...)
	}

	log.Debugf("executing: vfkit %s", strings.Join(startCmd, " "))
	os.Remove(d.sockfilePath())
	cmd := exec.Command("vfkit", startCmd...)

	// Create vfkit in a new process group, so minikube caller can use killpg
	// to terminate the entire process group without harming the vfkit process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logfile, err := d.openLogfile()
	if err != nil {
		return fmt.Errorf("failed to open vfkit logfile: %w", err)
	}
	defer logfile.Close()
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		return err
	}
	return process.WritePidfile(d.pidfilePath(), cmd.Process.Pid)
}

// rosettaOptions returns the vfkit command line options for Rosetta support.
func (d *Driver) rosettaOptions() []string {
	options := []string{
		"rosetta",
		"mountTag=" + rosettaMountTag,
	}

	// Try to install rosetta automatically for best user experience. The
	// installation requires user interaction so we must skip it in
	// non-interactive mode. If Rosetta is not installed vfkit will fail.
	// For more info see https://support.apple.com/en-us/102527
	if !d.CommandOptions.NonInteractive {
		options = append(options, "install")
	}

	return []string{"--device", strings.Join(options, ",")}
}

func (d *Driver) setupRosetta() error {
	// See https://docs.kernel.org/admin-guide/binfmt-misc.html
	binfmt := strings.Join([]string{
		// name
		":rosetta",
		// type: M (magic number matching), E (extension matching)
		":M",
		// offset (default 0)
		":",
		// magic
		`:\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x3e\x00`,
		// mask
		`:\xff\xff\xff\xff\xff\xfe\xfe\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff`,
		// interpreter
		":" + filepath.Join(rosettaMountPoint, "rosetta"),
		// flags: F (fix binary), C (credentials), or O (open binary)
		":F",
	}, "")

	var b strings.Builder

	fmt.Fprintf(&b, "set -e\n")
	fmt.Fprintf(&b, "sudo mount binfmt_misc -t binfmt_misc /proc/sys/fs/binfmt_misc\n")
	fmt.Fprintf(&b, "sudo mkdir -p %s\n", rosettaMountPoint)
	fmt.Fprintf(&b, "sudo mount -t virtiofs %s %s\n", rosettaMountTag, rosettaMountPoint)
	fmt.Fprintf(&b, "echo '%s' | sudo tee /proc/sys/fs/binfmt_misc/register\n", binfmt)

	if _, err := drivers.RunSSHCommandFromDriver(d, b.String()); err != nil {
		return fmt.Errorf("failed to setup rosetta: %w", err)
	}

	return nil
}

func (d *Driver) setupIP(mac string) error {
	var err error
	getIP := func() error {
		d.IPAddress, err = common.GetIPAddressByMACAddress(mac)
		if err != nil {
			return fmt.Errorf("failed to get IP address: %w", err)
		}
		return nil
	}
	// Implement a retry loop because IP address isn't added to dhcp leases file immediately
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
		time.Sleep(2 * time.Second)
	}

	if err == nil {
		log.Debugf("IP: %s", d.IPAddress)
		return nil
	}
	if !isBootpdError(err) {
		return fmt.Errorf("IP address never found in dhcp leases file: %w", err)
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

func (d *Driver) openLogfile() (*os.File, error) {
	logfile := d.ResolveStorePath(logFileName)
	return os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}

func (d *Driver) stopVfkit() error {
	if err := d.SetVFKitState("Stop"); err != nil {
		// vfkit may be already stopped, shutting down, or not listening.
		// We don't fallback to "HardStop" since it typically fails due to
		// https://github.com/crc-org/vfkit/issues/277.
		log.Debugf("Failed to set vfkit state to 'Stop': %s", err)
		pidfile := d.pidfilePath()
		pid, err := process.ReadPidfile(pidfile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			// No pidfile.
			return nil
		}
		if err := process.Terminate(pid, "vfkit"); err != nil {
			if err != os.ErrProcessDone {
				return err
			}
			// No process, stale pidfile.
			if err := os.Remove(pidfile); err != nil {
				log.Debugf("failed to remove %q: %s", pidfile, err)
			}
			return nil
		}
	}
	return nil
}

func (d *Driver) stopVmnetHelper() error {
	if d.VmnetHelper == nil {
		return nil
	}
	return d.VmnetHelper.Stop()
}

func (d *Driver) Stop() error {
	if err := d.stopVfkit(); err != nil {
		return err
	}
	return d.stopVmnetHelper()
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return fmt.Errorf("kill: %w", err)
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

func (d *Driver) killVfkit() error {
	if err := d.SetVFKitState("HardStop"); err != nil {
		// Typically fails with EOF due to https://github.com/crc-org/vfkit/issues/277.
		log.Debugf("Failed to set vfkit state to 'HardStop': %s", err)
		pidfile := d.pidfilePath()
		pid, err := process.ReadPidfile(pidfile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			// No pidfile.
			return nil
		}
		if err := process.Kill(pid, "vfkit"); err != nil {
			if err != os.ErrProcessDone {
				return err
			}
			// No process, stale pidfile.
			if err := os.Remove(pidfile); err != nil {
				log.Debugf("failed to remove %q: %s", pidfile, err)
			}
			return nil
		}
	}
	return nil
}

func (d *Driver) killVmnetHelper() error {
	if d.VmnetHelper == nil {
		return nil
	}
	return d.VmnetHelper.Kill()
}

func (d *Driver) Kill() error {
	if err := d.killVfkit(); err != nil {
		return err
	}
	return d.killVmnetHelper()
}

func (d *Driver) StartDocker() error {
	return fmt.Errorf("hosts without a driver cannot start docker")
}

func (d *Driver) StopDocker() error {
	return fmt.Errorf("hosts without a driver cannot stop docker")
}

func (d *Driver) GetDockerConfigDir() string {
	return ""
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("hosts without a driver cannot be upgraded")
}

func (d *Driver) sshKeyPath() string {
	machineDir := filepath.Join(d.StorePath, "machines", d.GetMachineName())
	return filepath.Join(machineDir, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) isoPath() string {
	return d.ResolveStorePath(isoFilename)
}

func (d *Driver) kernelPath() string {
	return d.ResolveStorePath("bzimage")
}

func (d *Driver) initrdPath() string {
	return d.ResolveStorePath("initrd")
}

func (d *Driver) diskPath() string {
	machineDir := filepath.Join(d.StorePath, "machines", d.GetMachineName())
	return filepath.Join(machineDir, "disk.img")
}

func (d *Driver) sockfilePath() string {
	machineDir := filepath.Join(d.StorePath, "machines", d.GetMachineName())
	return filepath.Join(machineDir, sockFilename)
}

func (d *Driver) pidfilePath() string {
	machineDir := filepath.Join(d.StorePath, "machines", d.GetMachineName())
	return filepath.Join(machineDir, pidFileName)
}

// Make a boot2docker VM disk image.
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

type VMState struct {
	State string `json:"state"`
}

func (d *Driver) GetVFKitState() (string, error) {
	httpc := httpUnixClient(d.sockfilePath())
	var vmstate VMState
	response, err := httpc.Get("http://_/vm/state")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&vmstate)
	if err != nil {
		return "", err
	}
	log.Debugf("get state: %+v", vmstate)
	return vmstate.State, nil
}

// SetVFKitState sets the state of the vfkit VM, (s is the state)
func (d *Driver) SetVFKitState(s string) error {
	httpc := httpUnixClient(d.sockfilePath())
	var vmstate VMState
	vmstate.State = s
	data, err := json.Marshal(&vmstate)
	if err != nil {
		return err
	}
	_, err = httpc.Post("http://_/vm/state", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	log.Infof("Set vfkit state: %+v", vmstate)
	return nil
}
