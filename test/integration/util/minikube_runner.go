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
func (m *MinikubeRunner) teeRun(cmd *exec.Cmd) (string, string, error) {
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

// TODO medya
// RunCommand executes a command, optionally checking for error
func (m *MinikubeRunner) RunCommand(command string, checkError bool) string {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	Logf("Run: %s", cmd.Args)
	stdout, stderr, err := m.teeRun(cmd)
	if checkError && err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			m.T.Fatalf("Error running command: %s %s. Output: %s", command, exitError.Stderr, stdout)
		} else {
			m.T.Fatalf("Error running command: %s %v. Output: %s", command, err, stderr)
		}
	}
	return stdout
}

// RunWithContext calls the minikube command with a context, useful for timeouts.
func (m *MinikubeRunner) RunWithContext(ctx context.Context, command string) (string, string, error) {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.CommandContext(ctx, path, commandArr...)
	Logf("Run: %s", cmd.Args)
	return m.teeRun(cmd)
}

// RunDaemon executes a command, returning the stdout
func (m *MinikubeRunner) RunDaemon(command string) (*exec.Cmd, *bufio.Reader) {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.T.Fatalf("stdout pipe failed: %s %v", command, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.T.Fatalf("stderr pipe failed: %s %v", command, err)
	}

	var errB bytes.Buffer
	go func() {
		if err := commonutil.TeePrefix(commonutil.ErrPrefix, stderrPipe, &errB, Logf); err != nil {
			m.T.Logf("tee: %v", err)
		}
	}()

	err = cmd.Start()
	if err != nil {
		m.T.Fatalf("Error running command: %s %v", command, err)
	}
	return cmd, bufio.NewReader(stdoutPipe)

}

// RunDaemon2 executes a command, returning the stdout and stderr
func (m *MinikubeRunner) RunDaemon2(command string) (*exec.Cmd, *bufio.Reader, *bufio.Reader) {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.T.Fatalf("stdout pipe failed: %s %v", command, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		m.T.Fatalf("stderr pipe failed: %s %v", command, err)
	}

	err = cmd.Start()
	if err != nil {
		m.T.Fatalf("Error running command: %s %v", command, err)
	}
	return cmd, bufio.NewReader(stdoutPipe), bufio.NewReader(stderrPipe)
}

// TODO Medya
// SSH returns the output of running a command using SSH
func (m *MinikubeRunner) SSH(command string) (string, error) {
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, "ssh", command)
	Logf("SSH: %s", command)

	stdout, err := cmd.CombinedOutput()
	Logf("Output: %s", stdout)
	if err, ok := err.(*exec.ExitError); ok {
		return string(stdout), err
	}
	return string(stdout), nil
}

// Start starts the cluster
func (m *MinikubeRunner) Start(opts ...string) {
	cmd := fmt.Sprintf("start %s %s %s --alsologtostderr --v=2", m.StartArgs, m.GlobalArgs, strings.Join(opts, " "))
	m.RunCommand(cmd, true)
}

// EnsureRunning makes sure the container runtime is running
func (m *MinikubeRunner) EnsureRunning() {
	if m.GetStatus() != "Running" {
		m.Start()
	}
	m.CheckStatus("Running")
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
