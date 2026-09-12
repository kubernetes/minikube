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

// Package runtimedir resolves the per-machine directory for AF_UNIX sockets.
//
// Sockets created under MINIKUBE_HOME (default $HOME/.minikube/machines/<name>/)
// can exceed sockaddr_un.sun_path (108 bytes on Linux, 104 on macOS) when the
// home path, user name or machine name are long. This package keeps only the
// sockets in a short runtime directory; the rest of the per-machine state
// (disk images, logs, configs, pidfiles) stays under MINIKUBE_HOME.
//
// Layout:
//
//	Linux:   $XDG_RUNTIME_DIR/minikube/<machine>
//	         /tmp/user/<uid>/minikube/<machine> when XDG_RUNTIME_DIR is unset
//	macOS:   /tmp/user/<uid>/minikube/<machine>
//	Windows: %TEMP%\minikube\<machine>
package runtimedir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"k8s.io/klog/v2"
)

// dirMode is the mode of every directory created by this package. Sockets
// bound inside need to be reachable only by the owning user.
const dirMode = 0o700

// Indirections for testing, so tests can run without touching the environment
// or the real temporary directory.
var (
	// baseDir is the root of the fallback runtime directory.
	baseDir = func() string { return "/tmp" }

	// xdgRuntimeDir is the user runtime directory on Linux, typically
	// /run/user/<uid>.
	xdgRuntimeDir = func() string { return os.Getenv("XDG_RUNTIME_DIR") }

	// tempDir is the per-user temporary directory on Windows, typically
	// C:\Users\<username>\AppData\Local\Temp.
	tempDir = os.TempDir
)

// EnsureSocketDir returns the directory holding the AF_UNIX sockets of the
// given machine, creating it with mode 0700 if needed.
func EnsureSocketDir(machineName string) (string, error) {
	dir, err := resolve(machineName)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", dir, err)
	}
	// MkdirAll applies the umask, and the directory may already exist with
	// a more permissive mode.
	if err := os.Chmod(dir, dirMode); err != nil {
		return "", fmt.Errorf("chmod %q: %w", dir, err)
	}
	return dir, nil
}

// RemoveSocketDir removes the socket directory of the given machine. It is
// idempotent: removing a directory that does not exist is not an error.
func RemoveSocketDir(machineName string) error {
	dir, err := resolve(machineName)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove %q: %w", dir, err)
	}
	return nil
}

// resolve returns the socket directory of the given machine without touching
// the filesystem.
func resolve(machineName string) (string, error) {
	if machineName == "" {
		return "", errors.New("machine name is empty")
	}
	switch runtime.GOOS {
	case "windows":
		// User names are limited to 20 characters, so the per-user temporary
		// directory is short enough.
		// https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/net-user
		return filepath.Join(tempDir(), "minikube", machineName), nil
	case "linux":
		xdg := xdgRuntimeDir()
		if xdg == "" {
			break
		}
		// We cannot create the user runtime directory ourselves, since /run
		// and /run/user are owned by root.
		switch _, err := os.Stat(xdg); {
		case err == nil:
			return filepath.Join(xdg, "minikube", machineName), nil
		case errors.Is(err, os.ErrNotExist):
			klog.Warningf("XDG_RUNTIME_DIR %q does not exist, falling back to %s", xdg, baseDir())
		default:
			return "", fmt.Errorf("stat %q: %w", xdg, err)
		}
	}
	// Fallback mirroring the XDG layout (/run/user/<uid>) under the temporary
	// directory. Used on macOS, and on Linux when XDG_RUNTIME_DIR is unset or
	// points to a missing directory.
	return filepath.Join(baseDir(), "user", strconv.Itoa(os.Getuid()), "minikube", machineName), nil
}
