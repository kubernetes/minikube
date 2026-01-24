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
	"sync"
	"syscall"

	"github.com/blang/semver/v4"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"

	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/process"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	pidfileName  = "vmnet-helper.pid"
	logfileName  = "vmnet-helper.log"
	sockfileName = "vmnet-helper.sock"

	// Installed from GitHub releases.
	installPath = "/opt/vmnet-helper/bin/vmnet-helper"
	// Installed via Homebrew (macOS 26+ only).
	brewInstallPath = "/opt/homebrew/opt/vmnet-helper/libexec/vmnet-helper"
)

// Helper manages the vmnet-helper process.
type Helper struct {
	// The pidfile and log are stored here.
	MachineDir string

	// InterfaceID is a random UUID for the vmnet interface. Using the same UUID
	// will obtain the same MAC address from vmnet.
	InterfaceID string

	// Offloading is required for krunkit, does not work with vfkit.
	Offloading bool

	// Set when vmnet interface is started.
	macAddress string

	// NeedsSudo indicates whether sudo was used when starting the helper. Set
	// by Start(), used by Stop(), Kill(), and GetState(). Required for managing
	// the helper child process using the pid file.
	NeedsSudo bool
}

// helperVersion is the version of vmnet-helper.
type helperVersion struct {
	Version string `yaml:"version"`
	Commit  string `yaml:"commit"`
}

// helperInfo contains cached information about vmnet-helper.
type helperInfo struct {
	Path      string
	Version   helperVersion
	NeedsSudo bool
	Err       error
}

var (
	cached helperInfo
	once   sync.Once
)

// getHelperInfo returns cached information about vmnet-helper, initializing it
// on the first call.
func getHelperInfo() (helperInfo, error) {
	once.Do(func() {
		cached.Path, cached.Err = findHelper()
		if cached.Err != nil {
			return
		}
		cached.Version, cached.Err = getHelperVersion(cached.Path)
		if cached.Err != nil {
			return
		}
		var macosVersion string
		macosVersion, cached.Err = macOSVersion()
		if cached.Err != nil {
			return
		}
		cached.NeedsSudo, cached.Err = helperNeedsSudo(cached.Version, macosVersion)
	})
	return cached, cached.Err
}

type interfaceInfo struct {
	MACAddress string `json:"vmnet_mac_address"`
}

// ValidateHelper validates that vmnet-helper is installed and configured
// correctly.
//
// If the helper needs sudo, we validate that we can run vmnet-helper without a
// password, or fallback to interactive sudo and update the user's cached sudo
// credentials.
//
// If we fail with an expected error the returned error.Kind can be used to
// provide a suggestion for resolving the issue.
func ValidateHelper(options *run.CommandOptions) error {
	// Ideally minikube will not try to validate in download-only mode, but this
	// is called from different places in different drivers, so the easier way
	// to skip validation is to skip it here.
	if options.DownloadOnly {
		log.Debug("Skipping vmnet-helper validation in download-only mode")
		return nil
	}

	helper, err := getHelperInfo()
	if err != nil {
		return err
	}

	if helper.NeedsSudo {
		if err := validateRunningWithSudo(helper.Path, options); err != nil {
			return err
		}
	}

	log.Debugf("Validated vmnet-helper (path=%q, version=%q, commit=%q, needsSudo=%v)",
		helper.Path, helper.Version.Version, helper.Version.Commit, helper.NeedsSudo)

	return nil
}

// Start the vmnet-helper child process, creating the vmnet interface for the
// machine. The helper will create a unix datagram socket at the specified path.
// The client (e.g. vfkit) will connect to this socket.
func (h *Helper) Start(socketPath string) error {
	helper, err := getHelperInfo()
	if err != nil {
		return err
	}

	// Persist for Stop(), Kill(), and GetState().
	h.NeedsSudo = helper.NeedsSudo

	var executable string
	var args []string

	if helper.NeedsSudo {
		executable = "sudo"
		args = append(args, "--non-interactive", helper.Path)
	} else {
		executable = helper.Path
	}

	args = append(args, "--socket", socketPath, "--interface-id", h.InterfaceID)

	if h.Offloading {
		args = append(args, "--enable-tso", "--enable-checksum-offload")
	}

	cmd := exec.Command(executable, args...)

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
		if data, err := os.ReadFile(logfile.Name()); err == nil {
			log.Infof("vmnet-helper logfile %q content:\n%s", logfile.Name(), string(data))
		} else {
			log.Infof("failed to read vmnet-helper logfile %q: %s", logfile.Name(), err)
		}
		return fmt.Errorf("failed to decode vmnet interface info: %w", err)
	}

	log.Infof("Got mac address %q", info.MACAddress)
	h.macAddress = info.MACAddress

	return nil
}

// GetMACAddress returns the mac address assigned by vmnet framework.
func (h *Helper) GetMACAddress() string {
	return h.macAddress
}

// Stop terminates the executable. If running with sudo, sudo will terminate the
// helper.
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
	name := h.executableName()
	log.Debugf("Terminate %s (pid=%v)", name, pid)
	if err := process.Terminate(pid, name); err != nil {
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

// Kill the entire process group. If running with sudo, both sudo and
// vmnet-helper will be killed.
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
	name := h.executableName()
	exists, err := process.Exists(pid, name)
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

// GetState returns the child process state.
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
	name := h.executableName()
	exists, err := process.Exists(pid, name)
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

// executableName returns the name of the executable used by the last Start()
// call.  Required for checking process state and terminating it. Returns "sudo"
// if vmnet-helper was started with sudo, otherwise "vmnet-helper".
func (h *Helper) executableName() string {
	if h.NeedsSudo {
		return "sudo"
	}
	return "vmnet-helper"
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
// MiB buffer to make ENOBUFS errors less likely for the peer and allows queueing
// more packets when using the vmnet_enable_tso option.
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

// validateRunningWithSudo validates that we can run vmnet-helper with sudo
// without a password. If running with sudo fails, and we run in interactive
// mode, we fall back to interactive sudo and update the user's cached sudo
// credentials. This ensures that the next attempt to run vmnet-helper with sudo
// will succeed in the next 5 minutes (default sudo timeout).
func validateRunningWithSudo(helperPath string, options *run.CommandOptions) error {
	cmd := exec.Command("sudo", "--non-interactive", helperPath, "--version")
	if _, err := cmd.Output(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			err = fmt.Errorf("%w: %s", err, stderr)
		}
		// If we are not interactive, we can't authenticate the user, so we fail.
		if options.NonInteractive {
			return &Error{Kind: reason.NotConfiguredVmnetHelper, Err: err}
		}

		log.Debugf("Unable to run vmnet-helper without a password: %v", err)

		// We can fall back to interactive sudo this time, but the user should
		// configure a sudoers rule.
		out.ErrT(style.Tip, "Unable to run vmnet-helper without a password")
		out.ErrT(style.Indent, "To configure vmnet-helper to run without a password, please check the documentation:")
		out.ErrT(style.Indent, "https://github.com/nirs/vmnet-helper/#granting-permission-to-run-vmnet-helper")

		// Authenticate the user, updating the user's cached credentials.
		cmd = exec.Command("sudo", "--validate")
		if _, err := cmd.Output(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := strings.TrimSpace(string(exitErr.Stderr))
				err = fmt.Errorf("%w: %s", err, stderr)
			}
			// If we fail to authenticate the user, we can't run vmnet-helper
			// with sudo so we must fail.
			return &Error{Kind: reason.NotConfiguredVmnetHelper, Err: err}
		}

		log.Debugf("Authenticated user with sudo")
		return nil
	}

	log.Debug("Validated running vmnet-helper without a password")
	return nil
}

// findHelper finds the path to the vmnet-helper executable.
func findHelper() (string, error) {
	paths := []string{brewInstallPath, installPath}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return "", &Error{Kind: reason.HostPathStat, Err: err}
			}
			continue
		}
		return path, nil
	}
	err := fmt.Errorf("failed to find vmnet-helper at %q", paths)
	return "", &Error{Kind: reason.NotFoundVmnetHelper, Err: err}
}

// helperNeedsSudo returns true if vmnet-helper needs sudo to run based on the
// helper version and macOS version.
func helperNeedsSudo(version helperVersion, macosVersion string) (bool, error) {
	// ParseTolerant handles "26.2" by normalizing it to "26.2.0".
	macVer, err := semver.ParseTolerant(macosVersion)
	if err != nil {
		return false, fmt.Errorf("invalid macOS version %q: %w", macosVersion, err)
	}
	if macVer.LT(semver.MustParse("26.0.0")) {
		return true, nil
	}

	// semver.Parse does not support "v" prefix.
	helperVer, err := semver.Parse(strings.TrimPrefix(version.Version, "v"))
	if err != nil {
		return false, fmt.Errorf("invalid helper version %q: %w", version.Version, err)
	}

	// Since v0.9.0, vmnet-helper is signed with the
	// 'com.apple.security.virtualization' entitlement and does not need root.
	return helperVer.LT(semver.MustParse("0.9.0")), nil
}

// macOSVersion returns the macOS product version string. The format is
// "major.minor[.patch]"
func macOSVersion() (string, error) {
	version, err := unix.Sysctl("kern.osproductversion")
	if err != nil {
		return "", fmt.Errorf("failed to get macOS version: %w", err)
	}
	return version, nil
}

// getHelperVersion returns the version of vmnet-helper.
func getHelperVersion(executablePath string) (helperVersion, error) {
	cmd := exec.Command(executablePath, "--version")
	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%w: %s", err, exitErr.Stderr)
		}
		return helperVersion{}, fmt.Errorf("failed to get vmnet-helper version: %w", err)
	}
	return parseHelperVersion(stdout)
}

// parseHelperVersion parses vmnet-helper version output.
func parseHelperVersion(stdout []byte) (helperVersion, error) {
	var version helperVersion

	// Unmarshal current format (>= v0.7.0): "version: v0.7.0\ncommit: 7ee60de20...\n"
	if err := yaml.Unmarshal(stdout, &version); err != nil {
		// Fallback for older helper (< v0.7.0): "v0.6.0\n"
		version.Version = strings.TrimSpace(string(stdout))
	}

	if version.Version == "" {
		return version, fmt.Errorf("failed to parse vmnet-helper version: %q", stdout)
	}

	return version, nil
}
