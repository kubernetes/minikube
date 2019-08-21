// +build integration
// +build linux

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
)

func TestNone(t *testing.T) {
	if !isTestNoneDriver(t) {
		t.Skip("Only test none driver.")
	}
	if shouldRunInParallel(t) {
		t.Parallel()
	}

	err := os.Setenv("CHANGE_MINIKUBE_NONE_USER", "true")
	if err != nil {
		t.Fatalf("Failed to setup TestNone: set env: %v", err)
	}

	p := profileName(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")
	mk.RunCommand("delete", false)
	stdout, stderr, err := mk.Start()
	if err != nil {
		t.Fatalf("failed to start minikube (for profile %s) failed : %v\nstdout: %s\nstderr: %s", p, err, stdout, stderr)
	}
	msg := "Configuring local host environment"
	if !strings.Contains(stdout, msg) {
		t.Errorf("Expected: stdout to contain %q, got: %s", msg, stdout)
	}
	msg = "may reduce system security and reliability."
	if !strings.Contains(stderr, msg) {
		t.Errorf("Expected: stderr to contain %q, got: %s", msg, stderr)
	}

	t.Run("minikube permissions", testNoneMinikubeFolderPermissions)
	t.Run("kubeconfig permissions", testNoneKubeConfigPermissions)

}

func testNoneMinikubeFolderPermissions(t *testing.T) {
	username := os.Getenv("SUDO_USER")
	if username == "" {
		t.Fatal("Expected $SUDO_USER env to not be empty")
	}
	u, err := user.Lookup(username)
	if err != nil {
		t.Fatalf("Getting user failed: %v", err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		t.Errorf("Failed to convert uid to int: %v", err)
	}
	info, err := os.Stat(constants.GetMinipath())
	if err != nil {
		t.Fatalf("Failed to get .minikube dir info, %v", err)
	}
	fileUID := info.Sys().(*syscall.Stat_t).Uid

	if fileUID != uint32(uid) {
		t.Errorf("Expected .minikube folder user: %d, got: %d", uint32(uid), fileUID)
	}

}

func testNoneKubeConfigPermissions(t *testing.T) {
	username := os.Getenv("SUDO_USER")
	if username == "" {
		t.Fatal("Expected $SUDO_USER env to not be empty")
	}
	u, err := user.Lookup(username)
	if err != nil {
		t.Fatalf("Getting user failed: %v", err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		t.Errorf("Failed to convert uid to int: %v", err)
	}
	info, err := os.Stat(filepath.Join(u.HomeDir, ".kube/config"))
	if err != nil {
		t.Errorf("Failed to get .minikube dir info, %v", err)
	}
	fileUID := info.Sys().(*syscall.Stat_t).Uid

	if fileUID != uint32(uid) {
		t.Errorf("Expected .minikube folder user: %d, got: %d", uint32(uid), fileUID)
	}

}
