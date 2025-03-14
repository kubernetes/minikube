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

// package vmnet provides the helper process connecting virtual machines to the
// vmnet network.
package vmnet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	pkgdrivers "k8s.io/minikube/pkg/drivers"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
)

const (
	pidfileName    = "vmnet-helper.pid"
	logfileName    = "vmnet-helper.log"
	executablePath = "/opt/vmnet-helper/bin/vmnet-helper"
)

// Helper manages the vmnet-helper process.
type Helper struct {
	// The pidfile and log are stored here.
	MachineDir string

	// InterfaceID is a random UUID for the vmnet interface. Using the same UUID
	// will obtain the same MAC address from vmnet.
	InterfaceID string

	// Set when vmnet interface is started.
	macAddress string
}

type interfaceInfo struct {
	MACAddress string `json:"vmnet_mac_address"`
}

// HelperAvailable tells if vmnet-helper executable is installed and configured
// correctly.
func HelperAvailable() bool {
	version, err := exec.Command("sudo", "--non-interactive", executablePath, "--version").Output()
	if err != nil {
		log.Debugf("Failed to run vmnet-helper: %w", err)
		return false
	}
	log.Debugf("Using vmnet-helper version %q", version)
	return true
}

// Start the vmnet-helper child process, creating the vmnet interface for the
// machine. sock is a connected unix datagram socket to pass the helper child
// process.
func (h *Helper) Start(sock *os.File) error {
	cmd := exec.Command(
		"sudo",
		"--non-interactive",
		"--close-from", fmt.Sprintf("%d", sock.Fd()+1),
		executablePath,
		"--fd", fmt.Sprintf("%d", sock.Fd()),
		"--interface-id", h.InterfaceID,
	)

	cmd.ExtraFiles = []*os.File{sock}

	// Create vmnet-helper in a new process group so it is not harmed when
	// terminating the minikube process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logfile, err := h.openLogfile()
	if err != nil {
		return fmt.Errorf("failed to open helper logfile: %w", err)
	}
	defer logfile.Close()
	cmd.Stderr = logfile

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create helper stdout pipe: %w", err)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start vmnet-helper: %w", err)
	}

	log.Infof("Started vmnet-helper (pid=%v)", cmd.Process.Pid)

	if err := pkgdrivers.WritePidfile(h.pidfilePath(), cmd.Process.Pid); err != nil {
		return fmt.Errorf("failed to write vmnet-helper pidfile: %w", err)
	}

	var info interfaceInfo
	if err := json.NewDecoder(stdout).Decode(&info); err != nil {
		return fmt.Errorf("failed to decode vmnet interface info: %w", err)
	}

	log.Infof("Got mac address %q", info.MACAddress)
	h.macAddress = info.MACAddress

	return nil
}

// GetMACAddress reutuns the mac address assigned by vmnet framework.
func (h *Helper) GetMACAddress() string {
	return h.macAddress
}

// Stop terminate sudo, which will terminate vmnet-helper.
func (h *Helper) Stop() error {
	log.Info("Stop vmnet-helper")
	return pkgdrivers.SignalPidfile(h.pidfilePath(), syscall.SIGTERM)
}

// Kill both sudo and vmnet-helper by killing the process group.
func (h *Helper) Kill() error {
	pidfile := h.pidfilePath()
	pid, err := pkgdrivers.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// Already stopped.
		os.Remove(pidfile)
		return nil
	}
	log.Infof("Kill vmnet-helper process group (pgid=%v)", pid)
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if err != syscall.ESRCH {
			return err
		}
		// Process done.
		os.Remove(pidfile)
		return nil
	}
	return nil
}

// GetState returns the vmnet-helper child process state.
func (h *Helper) GetState() (state.State, error) {
	pidfile := h.pidfilePath()
	pid, err := pkgdrivers.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return state.Error, err
		}
		return state.Stopped, nil
	}
	if err := pkgdrivers.CheckPid(pid); err != nil {
		// No pid, remove pidfile
		os.Remove(pidfile)
		return state.Stopped, nil
	}
	return state.Running, nil
}

func (h *Helper) openLogfile() (*os.File, error) {
	logfile := filepath.Join(h.MachineDir, logfileName)
	return os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}

func (h *Helper) pidfilePath() string {
	return filepath.Join(h.MachineDir, pidfileName)
}

// Apple recommend receive buffer size to be 4 times the size of the send buffer
// size, but send buffer size is not used to allocate a buffer in datagram
// sockets, it only limits the maximum packet size. Must be larger than TSO
// packets size (65550 bytes).
const sendBufferSize = 65 * 1024

// The receive buffer size determine how many packets can be queued by the
// peer. Using bigger receive buffer size make ENOBUFS error less likely for the
// peer and improves throughput.
const recvBufferSize = 4 * 1024 * 1024

// Socketpair returns a pair of connected unix datagram sockets that can be used
// to connect the helper and a vm. Pass one socket to the helper child process
// and the other to the vm child process.
func Socketpair() (*os.File, *os.File, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, nil, err
	}
	// Setting buffer size is an optimization - don't fail on errors.
	for _, fd := range fds {
		_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, sendBufferSize)
		_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, recvBufferSize)
	}
	return os.NewFile(uintptr(fds[0]), "sock1"), os.NewFile(uintptr(fds[1]), "sock2"), nil
}
