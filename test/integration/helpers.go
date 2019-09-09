/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package integration

// These are test helpers that:
//
// - Accept *testing.T arguments (see helpers.go)
// - Are used in multiple tests
// - Must not compare test values

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/minikube/pkg/kapi"
)

// RunResult stores the result of an cmd.Run call
type RunResult struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
	Args     []string
}

// Command returns a human readable command string that does not induce eye fatigue
func (rr RunResult) Command() string {
	return fmt.Sprintf(`"%s %s"`, strings.TrimPrefix(rr.Args[0], "../../"), strings.Join(rr.Args[1:], " "))
}

func (rr RunResult) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Command: %v\n", rr.Command()))
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n-- stdout -- \n%s\n", rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n** stderr ** \n%s\n", rr.Stderr.Bytes()))
	}
	return sb.String()
}

// RunCmd is a test helper to log a command being executed \_(ツ)_/¯
func RunCmd(ctx context.Context, t *testing.T, name string, arg ...string) (*RunResult, error) {
	t.Helper()

	cmd := exec.CommandContext(ctx, name, arg...)
	rr := &RunResult{Args: cmd.Args}
	if ctx.Err() != nil {
		return rr, fmt.Errorf("test context: %v", ctx.Err())
	}
	t.Logf("Run:    %v", rr.Command())

	var outb, errb bytes.Buffer
	cmd.Stdout, rr.Stdout = &outb, &outb
	cmd.Stderr, rr.Stderr = &errb, &errb
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			t.Logf("Done:   %v: (%s)", rr.Command(), elapsed)
		}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			rr.ExitCode = exitError.ExitCode()
		}
		t.Logf("Non-zero exit: %v: %v (%s)", rr.Command(), err, elapsed)
		t.Logf(rr.String())
	}
	return rr, err
}

// StartSession stores the result of an cmd.Start call
type StartSession struct {
	Stdout *bufio.Reader
	Stderr *bufio.Reader
	cmd    *exec.Cmd
}

// StartCmd starts a process in the background, streaming output
func StartCmd(ctx context.Context, t *testing.T, name string, arg ...string) (*StartSession, error) {
	t.Helper()
	cmd := exec.CommandContext(ctx, name, arg...)
	t.Logf("Daemon: %v", cmd.Args)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe failed: %v %v", cmd.Args, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe failed: %v %v", cmd.Args, err)
	}

	sr := &StartSession{Stdout: bufio.NewReader(stdoutPipe), Stderr: bufio.NewReader(stderrPipe), cmd: cmd}
	return sr, cmd.Start()
}

// Stop stops the started process
func (ss *StartSession) Stop(t *testing.T) error {
	t.Logf("Stopping %s ...", ss.cmd.Args)
	if t.Failed() {
		if ss.Stdout.Size() > 0 {
			stdout, err := ioutil.ReadAll(ss.Stdout)
			if err != nil {
				return fmt.Errorf("read stdout failed: %v", err)
			}
			t.Logf("%s stdout:\n%s", ss.cmd.Args, stdout)
		}
		if ss.Stderr.Size() > 0 {
			stderr, err := ioutil.ReadAll(ss.Stderr)
			if err != nil {
				return fmt.Errorf("read stderr failed: %v", err)
			}
			t.Logf("%s stderr:\n%s", ss.cmd.Args, stderr)
		}
	}
	return ss.cmd.Process.Kill()
}

// Cleanup cleans up after a test run
func Cleanup(t *testing.T, profile string, cancel context.CancelFunc) {
	t.Helper()
	if *cleanup {
		_, err := RunCmd(context.Background(), t, Target(), "delete", "-p", profile)
		if err != nil {
			t.Logf("failed cleanup: %v", err)
		}
	} else {
		t.Logf("Skipping cleanup of %s (--cleanup=false)", profile)
	}
	cancel()
}

// CleanupWithLogs cleans up after a test run, fetching logs and deleting the profile
func CleanupWithLogs(t *testing.T, profile string, cancel context.CancelFunc) {
	t.Helper()
	if t.Failed() {
		t.Logf("%s failed, collecting logs ...", t.Name())
		rr, err := RunCmd(context.Background(), t, Target(), "-p", profile, "logs", "-n", "10")
		if err != nil {
			t.Logf("failed logs error: %v", err)
		}
		t.Logf("%s logs: %s\n", t.Name(), rr)
		t.Logf("Sorry that %s failed :(", t.Name())
	}
	Cleanup(t, profile, cancel)
}

// PodWait waits for pods to achieve a running state.
func PodWait(ctx context.Context, t *testing.T, profile string, ns string, selector string, timeout time.Duration) ([]string, error) {
	t.Helper()
	client, err := kapi.Client(profile)
	if err != nil {
		return nil, err
	}

	// For example: kubernetes.io/minikube-addons=gvisor
	listOpts := meta.ListOptions{LabelSelector: selector}
	minUptime := 5 * time.Second
	podStart := time.Time{}
	foundNames := map[string]bool{}
	lastMsg := ""

	start := time.Now()
	t.Logf("Waiting for pods with labels %q in namespace %q ...", selector, ns)
	f := func() (bool, error) {
		pods, err := client.CoreV1().Pods(ns).List(listOpts)
		if err != nil {
			t.Logf("Pod(%s).List(%v) returned error: %v", ns, selector, err)
			podStart = time.Time{}
			return false, nil
		}
		if len(pods.Items) == 0 {
			podStart = time.Time{}
			return false, nil
		}
		for _, pod := range pods.Items {
			foundNames[pod.ObjectMeta.Name] = true
			// Prevent spamming logs with identical messages
			msg := fmt.Sprintf("%q (%s) %s", pod.ObjectMeta.GetName(), pod.ObjectMeta.GetUID(), pod.Status.Phase)
			if msg != lastMsg {
				t.Logf("%s. Status: %+v", msg, pod.Status)
				lastMsg = msg
			}
			if pod.Status.Phase != core.PodRunning {
				if !podStart.IsZero() {
					t.Logf("WARNING: %s was running %s ago - may be unstable", selector, time.Since(podStart))
				}
				podStart = time.Time{}
				return false, nil
			}

			if podStart.IsZero() {
				podStart = time.Now()
			}
			if time.Since(podStart) > minUptime {
				return true, nil
			}
		}
		return false, nil
	}

	err = wait.PollImmediate(1*time.Second, timeout, f)
	names := []string{}
	for n := range foundNames {
		names = append(names, n)
	}

	if err == nil {
		t.Logf("pods %s up and healthy within %s", selector, time.Since(start))
		return names, nil
	}

	t.Logf("pods %q: %v", selector, err)
	debugFailedPods(ctx, t, profile, ns, names)
	return names, fmt.Errorf("%s: %v", fmt.Sprintf("%s within %s", selector, timeout), err)
}

// debugFailedPods logs debug info for failed pods
func debugFailedPods(ctx context.Context, t *testing.T, profile string, ns string, names []string) {
	rr, rerr := RunCmd(ctx, t, "kubectl", "--context", profile, "get", "po", "-A", "--show-labels")
	if rerr != nil {
		t.Logf("%s: %v", rr.Command(), rerr)
	} else {
		t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
	}

	for _, name := range names {
		rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "describe", "po", name, "-n", ns)
		if err != nil {
			t.Logf("%s: %v", rr.Command(), err)
		} else {
			t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
		}

		rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "logs", name, "-n", ns)
		if err != nil {
			t.Logf("%s: %v", rr.Command(), err)
		} else {
			t.Logf("(debug) %s:\n%s", rr.Command(), rr.Stdout)
		}
	}
}

// Status returns the minikube cluster status as a string
func Status(ctx context.Context, t *testing.T, path string, profile string) string {
	t.Helper()
	rr, err := RunCmd(ctx, t, path, "status", "--format={{.Host}}", "-p", profile)
	if err != nil {
		t.Logf("status error: %v (may be ok)", err)
	}
	return strings.TrimSpace(rr.Stdout.String())
}

// MaybeParallel sets that the test should run in parallel
func MaybeParallel(t *testing.T) {
	t.Helper()
	// TODO: Allow paralellized tests on "none" that do not require independent clusters
	if NoneDriver() {
		return
	}
	t.Parallel()
}
