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

package process

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var waitTimeout = 250 * time.Millisecond

func TestPidfile(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pid")
	if err := WritePidfile(pidfile, 42); err != nil {
		t.Fatal(err)
	}
	pid, err := ReadPidfile(pidfile)
	if err != nil {
		t.Fatal(err)
	}
	if pid != 42 {
		t.Fatalf("expected 42, got %d", pid)
	}
}

func TestPidfileMissing(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pid")
	_, err := ReadPidfile(pidfile)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestPidfileInvalid(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "pid")
	if err := os.WriteFile(pidfile, []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ReadPidfile(pidfile)
	if err == nil {
		t.Fatal("parsing invalid pidfile did not fail")
	}
}

func TestProcess(t *testing.T) {
	sleep, err := build(t, "sleep")
	if err != nil {
		t.Fatal(err)
	}

	noterm, err := build(t, "noterm")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("exists", func(t *testing.T) {
		cmd := startProcess(t, sleep)
		exists, err := Exists(cmd.Process.Pid, filepath.Base(sleep))
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("existing process not found")
		}
		exists, err = Exists(0xfffffffc, filepath.Base(sleep))
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("process with non-existing pid found")
		}
		exists, err = Exists(cmd.Process.Pid, "no-such-executable")
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("process with non-existing executable found")
		}
	})

	t.Run("terminate", func(t *testing.T) {
		cmd := startProcess(t, sleep)
		err := Terminate(cmd.Process.Pid, filepath.Base(sleep))
		if err != nil {
			t.Fatal(err)
		}
		waitForTermination(t, cmd)
		exists, err := Exists(cmd.Process.Pid, filepath.Base(sleep))
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("reaped process exists")
		}
	})

	t.Run("terminate name mismatch", func(t *testing.T) {
		cmd := startProcess(t, sleep)
		err := Terminate(cmd.Process.Pid, "no-such-process")
		if err == nil {
			t.Fatalf("Signaled unrelated process")
		}
		if err != os.ErrProcessDone {
			t.Fatalf("Unexpected error: %s", err)
		}
	})

	t.Run("terminate ignored", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("no way to ignore termination on windows")
		}
		cmd := startProcess(t, noterm)
		err := Terminate(cmd.Process.Pid, filepath.Base(noterm))
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(waitTimeout)
		exists, err := Exists(cmd.Process.Pid, filepath.Base(noterm))
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("process terminated")
		}
	})

	t.Run("kill", func(t *testing.T) {
		cmd := startProcess(t, noterm)
		err := Kill(cmd.Process.Pid, filepath.Base(noterm))
		if err != nil {
			t.Fatal(err)
		}
		waitForTermination(t, cmd)
		exists, err := Exists(cmd.Process.Pid, filepath.Base(noterm))
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("reaped process exists")
		}
	})

	t.Run("kill name mismatch", func(t *testing.T) {
		cmd := startProcess(t, noterm)
		err := Kill(cmd.Process.Pid, "no-such-process")
		if err == nil {
			t.Fatalf("Killed unrelated process")
		}
		if err != os.ErrProcessDone {
			t.Fatalf("Unexpected error: %s", err)
		}
	})
}

func startProcess(t *testing.T, cmd string) *exec.Cmd {
	name := filepath.Base(cmd)
	c := exec.Command(cmd)
	stdout, err := c.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Logf("Started process %q (pid=%v)", name, c.Process.Pid)

	t.Cleanup(func() {
		_ = c.Process.Kill()
		_ = c.Wait()
	})

	// Synchronize with the process to ensure it set up signal handlers before
	// we send a signal.
	r := bufio.NewReader(stdout)
	line, err := r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != "READY\n" {
		t.Fatalf("Unexpected response: %q", line)
	}
	t.Logf("Process %q ready in %.6f seconds", name, time.Since(start).Seconds())

	return c
}

func waitForTermination(t *testing.T, cmd *exec.Cmd) {
	name := filepath.Base(cmd.Path)
	timer := time.AfterFunc(waitTimeout, func() {
		t.Fatalf("Timeout waiting for %q", name)
	})
	defer timer.Stop()
	start := time.Now()
	err := cmd.Wait()
	t.Logf("Process %q terminated in %.6f seconds: %s", name, time.Since(start).Seconds(), err)
}

func build(t *testing.T, name string) (string, error) {
	source := fmt.Sprintf("testdata/%s.go", name)

	out := filepath.Join(t.TempDir(), name)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	t.Logf("Building %q", name)
	cmd := exec.Command("go", "build", "-o", out, source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out, nil
}
