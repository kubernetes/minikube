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
)

// RunResult stores the result of an cmd.Run call
type RunResult struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
	Args     []string
}

func (rr RunResult) String() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n---- %s stdout ---- \n%s\n", rr.Args, rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("***** %s stderr ***** \n%s\n", rr.Args, rr.Stderr.Bytes()))
	}
	return sb.String()
}

// RunCmd is a test helper to log a command being executed \_(ツ)_/¯
func RunCmd(ctx context.Context, t *testing.T, name string, arg ...string) (*RunResult, error) {
	t.Helper()
	cmd := exec.CommandContext(ctx, name, arg...)
	t.Logf("Run:    %v", cmd.Args)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	rr := &RunResult{Stdout: &outb, Stderr: &errb, Args: cmd.Args}

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			t.Logf("Done:   %v: (%s)", cmd.Args, elapsed)
		}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			rr.ExitCode = exitError.ExitCode()
		}
		t.Logf("Non-zero exit: %v: %v (%s)", cmd.Args, err, elapsed)
		t.Logf("Output: %v: %s", cmd.Args, rr)
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
		t.Logf("skipping cleanuprofile, because --cleanup=false")
	}
	cancel()
}

// CleanupWithLogs cleans up after a test run, fetching logs and deleting the profile
func CleanupWithLogs(t *testing.T, profile string, cancel context.CancelFunc) {
	t.Helper()
	if t.Failed() {
		rr, err := RunCmd(context.Background(), t, Target(), "logs", "-p", profile)
		if err != nil {
			t.Logf("failed logs error: %v", err)
		}
		t.Logf("%s logs: %s\n", t.Name(), rr)
	}
	Cleanup(t, profile, cancel)
}

// Status returns the status as a string
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
	// TODO: Allow paralellized tests that don't require independent clusters
	if NoneDriver() {
		return
	}
	t.Helper()
	t.Logf("Setting %s to run in parallel", t.Name())
	t.Parallel()
}
