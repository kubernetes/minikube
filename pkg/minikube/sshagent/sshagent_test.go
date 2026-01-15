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
	"testing"
)

// TestParseOutput verifies that the ssh-agent output parsing logic correctly extracts
// the socket path and PID. This is critical because minikube relies on these environment
// variables to connect to the ssh-agent for key management.
func TestParseOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantSock  string
		wantPID   int
		wantErr   bool
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
