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
	"path"
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
	_, err := m.SSH(fmt.Sprintf("rm -rf %s", path.Join(f.GetTargetDir(), f.GetTargetName())))
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

// MustRun executes a command and fails if error, and and unless waitForRun is set to false it waits for it finish.
func (m *MinikubeRunner) MustRun(cmdStr string, waitForRun ...bool) (string, string) {
	stdout, stderr, err := m.RunCommand(cmdStr, true, waitForRun...)
	if err != nil {
		m.T.Logf("MusRun error: %v", err)
	}
	return stdout, stderr
}

// RunCommand executes a command, optionally checking for error and by default waits for run to finish
func (m *MinikubeRunner) RunCommand(cmdStr string, failError bool, waitForRun ...bool) (string, string, error) {
	profileArg := fmt.Sprintf("-p=%s ", m.Profile)
	cmdStr = profileArg + cmdStr
	cmdArgs := strings.Split(cmdStr, " ")
	path, _ := filepath.Abs(m.BinaryPath)

	cmd := exec.Command(path, cmdArgs...)
	Logf("Run: %s", cmd.Args)
	stdout, stderr, err := m.teeRun(cmd, waitForRun...)
	if err != nil {
		exitCode := ""
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = string(exitError.Stderr)
		}
		errMsg := fmt.Sprintf("Error RunCommand : %s \n\t Begin RunCommand log block ---> \n\t With Profile: %s \n\t With ExitCode: %q \n\t With STDOUT %s \n\t With STDERR %s \n\t <--- End of RunCommand log block", cmdStr, m.Profile, exitCode, stdout, stderr)
		if failError {
			m.T.Fatalf(errMsg)
		} else {
			m.T.Logf(errMsg)
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
	Logf("RunWithContext: %s", cmd.Args)
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
func (m *MinikubeRunner) start(opts ...string) (stdout string, stderr string, err error) {
	cmd := fmt.Sprintf("start %s %s %s", m.StartArgs, m.GlobalArgs, strings.Join(opts, " "))
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, m.TimeOutStart)
	defer cancel()
	return m.RunWithContext(ctx, cmd, true)
}

// MustStart starts the cluster and fail the test if error
func (m *MinikubeRunner) MustStart(opts ...string) (stdout string, stderr string) {
	stdout, stderr, err := m.start(opts...)
	// the reason for this formatting is, the logs are very big but useful and also in parallel testing logs are harder to identify
	if err != nil {
		m.T.Fatalf("%s Failed to start minikube With error: %v \n\t begin Start log block ------------> \n\t With Profile: %s \n\t With Args: %v \n\t With Global Args: %s  \n\t With Driver Args: %s \n\t With STDOUT: \n \t %s \n\t With STDERR: \n \t %s \n\t <------------ End of Start (%s) log block", m.T.Name(), err, m.Profile, strings.Join(opts, " "), m.GlobalArgs, m.StartArgs, stdout, stderr, m.Profile)
	}
	return stdout, stderr
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
		stdout, stderr, err := m.start(opts...)
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
	s := func() error {
		status, stderr, err = m.RunCommand("status --format={{.Host}} %s", false)
		status = strings.TrimRight(status, "\n")
		if err != nil && (status == state.None.String() || status == state.Stopped.String()) {
			err = nil // because https://github.com/kubernetes/minikube/issues/4932
		}
		return err
	}
	err = retry.Expo(s, 3*time.Second, 2*time.Minute)
	return status, stderr, err
}

// GetLogs returns the logs of a service
func (m *MinikubeRunner) GetLogs() (string, string) {
	stdout, stderr, err := m.RunCommand(fmt.Sprintf("logs %s", m.GlobalArgs), true)
	if err != nil {
		m.T.Logf("Error in GetLogs %v", err)
	}
	return stdout, stderr
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
