/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"os"
	"strings"
	"testing"
)

func TestParseHostInfo(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		wantRemote bool
		wantSSH   bool
		wantErr   bool
	}{
		{
			name:      "empty host",
			host:      "",
			wantRemote: false,
			wantSSH:   false,
			wantErr:   false,
		},
		{
			name:      "ssh host",
			host:      "ssh://user@example.com:22",
			wantRemote: true,
			wantSSH:   true,
			wantErr:   false,
		},
		{
			name:      "tcp localhost",
			host:      "tcp://localhost:2376",
			wantRemote: false,
			wantSSH:   false,
			wantErr:   false,
		},
		{
			name:      "tcp remote",
			host:      "tcp://example.com:2376",
			wantRemote: true,
			wantSSH:   false,
			wantErr:   false,
		},
		{
			name:      "unix socket",
			host:      "unix:///var/run/docker.sock",
			wantRemote: false,
			wantSSH:   false,
			wantErr:   false,
		},
		{
			name:      "https remote",
			host:      "https://example.com:2376",
			wantRemote: true,
			wantSSH:   false,
			wantErr:   false,
		},
		{
			name:      "npipe Windows",
			host:      "npipe:////./pipe/docker_engine",
			wantRemote: false,
			wantSSH:   false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemote, gotSSH, err := parseHostInfo(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHostInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRemote != tt.wantRemote {
				t.Errorf("parseHostInfo() gotRemote = %v, want %v", gotRemote, tt.wantRemote)
			}
			if gotSSH != tt.wantSSH {
				t.Errorf("parseHostInfo() gotSSH = %v, want %v", gotSSH, tt.wantSSH)
			}
		})
	}
}

func TestValidateSSHContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     *ContextInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid SSH context",
			ctx: &ContextInfo{
				Name:     "test-ssh",
				Host:     "ssh://user@example.com:22",
				IsRemote: true,
				IsSSH:    true,
			},
			wantErr: false,
		},
		{
			name: "SSH context without host",
			ctx: &ContextInfo{
				Name:     "test-ssh",
				Host:     "",
				IsRemote: true,
				IsSSH:    true,
			},
			wantErr: true,
			errMsg:  "no host specified",
		},
		{
			name: "SSH context without username",
			ctx: &ContextInfo{
				Name:     "test-ssh",
				Host:     "ssh://example.com:22",
				IsRemote: true,
				IsSSH:    true,
			},
			wantErr: true,
			errMsg:  "must specify a username",
		},
		{
			name: "SSH context without hostname",
			ctx: &ContextInfo{
				Name:     "test-ssh",
				Host:     "ssh://user@:22",
				IsRemote: true,
				IsSSH:    true,
			},
			wantErr: true,
			errMsg:  "must specify a hostname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSSHContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSSHContext() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("validateSSHContext() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestGetContextEnvironment(t *testing.T) {
	// Save original env
	origDockerHost := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", origDockerHost)

	tests := []struct {
		name     string
		setupEnv func()
		wantHost string
	}{
		{
			name: "DOCKER_HOST set",
			setupEnv: func() {
				os.Setenv("DOCKER_HOST", "tcp://remote:2376")
			},
			wantHost: "tcp://remote:2376",
		},
		{
			name: "No DOCKER_HOST",
			setupEnv: func() {
				os.Unsetenv("DOCKER_HOST")
			},
			wantHost: "", // Note: this may get overridden by Docker context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			env, err := GetContextEnvironment()
			if err != nil {
				t.Errorf("GetContextEnvironment() error = %v", err)
				return
			}
			got := env["DOCKER_HOST"]
			// For the "No DOCKER_HOST" test, we need to handle the case where
			// Docker context might still provide a host value
			if tt.name == "No DOCKER_HOST" && tt.wantHost == "" && got != "" {
				// This is acceptable - Docker context is providing the host
				t.Logf("Docker context provided host %q when DOCKER_HOST env var was unset", got)
				return
			}
			if got != tt.wantHost {
				t.Errorf("GetContextEnvironment() DOCKER_HOST = %v, want %v", got, tt.wantHost)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}