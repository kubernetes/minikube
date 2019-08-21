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
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

// MinikubeRunner runs a command
type MinikubeRunner struct {
	Profile      string
	T            *testing.T
	BinaryPath   string
	GlobalArgs   string
	StartArgs    string
	MountArgs    string
	Runtime      string
	TimeOutStart time.Duration // time to wait for minikube start before killing it
}

// Remove removes a file
func (m *MinikubeRunner) Remove(f assets.CopyableFile) error {
	_, err := m.SSH(fmt.Sprintf("rm -rf %s", filepath.Join(f.GetTargetDir(), f.GetTargetName())))
	return err
}

// teeRun runs a command, streaming stdout, stderr to console
func (m *MinikubeRunner) teeRun(cmd *exec.Cmd, waitForRun ...bool) (string, string, error) {
	w := true
	if waitForRun != nil {
		w = waitForRun[0]
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

// RunCommand executes a command, optionally checking for error and by default waits for run to finish
func (m *MinikubeRunner) RunCommand(cmdStr string, failError bool, waitForRun ...bool) (string, string) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, cmdArgs...)
	Logf("Run: %s", cmd.Args)
	stdout, stderr, err := m.teeRun(cmd, waitForRun...)
	if err != nil {
		errMsg := ""
		if exitError, ok := err.(*exec.ExitError); ok {
			errMsg = fmt.Sprintf("Error running command: %s %s. Output: %s Stderr: %s", cmdStr, exitError.Stderr, stdout, stderr)
		} else {
			errMsg = fmt.Sprintf("Error running command: %s %s. Output: %s", cmdStr, stderr, stdout)
		}
		if failError {
			m.T.Fatalf(errMsg)
		} else {
			m.T.Errorf(errMsg)
		}
	}
	return stdout, stderr
}

// RunCommandRetriable Error  executes a command, returns error
// the purpose of this command is to make it retriable and
// better logging for retrying
func (m *MinikubeRunner) RunCommandRetriable(cmdStr string, waitForRun ...bool) (stdout string, stderr string, err error) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, cmdArgs...)
	Logf("Run: %s", cmd.Args)
	stdout, stderr, err = m.teeRun(cmd, waitForRun...)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			m.T.Logf("temporary error: running command: %s %s. Output: \n%s", cmdStr, exitError.Stderr, stdout)
		} else {
			m.T.Logf("temporary error: running command: %s %s. Output: \n%s", cmdStr, stderr, stdout)
		}
	}
	return stdout, stderr, err
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
func (m *MinikubeRunner) Start(opts ...string) (stdout string, stderr string, err error) {
	cmd := fmt.Sprintf("start %s %s %s", m.StartArgs, m.GlobalArgs, strings.Join(opts, " "))
	s := func() error {
		stdout, stderr, err = m.RunCommandRetriable(cmd)
		return err
	}
	err = retry.Expo(s, 10*time.Second, m.TimeOutStart)
	return stdout, stderr, err
}

// TearDown deletes minikube without waiting for it. used to free up ram/cpu after each test
func (m *MinikubeRunner) TearDown(t *testing.T) {
	profileArg := fmt.Sprintf("-p=%s", m.Profile)
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, profileArg, "delete")
	err := cmd.Start() // don't wait for it to finish
	if err != nil {
		t.Errorf("error tearing down minikube %s : %v", profileArg, err)
	}
}

// EnsureRunning makes sure the container runtime is running
func (m *MinikubeRunner) EnsureRunning(opts ...string) {
	s, _, err := m.Status()
	if err != nil {
		m.T.Errorf("error getting status for ensure running: %v", err)
	}
	if s != state.Running.String() {
		stdout, stderr, err := m.Start(opts...)
		if err != nil {
			m.T.Errorf("error starting while running EnsureRunning : %v , stdout %s stderr %s", err, stdout, stderr)
		}
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

// Status returns the status of a service
func (m *MinikubeRunner) Status() (status string, stderr string, err error) {
	cmd := fmt.Sprintf("status --format={{.Host}} %s", m.GlobalArgs)
	s := func() error {
		status, stderr, err = m.RunCommandRetriable(cmd)
		status = strings.TrimRight(status, "\n")
		return err
	}
	err = retry.Expo(s, 3*time.Second, 2*time.Minute)
	if err != nil && (status == state.None.String() || status == state.Stopped.String()) {
		err = nil // because https://github.com/kubernetes/minikube/issues/4932
	}
	return status, stderr, err
}

// GetLogs returns the logs of a service
func (m *MinikubeRunner) GetLogs() string {
	// TODO: this test needs to check sterr too !
	stdout, _ := m.RunCommand(fmt.Sprintf("logs %s", m.GlobalArgs), true)
	return stdout
}

// CheckStatus makes sure the service has the desired status, or cause fatal error
func (m *MinikubeRunner) CheckStatus(desired string) {
	err := m.CheckStatusNoFail(desired)
	if err != nil { // none status returns 1 exit code
		m.T.Fatalf("%v", err)
	}
}

// CheckStatusNoFail makes sure the service has the desired status, returning error
func (m *MinikubeRunner) CheckStatusNoFail(desired string) error {
	s, stderr, err := m.Status()
	if s != desired {
		return fmt.Errorf("got state: %q, expected %q : stderr: %s err: %v ", s, desired, stderr, err)
	}

	if err != nil {
		return errors.Wrapf(err, stderr)
	}
	return nil
}
