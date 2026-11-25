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
	"runtime"
	"strconv"
	"testing"
)

// fakeBaseDir replaces the base directory with a temporary directory, so tests
// do not depend on the real /tmp and do not leave anything behind.
func fakeBaseDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := baseDir
	baseDir = func() string { return dir }
	t.Cleanup(func() { baseDir = old })
	return dir
}

// fakeXDGRuntimeDir replaces the value of XDG_RUNTIME_DIR.
func fakeXDGRuntimeDir(t *testing.T, value string) {
	t.Helper()
	old := xdgRuntimeDir
	xdgRuntimeDir = func() string { return value }
	t.Cleanup(func() { xdgRuntimeDir = old })
}

// fakeTempDir replaces the temporary directory used on windows.
func fakeTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := tempDir
	tempDir = func() string { return dir }
	t.Cleanup(func() { tempDir = old })
	return dir
}

// fakeDirs isolates the package from the environment on any platform.
func fakeDirs(t *testing.T) (base, tmp string) {
	t.Helper()
	base = fakeBaseDir(t)
	tmp = fakeTempDir(t)
	fakeXDGRuntimeDir(t, "")
	return base, tmp
}

// wantDir returns the expected socket directory when XDG_RUNTIME_DIR is unset.
func wantDir(base, tmp, machineName string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(tmp, "minikube", machineName)
	}
	return filepath.Join(base, "user", strconv.Itoa(os.Getuid()), "minikube", machineName)
}

func TestResolve(t *testing.T) {
	base, tmp := fakeDirs(t)

	tests := []struct {
		name        string
		machineName string
		want        string
		wantErr     bool
	}{
		{
			name:        "machine",
			machineName: "minikube",
			want:        wantDir(base, tmp, "minikube"),
		},
		{
			name:        "node",
			machineName: "minikube-m02",
			want:        wantDir(base, tmp, "minikube-m02"),
		},
		{
			name:        "no machine name",
			machineName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolve(tt.machineName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolve(%q) = %q, want error", tt.machineName, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve(%q): %v", tt.machineName, err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveXDGRuntimeDir(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG_RUNTIME_DIR is used only on linux")
	}
	base, tmp := fakeDirs(t)
	userDir := t.TempDir()

	tests := []struct {
		name string
		xdg  string
		want string
	}{
		{
			name: "existing directory",
			xdg:  userDir,
			want: filepath.Join(userDir, "minikube", "minikube"),
		},
		{
			// A broken system: we cannot create a directory under /run/user,
			// so fall back to the base directory instead of failing.
			name: "missing directory",
			xdg:  filepath.Join(userDir, "missing"),
			want: wantDir(base, tmp, "minikube"),
		},
		{
			name: "unset",
			xdg:  "",
			want: wantDir(base, tmp, "minikube"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeXDGRuntimeDir(t, tt.xdg)
			got, err := resolve("minikube")
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureSocketDir(t *testing.T) {
	fakeDirs(t)

	dir, err := EnsureSocketDir("minikube")
	if err != nil {
		t.Fatalf("EnsureSocketDir: %v", err)
	}
	checkDirMode(t, dir)

	// Ensuring an existing directory must not fail.
	again, err := EnsureSocketDir("minikube")
	if err != nil {
		t.Fatalf("EnsureSocketDir: %v", err)
	}
	if again != dir {
		t.Errorf("got %q, want %q", again, dir)
	}
}

// EnsureSocketDir must fix the mode of a directory created with a more
// permissive mode, for example by a umask allowing group access.
func TestEnsureSocketDirExistingMode(t *testing.T) {
	fakeDirs(t)

	dir, err := resolve("minikube")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := EnsureSocketDir("minikube"); err != nil {
		t.Fatalf("EnsureSocketDir: %v", err)
	}
	checkDirMode(t, dir)
}

func TestEnsureSocketDirNoMachineName(t *testing.T) {
	if _, err := EnsureSocketDir(""); err == nil {
		t.Error("expected error for empty machine name")
	}
}

func TestRemoveSocketDir(t *testing.T) {
	fakeDirs(t)

	dir, err := EnsureSocketDir("minikube")
	if err != nil {
		t.Fatalf("EnsureSocketDir: %v", err)
	}
	if err := RemoveSocketDir("minikube"); err != nil {
		t.Fatalf("RemoveSocketDir: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory %q was not removed: %v", dir, err)
	}

	// Removing a missing directory must not fail.
	if err := RemoveSocketDir("minikube"); err != nil {
		t.Errorf("RemoveSocketDir: %v", err)
	}
}

func TestRemoveSocketDirNoMachineName(t *testing.T) {
	if err := RemoveSocketDir(""); err == nil {
		t.Error("expected error for empty machine name")
	}
}

func checkDirMode(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		// Unix permissions are not supported.
		return
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != dirMode {
		t.Errorf("directory %q mode is %o, want %o", dir, mode, dirMode)
	}
}
