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

package util

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/assets"
	commonutil "k8s.io/minikube/pkg/util"
)

// MinikubeRunner runs a command
type MinikubeRunner struct {
	Profile    string
	T          *testing.T
	BinaryPath string
	GlobalArgs string
	StartArgs  string
	MountArgs  string
	Runtime    string
}

// Copy copies a file
func (m *MinikubeRunner) Copy(f assets.CopyableFile) error {
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command("/bin/bash", "-c", path, "ssh", "--", fmt.Sprintf("cat >> %s", filepath.Join(f.GetTargetDir(), f.GetTargetName())))
	Logf("Running: %s", cmd.Args)
	return cmd.Run()
}

// Remove removes a file
func (m *MinikubeRunner) Remove(f assets.CopyableFile) error {
	_, err := m.SSH(fmt.Sprintf("rm -rf %s", filepath.Join(f.GetTargetDir(), f.GetTargetName())))
	return err
}

// teeRun runs a command, streaming stdout, stderr to console
func (m *MinikubeRunner) teeRun(cmd *exec.Cmd, wait ...bool) (string, string, error) {
	w := true
	if wait != nil {
		w = wait[0]
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}

	if err := cmd.Start(); err != nil {
		return "", "", err
	}
	if w {
		var outB bytes.Buffer
		var errB bytes.Buffer
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			if err := commonutil.TeePrefix(commonutil.ErrPrefix, errPipe, &errB, Logf); err != nil {
				m.T.Logf("tee: %v", err)
			}
			wg.Done()
		}()
		go func() {
			if err := commonutil.TeePrefix(commonutil.OutPrefix, outPipe, &outB, Logf); err != nil {
				m.T.Logf("tee: %v", err)
			}
			wg.Done()
		}()
		err = cmd.Wait()
		wg.Wait()
		return outB.String(), errB.String(), err

	}
	return "", "", err
}

// RunCommand executes a command, optionally checking for error
func (m *MinikubeRunner) RunCommand(cmdStr string, failError bool, wait ...bool) string {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, cmdArgs...)
	Logf("Run: %s", cmd.Args)
	stdout, stderr, err := m.teeRun(cmd, wait...)
	if failError && err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			m.T.Fatalf("Error running command: %s %s. Output: %s", cmdStr, exitError.Stderr, stdout)
		} else {
			m.T.Fatalf("Error running command: %s %v. Output: %s", cmdStr, err, stderr)
		}
	}
	return stdout
}

// RunWithContext calls the minikube command with a context, useful for timeouts.
func (m *MinikubeRunner) RunWithContext(ctx context.Context, cmdStr string, wait ...bool) (string, string, error) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.CommandContext(ctx, path, cmdArgs...)
	Logf("Run: %s", cmd.Args)
	return m.teeRun(cmd, wait...)
}

// RunDaemon executes a command, returning the stdout
func (m *MinikubeRunner) RunDaemon(cmdStr string) (*exec.Cmd, *bufio.Reader) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, cmdArgs...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.T.Fatalf("stdout pipe failed: %s %v", cmdStr, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.T.Fatalf("stderr pipe failed: %s %v", cmdStr, err)
	}

	var errB bytes.Buffer
	go func() {
		if err := commonutil.TeePrefix(commonutil.ErrPrefix, stderrPipe, &errB, Logf); err != nil {
			m.T.Logf("tee: %v", err)
		}
	}()

	err = cmd.Start()
	if err != nil {
		m.T.Fatalf("Error running command: %s %v", cmdStr, err)
	}
	return cmd, bufio.NewReader(stdoutPipe)

}

// RunDaemon2 executes a command, returning the stdout and stderr
func (m *MinikubeRunner) RunDaemon2(cmdStr string) (*exec.Cmd, *bufio.Reader, *bufio.Reader) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, cmdArgs...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.T.Fatalf("stdout pipe failed: %s %v", cmdStr, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.T.Fatalf("stderr pipe failed: %s %v", cmdStr, err)
	}

	err = cmd.Start()
	if err != nil {
		m.T.Fatalf("Error running command: %s %v", cmdStr, err)
	}
	return cmd, bufio.NewReader(stdoutPipe), bufio.NewReader(stderrPipe)
}

// SSH returns the output of running a command using SSH
func (m *MinikubeRunner) SSH(cmdStr string) (string, error) {
	profileArg := fmt.Sprintf("-p=%s", m.Profile)
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, profileArg, "ssh", cmdStr)
	Logf("SSH: %s", cmdStr)
	stdout, err := cmd.CombinedOutput()
	Logf("Output: %s", stdout)
	if err, ok := err.(*exec.ExitError); ok {
		return string(stdout), err
	}
	return string(stdout), nil
}

// Start starts the cluster
func (m *MinikubeRunner) Start(opts ...string) string {
	cmd := fmt.Sprintf("start %s %s %s --alsologtostderr --v=2", m.StartArgs, m.GlobalArgs, strings.Join(opts, " "))
	return m.RunCommand(cmd, true)
}

// StartWithStds starts the cluster with console output without verbose log
func (m *MinikubeRunner) StartWithStds(timeout time.Duration, opts ...string) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := fmt.Sprintf("start %s %s %s", m.StartArgs, m.GlobalArgs, strings.Join(opts, " "))
	return m.RunWithContext(ctx, cmd)
}

// Delete deletes the minikube profile to be used to clean up after a test
func (m *MinikubeRunner) Delete(wait ...bool) string {
	return m.RunCommand("delete", true, wait...)
}

// EnsureRunning makes sure the container runtime is running
func (m *MinikubeRunner) EnsureRunning(opts ...string) {
	if m.GetStatus() != state.Running.String() {
		m.Start(opts...)
	}
	m.CheckStatus(state.Running.String())
}

// ParseEnvCmdOutput parses the output of `env` (assumes bash)
func (m *MinikubeRunner) ParseEnvCmdOutput(out string) map[string]string {
	env := map[string]string{}
	re := regexp.MustCompile(`(\w+?) ?= ?"?(.+?)"?\n`)
	for _, m := range re.FindAllStringSubmatch(out, -1) {
		env[m[1]] = m[2]
	}
	return env
}

// GetStatus returns the status of a service
func (m *MinikubeRunner) GetStatus() string {
	return m.RunCommand(fmt.Sprintf("status --format={{.Host}} %s", m.GlobalArgs), false)
}

// GetLogs returns the logs of a service
func (m *MinikubeRunner) GetLogs() string {
	return m.RunCommand(fmt.Sprintf("logs %s", m.GlobalArgs), true)
}

// CheckStatus makes sure the service has the desired status, or cause fatal error
func (m *MinikubeRunner) CheckStatus(desired string) {
	if err := m.CheckStatusNoFail(desired); err != nil {
		m.T.Fatalf("%v", err)
	}
}

// CheckStatusNoFail makes sure the service has the desired status, returning error
func (m *MinikubeRunner) CheckStatusNoFail(desired string) error {
	s := m.GetStatus()
	if s != desired {
		return fmt.Errorf("got state: %q, expected %q", s, desired)
	}
	return nil
}
