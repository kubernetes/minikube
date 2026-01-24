//go:build darwin

/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package vmnet

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestMacOSVersion(t *testing.T) {
	version, err := macOSVersion()
	if err != nil {
		t.Fatalf("macOSVersion() failed: %v", err)
	}

	// Verify against sw_vers output.
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		t.Fatalf("sw_vers failed: %v", err)
	}
	expected := strings.TrimSpace(string(out))

	if version != expected {
		t.Errorf("macOSVersion() = %s, sw_vers reports %s", version, expected)
	}
}

func TestHelperNeedsSudo(t *testing.T) {
	tests := []struct {
		name          string
		helperVersion helperVersion
		macOSVersion  string
		want          bool
	}{
		{
			name:          "old macOS new helper",
			helperVersion: helperVersion{Version: "v0.9.0"},
			macOSVersion:  "15.0",
			want:          true,
		},
		{
			name:          "macOS 26 old helper",
			helperVersion: helperVersion{Version: "v0.8.0"},
			macOSVersion:  "26.0",
			want:          true,
		},
		{
			name:          "macOS 26 new helper",
			helperVersion: helperVersion{Version: "v0.9.0"},
			macOSVersion:  "26.0",
			want:          false,
		},
		{
			name:          "macOS 26 newer helper",
			helperVersion: helperVersion{Version: "v1.0.0"},
			macOSVersion:  "26.0",
			want:          false,
		},
		{
			name:          "macOS 27 new helper",
			helperVersion: helperVersion{Version: "v0.9.0"},
			macOSVersion:  "27.0",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := helperNeedsSudo(tt.helperVersion, tt.macOSVersion)
			if err != nil {
				t.Fatalf("helperNeedsSudo() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("helperNeedsSudo(%v, %s) = %v, want %v",
					tt.helperVersion.Version, tt.macOSVersion, got, tt.want)
			}
		})
	}
}

func TestHelperVersion(t *testing.T) {
	tests := []struct {
		name   string
		stdout string
		want   helperVersion
	}{
		{
			name:   "current version",
			stdout: "version: v0.9.0\ncommit: b76a29eb542d3ce4df18c4ebef6b1498174a02e5\n",
			want:   helperVersion{Version: "v0.9.0", Commit: "b76a29eb542d3ce4df18c4ebef6b1498174a02e5"},
		},
		{
			name:   "legacy version",
			stdout: "v0.6.0\n",
			want:   helperVersion{Version: "v0.6.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHelperVersion([]byte(tt.stdout))
			if err != nil {
				t.Fatalf("parseHelperVersion() failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseHelperVersion() = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("error", func(t *testing.T) {
		stdout := []byte("")
		_, err := parseHelperVersion(stdout)
		if err == nil {
			t.Fatalf("parseHelperVersion(%q) expected error, got nil", stdout)
		}
		// Error should include the quoted input for debugging.
		quoted := fmt.Sprintf("%q", stdout)
		if !strings.Contains(err.Error(), quoted) {
			t.Errorf("error %q should contain quoted input %s", err, quoted)
		}
	})
}

func TestExecutableName(t *testing.T) {
	tests := []struct {
		name      string
		needsSudo bool
		want      string
	}{
		{
			name:      "needs sudo",
			needsSudo: true,
			want:      "sudo",
		},
		{
			name:      "no sudo",
			needsSudo: false,
			want:      "vmnet-helper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Helper{NeedsSudo: tt.needsSudo}
			got := h.executableName()
			if got != tt.want {
				t.Errorf("executableName() = %q, want %q", got, tt.want)
			}
		})
	}
}
