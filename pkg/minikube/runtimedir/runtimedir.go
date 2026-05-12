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

// Package runtimedir resolves per-machine paths for AF_UNIX sockets in a
// directory short enough to fit inside sockaddr_un.sun_path (108 bytes on
// Linux, 104 on macOS/BSD).
//
// Sockets created under MINIKUBE_HOME (default $HOME/.minikube/machines/<name>/)
// can exceed this limit when the home path, username, machine name or socket
// name are long. This package moves only the sockets to a short runtime
// directory while leaving other per-machine state (disk images, logs,
// configs) under MINIKUBE_HOME.
//
// Layout:
//
//	Linux:    $XDG_RUNTIME_DIR/minikube/<hash>/<name>
//	          fallback /tmp/<uid>/minikube/<hash>/<name>
//	macOS:    /tmp/<uid>/minikube/<hash>/<name>
//	Windows:  <os.TempDir()>/minikube/<hash>/<name>
//
// <hash> is the first 32 hex characters of sha256(machineName).
//
// Two entry points are exposed so callers can opt in to filesystem side
// effects:
//
//	SocketPath        — pure resolver, no filesystem writes. Use when
//	                    dialing a socket the server already bound.
//	EnsureSocketPath  — resolves the path and creates the parent directory
//	                    with mode 0700. Use when binding a new socket.
package runtimedir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	// runtimeDirMode is the permission applied to every directory created
	// by this package. AF_UNIX sockets bound inside need only be reachable
	// by the owning user.
	runtimeDirMode = 0o700

	// machineHashLen is the number of hex characters kept from sha256.
	// 32 hex chars = 128 bits, enough to avoid collisions across machine names.
	machineHashLen = 32
)

// SocketPath returns the AF_UNIX-safe path for the given per-machine socket
// without writing to the filesystem. Use this when dialing a socket the server
// already bound (e.g. reading VM state).
//
// The returned path is deterministic for the same (machineName, sockName)
// pair on the same host, so the server and the client agree on the location.
func SocketPath(machineName, sockName string) (string, error) {
	return resolveSocketPath(runtime.GOOS, os.Getuid(), os.Getenv("XDG_RUNTIME_DIR"), os.TempDir(), machineName, sockName, false)
}

// EnsureSocketPath behaves like SocketPath but also creates the parent
// directory with mode 0700. Use this when binding a new socket.
func EnsureSocketPath(machineName, sockName string) (string, error) {
	return resolveSocketPath(runtime.GOOS, os.Getuid(), os.Getenv("XDG_RUNTIME_DIR"), os.TempDir(), machineName, sockName, true)
}

// Cleanup removes the per-machine runtime directory. It is idempotent and
// safe to call when the directory does not exist.
func Cleanup(machineName string) error {
	return cleanup(runtime.GOOS, os.Getuid(), os.Getenv("XDG_RUNTIME_DIR"), os.TempDir(), machineName)
}

// resolveSocketPath is the env-free variant used by the public entry points
// and by tests. When ensure is true, the parent directory is created with
// mode 0700; otherwise the filesystem is left untouched.
func resolveSocketPath(goos string, uid int, xdg, winTmp, machineName, sockName string, ensure bool) (string, error) {
	if machineName == "" {
		return "", fmt.Errorf("machine name is empty")
	}
	if sockName == "" {
		return "", fmt.Errorf("socket name is empty")
	}
	base := baseDir(goos, uid, xdg, winTmp)
	dir := filepath.Join(base, machineHash(machineName))
	if ensure {
		if err := ensureDir(dir); err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, sockName), nil
}

// cleanup is the env-free variant used by Cleanup and tests.
func cleanup(goos string, uid int, xdg, winTmp, machineName string) error {
	if machineName == "" {
		return fmt.Errorf("machine name is empty")
	}
	base := baseDir(goos, uid, xdg, winTmp)
	dir := filepath.Join(base, machineHash(machineName))
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove runtime dir %q: %w", dir, err)
	}
	return nil
}

// baseDir returns the platform-specific minikube runtime directory.
//
//   - macOS always uses /tmp/<uid>/minikube. os.TempDir() on macOS returns a
//     long path under /var/folders that would defeat the whole point.
//   - Linux prefers $XDG_RUNTIME_DIR when it exists and is a directory,
//     falling back to /tmp/<uid>/minikube.
//   - Windows uses os.TempDir() (i.e. %TEMP%). os.Getuid() returns -1 on
//     Windows, so /tmp/<uid> is not usable. AF_UNIX support on Windows 10+
//     has the same sun_path[108] limit; %TEMP% paths are typically short
//     enough.
func baseDir(goos string, uid int, xdg, winTmp string) string {
	if goos == "darwin" {
		return filepath.Join("/tmp", strconv.Itoa(uid), "minikube")
	}
	if goos == "windows" {
		return filepath.Join(winTmp, "minikube")
	}
	if xdg != "" {
		fi, err := os.Stat(xdg)
		if err == nil && fi.IsDir() {
			return filepath.Join(xdg, "minikube")
		}
	}
	return filepath.Join("/tmp", strconv.Itoa(uid), "minikube")
}

// ensureDir creates dir (and parents) with mode 0700, then forces the mode
// with Chmod to bypass any umask that would relax the permission.
func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, runtimeDirMode); err != nil {
		return fmt.Errorf("mkdir %q: %w", dir, err)
	}
	if err := os.Chmod(dir, runtimeDirMode); err != nil {
		return fmt.Errorf("chmod %q: %w", dir, err)
	}
	return nil
}

// machineHash returns a stable per-machine identifier short enough to keep
// the full socket path under sun_path[108]/[104]. See package doc for the
// budget breakdown.
func machineHash(machineName string) string {
	sum := sha256.Sum256([]byte(machineName))
	return hex.EncodeToString(sum[:])[:machineHashLen]
}
