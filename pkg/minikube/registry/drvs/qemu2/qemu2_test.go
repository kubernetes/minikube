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

package qemu2

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
)

const qemuCopyright = "Copyright (c) 2003-2023 Fabrice Bellard and the QEMU Project developers\n"

func TestParseQemuVersion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "plain release",
			output: "QEMU emulator version 8.2.1\n" + qemuCopyright,
			want:   "8.2.1",
		},
		{
			name:   "self-compiled build with git tag suffix (issue #15104)",
			output: "QEMU emulator version 8.2.1 (v8.2.1)\n" + qemuCopyright,
			want:   "8.2.1",
		},
		// QEMU appends " (<pkgversion>)" to the version, where pkgversion is
		// --with-pkgversion or `git describe --match 'v*' --dirty` for builds
		// from a git checkout (scripts/qemu-version.sh). Tarball and Homebrew
		// builds print no suffix.
		{
			name:   "11.x tarball build",
			output: "QEMU emulator version 11.0.5\n" + qemuCopyright,
			want:   "11.0.5",
		},
		{
			name:   "11.x git tag build",
			output: "QEMU emulator version 11.0.5 (v11.0.5)\n" + qemuCopyright,
			want:   "11.0.5",
		},
		{
			name:   "11.x git build ahead of tag and dirty",
			output: "QEMU emulator version 11.0.5 (v11.0.5-3-g1a2b3c4d5e-dirty)\n" + qemuCopyright,
			want:   "11.0.5",
		},
		{
			name:   "custom pkgversion",
			output: "QEMU emulator version 11.0.5 (my build 2026-11-01)\n" + qemuCopyright,
			want:   "11.0.5",
		},
		{
			name:   "7.x git tag build",
			output: "QEMU emulator version 7.1.0 (v7.1.0)\n" + qemuCopyright,
			want:   "7.1.0",
		},
		{
			name:   "7.x git build dirty",
			output: "QEMU emulator version 7.2.0 (v7.2.0-dirty)\n" + qemuCopyright,
			want:   "7.2.0",
		},
		{
			name:   "dirty git build",
			output: "QEMU emulator version 7.1.50 (v7.1.0-1234-gdeadbeef-dirty)\n" + qemuCopyright,
			want:   "7.1.50",
		},
		{
			name:   "debian package suffix",
			output: "QEMU emulator version 7.2.0 (Debian 1:7.2+dfsg-7)\n" + qemuCopyright,
			want:   "7.2.0",
		},
		{
			name:   "fedora package suffix",
			output: "QEMU emulator version 6.2.0 (qemu-6.2.0-11.fc36)\n" + qemuCopyright,
			want:   "6.2.0",
		},
		{
			name:   "ubuntu suffix without separating space",
			output: "QEMU emulator version 2.11.1(Debian 1:2.11+dfsg-1ubuntu7.42)\n" + qemuCopyright,
			want:   "2.11.1",
		},
		{
			name:   "release candidate",
			output: "QEMU emulator version 8.1.90 (v8.2.0-rc0)\n" + qemuCopyright,
			want:   "8.1.90",
		},
		{
			name:   "windows line endings",
			output: "QEMU emulator version 9.0.0 (v9.0.0-12063-g2c1a0ea2a5)\r\nCopyright (c) 2003-2024 Fabrice Bellard\r\n",
			want:   "9.0.0",
		},
		{
			name:   "missing patch component defaults to zero",
			output: "QEMU emulator version 9.1\n",
			want:   "9.1.0",
		},
		{
			name:   "leading whitespace",
			output: "\n  QEMU emulator version 10.0.2\n",
			want:   "10.0.2",
		},
		{
			name:   "multi-digit components",
			output: "QEMU emulator version 10.12.103\n",
			want:   "10.12.103",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseQemuVersion(tc.output)
			if err != nil {
				t.Fatalf("parseQemuVersion(%q) returned unexpected error: %v", tc.output, err)
			}
			want := semver.MustParse(tc.want)
			if !got.Equals(want) {
				t.Errorf("parseQemuVersion(%q) = %s, want %s", tc.output, got, want)
			}
		})
	}
}

func TestParseQemuVersionInvalid(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{name: "empty output", output: ""},
		{name: "whitespace only", output: " \n\t\n"},
		{name: "unrelated program", output: "Python 3.12.1\n"},
		{name: "missing version number", output: "QEMU emulator version\n" + qemuCopyright},
		{name: "major version only", output: "QEMU emulator version 8\n"},
		{name: "non-numeric version", output: "QEMU emulator version x.y.z\n"},
		{name: "version not on the expected line", output: qemuCopyright + "8.2.1\n"},
		{name: "component overflows uint64", output: "QEMU emulator version 99999999999999999999.0.0\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseQemuVersion(tc.output)
			if err == nil {
				t.Fatalf("parseQemuVersion(%q) = %s, want error", tc.output, got)
			}
		})
	}
}

func TestNeedsHighmemOff(t *testing.T) {
	tests := []struct {
		version  string
		memoryMB int
		want     bool
	}{
		{version: "6.2.0", memoryMB: 8192, want: true},
		{version: "6.99.99", memoryMB: 8192, want: true},
		{version: "7.0.0", memoryMB: 8192, want: false},
		{version: "8.2.1", memoryMB: 8192, want: false},
		{version: "11.0.5", memoryMB: 8192, want: false},
		{version: "8.2.1", memoryMB: 3073, want: false},
		{version: "8.2.1", memoryMB: 3072, want: true},
		{version: "8.2.1", memoryMB: 2048, want: true},
		{version: "6.2.0", memoryMB: 2048, want: true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s/%dMB", tc.version, tc.memoryMB), func(t *testing.T) {
			if got := needsHighmemOff(semver.MustParse(tc.version), tc.memoryMB); got != tc.want {
				t.Errorf("needsHighmemOff(%s, %d) = %t, want %t", tc.version, tc.memoryMB, got, tc.want)
			}
		})
	}
}

// TestNeedsHighmemOffFromOutput checks the parsed version of the output from
// issue #15104 end to end through the highmem decision.
func TestNeedsHighmemOffFromOutput(t *testing.T) {
	v, err := parseQemuVersion("QEMU emulator version 8.2.1 (v8.2.1)\n" + qemuCopyright)
	if err != nil {
		t.Fatalf("parseQemuVersion returned unexpected error: %v", err)
	}
	if needsHighmemOff(v, 20480) {
		t.Errorf("needsHighmemOff(%s, 20480) = true, want false", v)
	}
}

func TestQemuSystemProgram(t *testing.T) {
	want := map[string]string{
		"amd64": "qemu-system-x86_64",
		"arm64": "qemu-system-aarch64",
	}
	got, err := qemuSystemProgram()
	expected, supported := want[runtime.GOARCH]
	if !supported {
		if err == nil {
			t.Fatalf("qemuSystemProgram() = %q on unsupported arch %s, want error", got, runtime.GOARCH)
		}
		return
	}
	if err != nil {
		t.Fatalf("qemuSystemProgram() returned unexpected error: %v", err)
	}
	if got != expected {
		t.Errorf("qemuSystemProgram() = %q, want %q", got, expected)
	}
}

// installFakeQemu puts a shell script named like the qemu system binary for
// the current arch in a temporary directory and makes it the only PATH entry.
func installFakeQemu(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake qemu binary uses a shell script")
	}
	prog, err := qemuSystemProgram()
	if err != nil {
		t.Skipf("unsupported arch: %v", err)
	}
	dir := t.TempDir()
	if script != "" {
		path := filepath.Join(dir, prog)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
			t.Fatalf("writing fake qemu: %v", err)
		}
	}
	t.Setenv("PATH", dir)
	return prog
}

func TestQemuVersion(t *testing.T) {
	installFakeQemu(t, `if [ "$1" != "-version" ]; then exit 2; fi
printf 'QEMU emulator version 11.0.5 (v11.0.5-dirty)\nCopyright (c) 2003-2026 Fabrice Bellard and the QEMU Project developers\n'
`)
	got, err := qemuVersion()
	if err != nil {
		t.Fatalf("qemuVersion() returned unexpected error: %v", err)
	}
	if want := semver.MustParse("11.0.5"); !got.Equals(want) {
		t.Errorf("qemuVersion() = %s, want %s", got, want)
	}
}

func TestQemuVersionErrors(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name:    "binary not found",
			script:  "",
			wantErr: "-version:",
		},
		{
			name:    "binary exits with error",
			script:  "echo 'boom' >&2\nexit 1\n",
			wantErr: "-version:",
		},
		{
			name:    "unparsable output",
			script:  "echo 'not qemu'\n",
			wantErr: `unable to parse QEMU version from "not qemu"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog := installFakeQemu(t, tc.script)
			got, err := qemuVersion()
			if err == nil {
				t.Fatalf("qemuVersion() = %s, want error", got)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("qemuVersion() error = %q, want it to contain %q", err, tc.wantErr)
			}
			if tc.wantErr == "-version:" && !strings.HasPrefix(err.Error(), prog+" -version:") {
				t.Errorf("qemuVersion() error = %q, want prefix %q", err, prog+" -version:")
			}
		})
	}
}
