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
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/firewall"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	isoFilename    = "boot2docker.iso"
	pidFileName    = "vfkit.pid"
	sockFilename   = "vfkit.sock"
	defaultSSHUser = "docker"
)

// Driver is the machine driver for vfkit (Virtualization.framework)
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	Boot2DockerURL string
	DiskSize       int
	CPU            int
	Memory         int
	Cmdline        string
	MACAddress     string
	ExtraDisks     int
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
		CommonDriver: &pkgdrivers.CommonDriver{},
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

func checkPid(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.Signal(0))
}

func (d *Driver) GetState() (state.State, error) {
	if _, err := os.Stat(d.pidfilePath()); err != nil {
		return state.Stopped, nil
	}
	p, err := os.ReadFile(d.pidfilePath())
	if err != nil {
		return state.Error, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(p)))
	if err != nil {
		return state.Error, err
	}
	if err := checkPid(pid); err != nil {
		// No pid, remove pidfile
		os.Remove(d.pidfilePath())
		return state.Stopped, nil
	}
	ret, err := d.GetVFKitState()
	if err != nil {
		return state.Error, err
	}
	switch ret {
	case "running", "VirtualMachineStateRunning":
		return state.Running, nil
	case "stopped", "VirtualMachineStateStopped":
		return state.Stopped, nil
	}
	return state.None, nil
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
	isoPath := d.ResolveStorePath(isoFilename)

	log.Info("Extracting Kernel...")
	if err := d.extractKernel(isoPath); err != nil {
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
			path := pkgdrivers.ExtraDiskPath(d.BaseDriver, i)
			if err := pkgdrivers.CreateRawDisk(path, d.DiskSize); err != nil {
				return err
			}
		}
	}

	log.Info("Starting vfkit VM...")
	return d.Start()
}

func (d *Driver) Start() error {
	machineDir := filepath.Join(d.StorePath, "machines", d.GetMachineName())

	var startCmd []string

	startCmd = append(startCmd,
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--cpus", fmt.Sprintf("%d", d.CPU),
		"--restful-uri", fmt.Sprintf("unix://%s", d.sockfilePath()))
	var isoPath = filepath.Join(machineDir, isoFilename)
	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-blk,path=%s", isoPath))

	var mac = d.MACAddress
	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-net,nat,mac=%s", mac))

	startCmd = append(startCmd,
		"--device", "virtio-rng")

	startCmd = append(startCmd,
		"--kernel", d.ResolveStorePath("bzimage"))
	startCmd = append(startCmd,
		"--kernel-cmdline", d.Cmdline)
	startCmd = append(startCmd,
		"--initrd", d.ResolveStorePath("initrd"))

	for i := 0; i < d.ExtraDisks; i++ {
		startCmd = append(startCmd,
			"--device", fmt.Sprintf("virtio-blk,path=%s", pkgdrivers.ExtraDiskPath(d.BaseDriver, i)))
	}

	startCmd = append(startCmd,
		"--device", fmt.Sprintf("virtio-blk,path=%s", d.diskPath()))

	log.Debugf("executing: vfkit %s", strings.Join(startCmd, " "))
	os.Remove(d.sockfilePath())
	cmd := exec.Command("vfkit", startCmd...)

	// Create vfkit in a new process group, so minikube caller can use killpg
	// to terminate the entire process group without harming the vfkit process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	if err := os.WriteFile(d.pidfilePath(), []byte(fmt.Sprintf("%v", pid)), 0600); err != nil {
		return err
	}

	if err := d.setupIP(mac); err != nil {
		return err
	}

	log.Infof("Waiting for VM to start (ssh -p %d docker@%s)...", d.SSHPort, d.IPAddress)

	return WaitForTCPWithDelay(fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), time.Second)
}

func (d *Driver) setupIP(mac string) error {
	var err error
	getIP := func() error {
		d.IPAddress, err = pkgdrivers.GetIPAddressByMACAddress(mac)
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
	if unblockErr := firewall.UnblockBootpd(); unblockErr != nil {
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
	if err := d.SetVFKitState("HardStop"); err != nil {
		return err
	}
	return nil
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
	if s != state.Stopped {
		if err := d.SetVFKitState("Stop"); err != nil {
			return errors.Wrap(err, "quit")
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

func (d *Driver) extractKernel(isoPath string) error {
	for _, f := range []struct {
		pathInIso string
		destPath  string
	}{
		{"/boot/bzimage", "bzimage"},
		{"/boot/initrd", "initrd"},
	} {
		fullDestPath := d.ResolveStorePath(f.destPath)
		if err := pkgdrivers.ExtractFile(isoPath, f.pathInIso, fullDestPath); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Kill() error {
	if err := d.SetVFKitState("HardStop"); err != nil {
		return err
	}
	return nil
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

func (d *Driver) SetVFKitState(state string) error {
	httpc := httpUnixClient(d.sockfilePath())
	var vmstate VMState
	vmstate.State = state
	data, err := json.Marshal(&vmstate)
	if err != nil {
		return err
	}
	_, err = httpc.Post("http://_/vm/state", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	log.Debugf("set state: %+v", vmstate)
	return nil
}

func WaitForTCPWithDelay(addr string, duration time.Duration) error {
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		if _, err := conn.Read(make([]byte, 1)); err != nil && err != io.EOF {
			time.Sleep(duration)
			continue
		}
		break
	}
	return nil
}
