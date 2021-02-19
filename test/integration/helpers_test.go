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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/shirou/gopsutil/v3/process"
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
	var sb strings.Builder
	sb.WriteString(strings.TrimPrefix(rr.Args[0], "../../"))
	for _, a := range rr.Args[1:] {
		if strings.Contains(a, " ") {
			sb.WriteString(fmt.Sprintf(` "%s"`, a))
			continue
		}
		sb.WriteString(fmt.Sprintf(" %s", a))
	}
	return sb.String()
}

// indentLines indents every line in a bytes.Buffer and returns it as string
func indentLines(b []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(b))
	var lines string
	for scanner.Scan() {
		lines = lines + "\t" + scanner.Text() + "\n"
	}
	return lines
}

// Output returns human-readable output for an execution result
func (rr RunResult) Output() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n-- stdout --\n%s\n-- /stdout --", indentLines(rr.Stdout.Bytes())))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n** stderr ** \n%s\n** /stderr **", indentLines(rr.Stderr.Bytes())))
	}
	return sb.String()
}

// Run is a test helper to log a command being executed \_(ツ)_/¯
func Run(t *testing.T, cmd *exec.Cmd) (*RunResult, error) {
	t.Helper()
	rr := &RunResult{Args: cmd.Args}
	t.Logf("(dbg) Run:  %v", rr.Command())

	var outb, errb bytes.Buffer
	cmd.Stdout, rr.Stdout = &outb, &outb
	cmd.Stderr, rr.Stderr = &errb, &errb
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			t.Logf("(dbg) Done: %v: (%s)", rr.Command(), elapsed)
		}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			rr.ExitCode = exitError.ExitCode()
		}
		t.Logf("(dbg) Non-zero exit: %v: %v (%s)\n%s", rr.Command(), err, elapsed, rr.Output())
	}
	return rr, err
}

// StartSession stores the result of an cmd.Start call
type StartSession struct {
	Stdout *bufio.Reader
	Stderr *bufio.Reader
	cmd    *exec.Cmd
}

// Start starts a process in the background, streaming output
func Start(t *testing.T, cmd *exec.Cmd) (*StartSession, error) {
	t.Helper()
	t.Logf("(dbg) daemon: %v", cmd.Args)

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
func (ss *StartSession) Stop(t *testing.T) {
	t.Helper()
	t.Logf("(dbg) stopping %s ...", ss.cmd.Args)
	if ss.cmd.Process == nil {
		t.Logf("%s has a nil Process. Maybe it's dead? How weird!", ss.cmd.Args)
		return
	}
	killProcessFamily(t, ss.cmd.Process.Pid)
	if t.Failed() {
		if ss.Stdout.Size() > 0 {
			stdout, err := ioutil.ReadAll(ss.Stdout)
			if err != nil {
				t.Logf("read stdout failed: %v", err)
			}
			t.Logf("(dbg) %s stdout:\n%s", ss.cmd.Args, stdout)
		}
		if ss.Stderr.Size() > 0 {
			stderr, err := ioutil.ReadAll(ss.Stderr)
			if err != nil {
				t.Logf("read stderr failed: %v", err)
			}
			t.Logf("(dbg) %s stderr:\n%s", ss.cmd.Args, stderr)
		}
	}
}

// Cleanup cleans up after a test run
func Cleanup(t *testing.T, profile string, cancel context.CancelFunc) {
	// No helper because it makes the call log confusing.
	if *cleanup {
		t.Logf("Cleaning up %q profile ...", profile)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		_, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
		if err != nil {
			t.Logf("failed cleanup: %v", err)
		}
	} else {
		t.Logf("skipping cleanup of %s (--cleanup=false)", profile)
	}
	cancel()
}

// CleanupWithLogs cleans up after a test run, fetching logs and deleting the profile
func CleanupWithLogs(t *testing.T, profile string, cancel context.CancelFunc) {
	t.Helper()
	if !t.Failed() {
		Cleanup(t, profile, cancel)
		return
	}

	t.Logf("*** %s FAILED at %s", t.Name(), time.Now())
	PostMortemLogs(t, profile)
	Cleanup(t, profile, cancel)
}

// PostMortemLogs shows logs for debugging a failed cluster
func PostMortemLogs(t *testing.T, profile string, multinode ...bool) {
	if !t.Failed() {
		return
	}

	if !*postMortemLogs {
		t.Logf("post-mortem logs disabled, oh well!")
		return
	}

	m := false
	if len(multinode) > 0 {
		m = multinode[0]
	}

	nodes := []string{profile}
	if m {
		nodes = append(nodes, SecondNodeName, ThirdNodeName)
	}

	t.Logf("-----------------------post-mortem--------------------------------")

	for _, n := range nodes {
		machine := profile
		if n != profile {
			machine = fmt.Sprintf("%s-%s", profile, n)
		}
		if DockerDriver() {
			t.Logf("======>  post-mortem[%s]: docker inspect <======", t.Name())
			rr, err := Run(t, exec.Command("docker", "inspect", machine))
			if err != nil {
				t.Logf("failed to get docker inspect: %v", err)
			} else {
				t.Logf("(dbg) %s:\n%s", rr.Command(), rr.Output())
			}
		}

		st := Status(context.Background(), t, Target(), profile, "Host", n)
		if st != state.Running.String() {
			t.Logf("%q host is not running, skipping log retrieval (state=%q)", profile, st)
			return
		}
		t.Logf("<<< %s FAILED: start of post-mortem logs <<<", t.Name())
		t.Logf("======>  post-mortem[%s]: minikube logs <======", t.Name())

		rr, err := Run(t, exec.Command(Target(), "-p", profile, "logs", "-n", "25"))
		if err != nil {
			t.Logf("failed logs error: %v", err)
			return
		}
		t.Logf("%s logs: %s", t.Name(), rr.Output())

		st = Status(context.Background(), t, Target(), profile, "APIServer", n)
		if st != state.Running.String() {
			t.Logf("%q apiserver is not running, skipping kubectl commands (state=%q)", profile, st)
			return
		}

		// Get non-running pods. NOTE: This does not yet contain pods which are "running", but not "ready"
		rr, rerr := Run(t, exec.Command("kubectl", "--context", profile, "get", "po", "-o=jsonpath={.items[*].metadata.name}", "-A", "--field-selector=status.phase!=Running"))
		if rerr != nil {
			t.Logf("%s: %v", rr.Command(), rerr)
			return
		}
		notRunning := strings.Split(rr.Stdout.String(), " ")
		t.Logf("non-running pods: %s", strings.Join(notRunning, " "))

		t.Logf("======> post-mortem[%s]: describe non-running pods <======", t.Name())

		args := append([]string{"--context", profile, "describe", "pod"}, notRunning...)
		rr, rerr = Run(t, exec.Command("kubectl", args...))
		if rerr != nil {
			t.Logf("%s: %v", rr.Command(), rerr)
			return
		}
		t.Logf("(dbg) %s:\n%s", rr.Command(), rr.Output())
	}

	t.Logf("<<< %s FAILED: end of post-mortem logs <<<", t.Name())
	t.Logf("---------------------/post-mortem---------------------------------")
}

// podStatusMsg returns a human-readable pod status, for generating debug status
func podStatusMsg(pod core.Pod) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%q [%s] %s", pod.ObjectMeta.GetName(), pod.ObjectMeta.GetUID(), pod.Status.Phase))
	for i, c := range pod.Status.Conditions {
		if c.Reason != "" {
			if i == 0 {
				sb.WriteString(": ")
			} else {
				sb.WriteString(" / ")
			}
			sb.WriteString(fmt.Sprintf("%s:%s", c.Type, c.Reason))
		}
		if c.Message != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", c.Message))
		}
	}
	return sb.String()
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
	t.Logf("(dbg) %s: waiting %s for pods matching %q in namespace %q ...", t.Name(), timeout, selector, ns)
	f := func() (bool, error) {
		pods, err := client.CoreV1().Pods(ns).List(listOpts)
		if err != nil {
			t.Logf("%s: WARNING: pod list for %q %q returned: %v", t.Name(), ns, selector, err)
			// Don't return the error upwards so that this is retried, in case the apiserver is rescheduled
			podStart = time.Time{}
			return false, nil
		}
		if len(pods.Items) == 0 {
			podStart = time.Time{}
			return false, nil
		}

		for _, pod := range pods.Items {
			foundNames[pod.ObjectMeta.Name] = true
			msg := podStatusMsg(pod)
			// Prevent spamming logs with identical messages
			if msg != lastMsg {
				t.Log(msg)
				lastMsg = msg
			}
			// Successful termination of a short-lived process, will not be restarted
			if pod.Status.Phase == core.PodSucceeded {
				return true, nil
			}
			// Long-running process state
			if pod.Status.Phase != core.PodRunning {
				if !podStart.IsZero() {
					t.Logf("%s: WARNING: %s was running %s ago - may be unstable", t.Name(), selector, time.Since(podStart))
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
		t.Logf("(dbg) %s: %s healthy within %s", t.Name(), selector, time.Since(start))
		return names, nil
	}

	t.Logf("***** %s: pod %q failed to start within %s: %v ****", t.Name(), selector, timeout, err)
	showPodLogs(ctx, t, profile, ns, names)
	return names, fmt.Errorf("%s: %v", fmt.Sprintf("%s within %s", selector, timeout), err)
}

// PVCWait waits for persistent volume claim to reach bound state
func PVCWait(ctx context.Context, t *testing.T, profile string, ns string, name string, timeout time.Duration) error {
	t.Helper()

	t.Logf("(dbg) %s: waiting %s for pvc %q in namespace %q ...", t.Name(), timeout, name, ns)

	f := func() (bool, error) {
		ret, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "pvc", name, "-o", "jsonpath={.status.phase}", "-n", ns))
		if err != nil {
			t.Logf("%s: WARNING: PVC get for %q %q returned: %v", t.Name(), ns, name, err)
			return false, nil
		}

		pvc := strings.TrimSpace(ret.Stdout.String())
		if pvc == string(core.ClaimBound) {
			return true, nil
		} else if pvc == string(core.ClaimLost) {
			return true, fmt.Errorf("PVC %q is LOST", name)
		}
		return false, nil
	}

	return wait.PollImmediate(1*time.Second, timeout, f)
}

//// VolumeSnapshotWait waits for volume snapshot to be ready to use
func VolumeSnapshotWait(ctx context.Context, t *testing.T, profile string, ns string, name string, timeout time.Duration) error {
	t.Helper()

	t.Logf("(dbg) %s: waiting %s for volume snapshot %q in namespace %q ...", t.Name(), timeout, name, ns)

	f := func() (bool, error) {
		res, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "volumesnapshot", name, "-o", "jsonpath={.status.readyToUse}", "-n", ns))
		if err != nil {
			t.Logf("%s: WARNING: volume snapshot get for %q %q returned: %v", t.Name(), ns, name, err)
			return false, nil
		}

		isReady, err := strconv.ParseBool(strings.TrimSpace(res.Stdout.String()))
		if err != nil {
			t.Logf("%s: WARNING: volume snapshot get for %q %q returned: %v", t.Name(), ns, name, res.Stdout.String())
			return false, nil
		}

		return isReady, nil
	}

	return wait.PollImmediate(1*time.Second, timeout, f)
}

// Status returns a minikube component status as a string
func Status(ctx context.Context, t *testing.T, path string, profile string, key string, node string) string {
	t.Helper()
	// Reminder of useful keys: "Host", "Kubelet", "APIServer"
	rr, err := Run(t, exec.CommandContext(ctx, path, "status", fmt.Sprintf("--format={{.%s}}", key), "-p", profile, "-n", node))
	if err != nil {
		t.Logf("status error: %v (may be ok)", err)
	}
	return strings.TrimSpace(rr.Stdout.String())
}

// showPodLogs logs debug info for pods
func showPodLogs(ctx context.Context, t *testing.T, profile string, ns string, names []string) {
	t.Helper()
	st := Status(context.Background(), t, Target(), profile, "APIServer", profile)
	if st != state.Running.String() {
		t.Logf("%q apiserver is not running, skipping kubectl commands (state=%q)", profile, st)
		return
	}

	t.Logf("%s: showing logs for failed pods as of %s", t.Name(), time.Now())

	for _, name := range names {
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "po", name, "-n", ns))
		if err != nil {
			t.Logf("%s: %v", rr.Command(), err)
		} else {
			t.Logf("(dbg) %s:\n%s", rr.Command(), rr.Stdout)
		}

		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", name, "-n", ns))
		if err != nil {
			t.Logf("%s: %v", rr.Command(), err)
		} else {
			t.Logf("(dbg) %s:\n%s", rr.Command(), rr.Stdout)
		}
	}
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

// killProcessFamily kills a pid and all of its children
func killProcessFamily(t *testing.T, pid int) {
	parent, err := process.NewProcess(int32(pid))
	if err != nil {
		t.Logf("unable to find parent, assuming dead: %v", err)
		return
	}
	procs := []*process.Process{}
	children, err := parent.Children()
	if err == nil {
		procs = append(procs, children...)
	}
	procs = append(procs, parent)

	for _, p := range procs {
		if err := p.Terminate(); err != nil {
			t.Logf("unable to terminate pid %d: %v", p.Pid, err)
			continue
		}
		// Allow process a chance to cleanup before instant death.
		time.Sleep(100 * time.Millisecond)
		if err := p.Kill(); err != nil {
			t.Logf("unable to kill pid %d: %v", p.Pid, err)
			continue
		}
	}
}
