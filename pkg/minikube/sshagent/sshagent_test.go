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

package sshagent

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/config"
)

// TestParseOutput verifies that the ssh-agent output parsing logic correctly extracts
// the socket path and PID. This is critical because minikube relies on these environment
// variables to connect to the ssh-agent for key management.
func TestParseOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantSock string
		wantPID  int
		wantErr  bool
	}{
		{
			name:     "standard output",
			output:   "SSH_AUTH_SOCK=/tmp/ssh-bar; export SSH_AUTH_SOCK;\nSSH_AGENT_PID=12345; export SSH_AGENT_PID;\necho Agent pid 12345;",
			wantSock: "/tmp/ssh-bar",
			wantPID:  12345,
			wantErr:  false,
		},
		{
			name:    "missing sock",
			output:  "SSH_AGENT_PID=12345; export SSH_AGENT_PID;",
			wantErr: true,
		},
		{
			name:    "missing pid",
			output:  "SSH_AUTH_SOCK=/tmp/ssh-bar; export SSH_AUTH_SOCK;",
			wantErr: true,
		},
		{
			name:    "invalid pid",
			output:  "SSH_AUTH_SOCK=/tmp/ssh-bar; export SSH_AUTH_SOCK;\nSSH_AGENT_PID=foo; export SSH_AGENT_PID;",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.authSock != tt.wantSock {
					t.Errorf("parseOutput() gotSock = %v, want %v", got.authSock, tt.wantSock)
				}
				if got.agentPID != tt.wantPID {
					t.Errorf("parseOutput() gotPID = %v, want %v", got.agentPID, tt.wantPID)
				}
			}
		})
	}
}

// TestIsRunning verifies that isRunning correctly identifies the ssh-agent process
// using the gopsutil library. It checks both positive (process running and named correctly)
// and negative (wrong process, non-existent PID) cases.
func TestIsRunning(t *testing.T) {
	// 1. Prepare a fake binary that waits for a signal
	var waitForSig = []byte(`
package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	<-ch
}
`)
	td := t.TempDir()
	srcFile := filepath.Join(td, "main.go")
	if err := os.WriteFile(srcFile, waitForSig, 0o600); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	// 2. Helper to build binaries
	buildBinary := func(name string) string {
		out := filepath.Join(td, name)
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", out, srcFile)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to build %s: %v\nOutput: %s", name, err, string(out))
		}
		return out
	}

	// 3. Build "fake-ssh-agent" and "other-process"
	sshAgentBin := buildBinary("fake-ssh-agent")
	otherBin := buildBinary("other-process")

	// 4. Start processes
	startProc := func(bin string) *exec.Cmd {
		cmd := exec.Command(bin)
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start %s: %v", bin, err)
		}
		return cmd
	}

	agentCmd := startProc(sshAgentBin)
	otherCmd := startProc(otherBin)

	defer func() {
		_ = agentCmd.Process.Kill()
		_ = otherCmd.Process.Kill()
		_ = agentCmd.Wait()
		_ = otherCmd.Wait()
	}()

	// Give them time to start
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name      string
		pid       int
		want      bool
		wantError bool
	}{
		{"ssh-agent running", agentCmd.Process.Pid, true, false},
		{"wrong process", otherCmd.Process.Pid, false, false},
		{"pid 0", 0, false, false},
		{"non-existent pid", 999999999, false, false}, // Assuming this PID doesn't exist
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cc := &config.ClusterConfig{
				SSHAgentPID: tc.pid,
			}
			got, err := isRunning(cc)
			if (err != nil) != tc.wantError {
				t.Errorf("isRunning() error = %v, wantError %v", err, tc.wantError)
				return
			}
			if got != tc.want {
				t.Errorf("isRunning() = %v, want %v", got, tc.want)
			}
		})
	}
}
