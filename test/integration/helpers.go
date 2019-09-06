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
	"os/exec"
	"strings"
	"testing"
	"time"
)

// RunResult stores the result of an cmd.Run call
type RunResult struct {
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Cmd    *exec.Cmd
}

func (rr RunResult) String() string {
	var sb strings.Builder
	if rr.Stdout.Len() > 0 {
		sb.WriteString(fmt.Sprintf("\n---- %s stdout ---- \n%s\n", rr.Cmd.Args, rr.Stdout.Bytes()))
	}
	if rr.Stderr.Len() > 0 {
		sb.WriteString(fmt.Sprintf("***** %s stderr ***** \n%s\n", rr.Cmd.Args, rr.Stderr.Bytes()))
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
	rr := &RunResult{Stdout: &outb, Stderr: &errb, Cmd: cmd}

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	if err == nil {
		// Reduce log spam
		if elapsed > (1 * time.Second) {
			t.Logf("Done:   %v: (%s)", cmd.Args, elapsed)
		}
	} else {
		t.Logf("Failed: %v: %v (%s)", cmd.Args, err, elapsed)
		t.Logf("Output: %v: %s", cmd.Args, rr)
	}
	return rr, err
}

// JSONCmd is a helper to run a command and decode the JSON output from it. Errors are always fatal.
func JSONCmd(ctx context.Context, t *testing.T, i interface{}, name string, arg ...string) (*RunResult) {
	rr, err := RunCmd(ctx, t, name, arg...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Cmd.Args, err)
	}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(i); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// StreamingResult stores the result of an cmd.Start call
type StreamingResult struct {
	Stdout *bufio.Reader
	Stderr *bufio.Reader
	Cmd    *exec.Cmd
}

// StartCmd is a test helper to start a command and stream output \_(ツ)_/¯
func StartCmd(ctx context.Context, t *testing.T, name string, arg ...string) (*StreamingResult, error) {
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

	sr := &StreamingResult{Stdout: bufio.NewReader(stdoutPipe), Stderr: bufio.NewReader(stderrPipe), Cmd: cmd}
	return sr, cmd.Start()
}

// Cleanup cleans up after a test run
func Cleanup(t *testing.T, profile string, cancel context.CancelFunc) {
	t.Helper()
	if *cleanup {
		_, err := Run(context.Background(), t, Target(), "delete", "-p", profile)
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
		rr, err := Run(context.Background(), t, Target(), "logs", "-p", profile)
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
	// TEMPORARY
	return

	t.Helper()
	if NoneDriver() {
		return
	}
	t.Logf("Setting %s to run in parallel", t.Name())
	t.Parallel()
}
