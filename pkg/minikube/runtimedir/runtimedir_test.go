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

package runtimedir

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// linuxSunPathLen is the size of sockaddr_un.sun_path on Linux (glibc).
const linuxSunPathLen = 108

// darwinSunPathLen is the size of sockaddr_un.sun_path on macOS/BSD.
const darwinSunPathLen = 104

// testUID is a high uid used to keep /tmp/<uid>/minikube test artifacts out
// of the way of any real user. removeAllRemaining cleans the whole subtree
// at the end of each subtest that touches it.
const testUID = 999999

func removeAllRemaining(t *testing.T, goos string) {
	t.Helper()
	if goos == "darwin" || goos == "linux" {
		_ = os.RemoveAll(filepath.Join("/tmp", strconv.Itoa(testUID)))
	}
}

func TestBaseDir(t *testing.T) {
	xdgPresent := t.TempDir()
	winTmp := t.TempDir()

	cases := []struct {
		name   string
		goos   string
		uid    int
		xdg    string
		winTmp string
		want   string
	}{
		{
			name:   "darwin always uses tmp",
			goos:   "darwin",
			uid:    501,
			xdg:    xdgPresent,
			winTmp: winTmp,
			want:   "/tmp/501/minikube",
		},
		{
			name:   "linux uses xdg when dir exists",
			goos:   "linux",
			uid:    1000,
			xdg:    xdgPresent,
			winTmp: winTmp,
			want:   filepath.Join(xdgPresent, "minikube"),
		},
		{
			name:   "linux falls back when xdg empty",
			goos:   "linux",
			uid:    1000,
			xdg:    "",
			winTmp: winTmp,
			want:   "/tmp/1000/minikube",
		},
		{
			name:   "linux falls back when xdg points to missing dir",
			goos:   "linux",
			uid:    1000,
			xdg:    "/nonexistent/xdg/runtime/dir",
			winTmp: winTmp,
			want:   "/tmp/1000/minikube",
		},
		{
			name:   "windows uses winTmp",
			goos:   "windows",
			uid:    -1, // os.Getuid() on Windows
			xdg:    "",
			winTmp: winTmp,
			want:   filepath.Join(winTmp, "minikube"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := baseDir(tc.goos, tc.uid, tc.xdg, tc.winTmp)
			if got != tc.want {
				t.Errorf("baseDir = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSocketPath_PureResolver(t *testing.T) {
	// SocketPath (ensure=false) must not touch the filesystem.
	xdg := t.TempDir()
	got, err := resolveSocketPath("linux", 1000, xdg, "", "minikube", "monitor", false)
	if err != nil {
		t.Fatalf("resolveSocketPath: %v", err)
	}
	wantPrefix := filepath.Join(xdg, "minikube") + string(filepath.Separator)
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("path %q does not start with %q", got, wantPrefix)
	}
	if _, err := os.Stat(filepath.Dir(got)); !os.IsNotExist(err) {
		t.Errorf("parent dir should NOT exist when ensure=false, got stat err = %v", err)
	}
}

func TestEnsureSocketPath_CreatesDir(t *testing.T) {
	t.Run("linux xdg present", func(t *testing.T) {
		xdg := t.TempDir()
		got, err := resolveSocketPath("linux", 1000, xdg, "", "minikube", "monitor", true)
		if err != nil {
			t.Fatalf("resolveSocketPath: %v", err)
		}
		wantPrefix := filepath.Join(xdg, "minikube") + string(filepath.Separator)
		if !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("path %q does not start with %q", got, wantPrefix)
		}
		if !strings.HasSuffix(got, string(filepath.Separator)+"monitor") {
			t.Errorf("path %q does not end with /monitor", got)
		}
		parent := filepath.Dir(got)
		info, err := os.Stat(parent)
		if err != nil {
			t.Fatalf("parent dir not created: %v", err)
		}
		if info.Mode().Perm() != runtimeDirMode {
			t.Errorf("parent dir mode = %v, want %v", info.Mode().Perm(), os.FileMode(runtimeDirMode))
		}
	})

	t.Run("windows uses winTmp", func(t *testing.T) {
		winTmp := t.TempDir()
		got, err := resolveSocketPath("windows", -1, "", winTmp, "minikube", "monitor", true)
		if err != nil {
			t.Fatalf("resolveSocketPath: %v", err)
		}
		wantPrefix := filepath.Join(winTmp, "minikube") + string(filepath.Separator)
		if !strings.HasPrefix(got, wantPrefix) {
			t.Errorf("path %q does not start with %q", got, wantPrefix)
		}
		if _, err := os.Stat(filepath.Dir(got)); err != nil {
			t.Errorf("parent dir not created: %v", err)
		}
	})

	// Cases where the resulting path is rooted under /tmp on the real
	// filesystem (no way to redirect — the whole point is a short path).
	// Use a synthetic high uid and clean the whole subtree afterwards.
	cases := []struct {
		name       string
		goos       string
		uid        int
		xdg        string
		machine    string
		sock       string
		wantPrefix string
		wantSuffix string
		maxPathLen int
	}{
		{
			name:       "linux xdg fallback",
			goos:       "linux",
			uid:        testUID,
			xdg:        "",
			machine:    "minikube",
			sock:       "monitor",
			wantPrefix: "/tmp/" + strconv.Itoa(testUID) + "/minikube/",
			wantSuffix: "/monitor",
			maxPathLen: linuxSunPathLen,
		},
		{
			name:       "darwin always tmp",
			goos:       "darwin",
			uid:        testUID,
			xdg:        "",
			machine:    "minikube",
			sock:       "vmnet-helper.sock",
			wantPrefix: "/tmp/" + strconv.Itoa(testUID) + "/minikube/",
			wantSuffix: "/vmnet-helper.sock",
			maxPathLen: darwinSunPathLen,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() { removeAllRemaining(t, tc.goos) })
			got, err := resolveSocketPath(tc.goos, tc.uid, tc.xdg, "", tc.machine, tc.sock, true)
			if err != nil {
				t.Fatalf("resolveSocketPath: %v", err)
			}
			if !strings.HasPrefix(got, tc.wantPrefix) {
				t.Errorf("path %q does not start with %q", got, tc.wantPrefix)
			}
			if !strings.HasSuffix(got, tc.wantSuffix) {
				t.Errorf("path %q does not end with %q", got, tc.wantSuffix)
			}
			if len(got) > tc.maxPathLen {
				t.Errorf("path length %d exceeds sun_path[%d]: %q", len(got), tc.maxPathLen, got)
			}
		})
	}
}

func TestSocketPath_ErrorCases(t *testing.T) {
	cases := []struct {
		name    string
		machine string
		sock    string
	}{
		{"empty machine name", "", "monitor"},
		{"empty socket name", "minikube", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := resolveSocketPath("linux", 1000, t.TempDir(), "", tc.machine, tc.sock, false); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestSocketPathDeterministic(t *testing.T) {
	xdg := t.TempDir()

	first, err := resolveSocketPath("linux", 1000, xdg, "", "minikube", "monitor", false)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := resolveSocketPath("linux", 1000, xdg, "", "minikube", "monitor", false)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if first != second {
		t.Errorf("non-deterministic path: first=%q second=%q", first, second)
	}
}

func TestMachineHashUnique(t *testing.T) {
	cases := []struct {
		a, b string
	}{
		{"minikube", "minikube-2"},
		{"profile-a", "profile-b"},
		{"", "x"},
	}

	for _, tc := range cases {
		t.Run(tc.a+"_vs_"+tc.b, func(t *testing.T) {
			if machineHash(tc.a) == machineHash(tc.b) {
				t.Errorf("hash collision for %q and %q", tc.a, tc.b)
			}
		})
	}
}

func TestMachineHashLength(t *testing.T) {
	got := machineHash("any-name")
	if len(got) != machineHashLen {
		t.Errorf("hash length = %d, want %d", len(got), machineHashLen)
	}
}

func TestCleanupIdempotent(t *testing.T) {
	xdg := t.TempDir()

	// EnsureSocketPath creates the dir.
	path, err := resolveSocketPath("linux", 1000, xdg, "", "minikube", "monitor", true)
	if err != nil {
		t.Fatalf("resolveSocketPath: %v", err)
	}
	parent := filepath.Dir(path)
	if _, err := os.Stat(parent); err != nil {
		t.Fatalf("parent should exist after EnsureSocketPath: %v", err)
	}

	// First cleanup removes it.
	if err := cleanup("linux", 1000, xdg, "", "minikube"); err != nil {
		t.Fatalf("first cleanup: %v", err)
	}
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		t.Errorf("parent should be gone after cleanup, got %v", err)
	}

	// Second cleanup is a no-op.
	if err := cleanup("linux", 1000, xdg, "", "minikube"); err != nil {
		t.Errorf("second cleanup should be idempotent: %v", err)
	}
}

func TestCleanupEmptyMachineName(t *testing.T) {
	if err := cleanup("linux", 1000, t.TempDir(), "", ""); err == nil {
		t.Error("expected error for empty machine name")
	}
}

// TestPathBudget asserts the hard guarantee of the package: for any plausible
// machine and socket name, the resulting path fits in sockaddr_un.sun_path.
func TestPathBudget(t *testing.T) {
	// Worst observed socket name in the wild: "vmnet-helper.sock-krun.sock"
	// (27 chars). Pick a realistic upper bound.
	longSock := "vmnet-helper.sock-krun.sock"
	// Worst machine name: minikube allows long profile names; cap our test
	// at 64 chars (longer than any default profile).
	longMachine := strings.Repeat("m", 64)

	cases := []struct {
		goos       string
		uid        int
		maxPathLen int
	}{
		{"linux", 4294967295, linuxSunPathLen},
		{"darwin", 4294967295, darwinSunPathLen},
	}

	for _, tc := range cases {
		t.Run(tc.goos, func(t *testing.T) {
			// Pure resolver — no FS writes, no pollution.
			got, err := resolveSocketPath(tc.goos, tc.uid, "", "", longMachine, longSock, false)
			if err != nil {
				t.Fatalf("resolveSocketPath: %v", err)
			}
			if len(got) > tc.maxPathLen {
				t.Errorf("path length %d exceeds sun_path[%d]: %q", len(got), tc.maxPathLen, got)
			}
		})
	}
}
