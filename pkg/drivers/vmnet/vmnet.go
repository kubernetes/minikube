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
	"strings"
	"syscall"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/process"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	pidfileName    = "vmnet-helper.pid"
	logfileName    = "vmnet-helper.log"
	sockfileName   = "vmnet-helper.sock"
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

// ValidateHelper checks if vmnet-helper is installed and we can run it as root.
// The returned error.Kind can be used to provide a suggestion for resolving the
// issue.
func ValidateHelper() error {
	// Is it installed?
	if _, err := os.Stat(executablePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Error{Kind: reason.NotFoundVmnetHelper, Err: err}
		}
		return &Error{Kind: reason.HostPathStat, Err: err}
	}

	// Can we run it as root without a password?
	cmd := exec.Command("sudo", "--non-interactive", executablePath, "--version")
	stdout, err := cmd.Output()
	if err != nil {
		// Can we interact with the user?
		if !viper.GetBool("interactive") {
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := strings.TrimSpace(string(exitErr.Stderr))
				err = fmt.Errorf("%w: %s", err, stderr)
			}
			return &Error{Kind: reason.NotConfiguredVmnetHelper, Err: err}
		}

		// We can fall back to intereactive sudo this time, but the user should
		// configure a sudoers rule.
		out.ErrT(style.Tip, "Unable to run vmnet-helper without a password")
		out.ErrT(style.Indent, "To configure vment-helper to run without a password, please check the documentation:")
		out.ErrT(style.Indent, "https://github.com/nirs/vmnet-helper/#granting-permission-to-run-vmnet-helper")

		// Authenticate the user, updateing the user's cached credentials.
		cmd = exec.Command("sudo", executablePath, "--version")
		stdout, err = cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := strings.TrimSpace(string(exitErr.Stderr))
				err = fmt.Errorf("%w: %s", err, stderr)
			}
			return &Error{Kind: reason.NotConfiguredVmnetHelper, Err: err}
		}
	}

	version := strings.TrimSpace(string(stdout))
	log.Debugf("Validated vmnet-helper version %q", version)
	return nil
}

// Start the vmnet-helper child process, creating the vmnet interface for the
// machine. The helper will create a unix datagram socket at the specfied path.
// The client (e.g. vfkit) will connect to this socket.
func (h *Helper) Start(socketPath string) error {
	cmd := exec.Command(
		"sudo",
		"--non-interactive",
		executablePath,
		"--socket", socketPath,
		"--interface-id", h.InterfaceID,
	)

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

	if err := process.WritePidfile(h.pidfilePath(), cmd.Process.Pid); err != nil {
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

// Stop terminates sudo, which will terminate vmnet-helper.
func (h *Helper) Stop() error {
	log.Info("Stop vmnet-helper")
	pidfile := h.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// No pidfile.
		return nil
	}
	log.Debugf("Terminate sudo (pid=%v)", pid)
	if err := process.Terminate(pid, "sudo"); err != nil {
		if err != os.ErrProcessDone {
			return err
		}
		// No process, stale pidfile.
		if err := os.Remove(pidfile); err != nil {
			log.Debugf("failed to remove %q: %s", pidfile, err)
		}
	}
	return nil
}

// Kill both sudo and vmnet-helper by killing the process group.
func (h *Helper) Kill() error {
	log.Info("Kill vmnet-helper")
	pidfile := h.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// No pidfile.
		return nil
	}
	exists, err := process.Exists(pid, "sudo")
	if err != nil {
		return err
	}
	if !exists {
		// No process, stale pidfile.
		if err := os.Remove(pidfile); err != nil {
			log.Debugf("failed to remove %q: %s", pidfile, err)
		}
		return nil
	}
	log.Debugf("Kill vmnet-helper process group (pgid=%v)", pid)
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if err != syscall.ESRCH {
			return err
		}
		// No process, stale pidfile.
		if err := os.Remove(pidfile); err != nil {
			log.Debugf("failed to remove %q: %s", pidfile, err)
		}
	}
	return nil
}

// GetState returns the sudo child process state.
func (h *Helper) GetState() (state.State, error) {
	pidfile := h.pidfilePath()
	pid, err := process.ReadPidfile(pidfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return state.Error, err
		}
		// No pidfile.
		return state.Stopped, nil
	}
	exists, err := process.Exists(pid, "sudo")
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

func (h *Helper) SocketPath() string {
	return filepath.Join(h.MachineDir, sockfileName)
}

func (h *Helper) openLogfile() (*os.File, error) {
	logfile := filepath.Join(h.MachineDir, logfileName)
	return os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
}

func (h *Helper) pidfilePath() string {
	return filepath.Join(h.MachineDir, pidfileName)
}

// Apple recommends sizing the receive buffer at 4 times the size of the send
// buffer, and other projects typically use a 1 MiB send buffer and a 4 MiB
// receive buffer. However the send buffer size is not used to allocate a buffer
// in datagram sockets, it only limits the maximum packet size. We use 65 KiB
// buffer to allow the largest possible packet size (65550 bytes) when using the
// vmnet_enable_tso option.
const sendBufferSize = 65 * 1024

// The receive buffer size determines how many packets can be queued by the
// peer. Testing shows good performance with a 2 MiB receive buffer. We use a 4
// MiB buffer to make ENOBUFS errors less likely for the peer and allowing to
// queue more packets when using the vmnet_enable_tso option.
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
