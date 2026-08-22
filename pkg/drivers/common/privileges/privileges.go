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

// Package privileges provides helpers for running child processes as the real
// (unprivileged) user when minikube is invoked via sudo. This is needed for CI
// testing on GitHub runners where sudo is required due to Local Network Privacy
// on headless macOS machines. Running minikube as root is not supported.
package privileges

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"k8s.io/minikube/pkg/libmachine/log"
)

// Credential returns the real user's credentials when running under sudo, or
// nil if not running as root or SUDO_UID/SUDO_GID are not set.
func Credential() *syscall.Credential {
	if os.Getuid() != 0 {
		return nil
	}
	uidStr := os.Getenv("SUDO_UID")
	gidStr := os.Getenv("SUDO_GID")
	if uidStr == "" || gidStr == "" {
		return nil
	}
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return nil
	}
	gid, err := strconv.Atoi(gidStr)
	if err != nil {
		return nil
	}
	return &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
}

// ChownDirAll recursively changes ownership of a directory and all its contents
// to the real user. Returns nil immediately if not running under sudo.
func ChownDirAll(dir string) error {
	cred := Credential()
	if cred == nil {
		return nil
	}
	uid := int(cred.Uid)
	gid := int(cred.Gid)
	log.Infof("Chown %s to %d:%d (running under sudo)", dir, uid, gid)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(path, uid, gid)
	})
}
