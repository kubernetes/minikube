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
	"context"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// TestChangeNoneUser tests to make sure the CHANGE_MINIKUBE_NONE_USER environemt variable is respected
// and changes the minikube file permissions from root to the correct user.
func TestChangeNoneUser(t *testing.T) {
	if !NoneDriver() {
		t.Skip("Only test none driver.")
	}
	MaybeParallel(t)

	profile := UniqueProfileName("none")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{"CHANGE_MINIKUBE_NONE_USER=true", Target(), "start", "--wait=false"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "delete"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, "/usr/bin/env", startArgs...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "status"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

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

	// Retrieve the kube config from env
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		kubeConfig = filepath.Join(u.HomeDir, ".kube/config")
	}

	for _, p := range []string{localpath.MiniPath(), kubeConfig} {
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("stat(%s): %v", p, err)
			continue
		}
		if info == nil || info.Sys() == nil {
			t.Errorf("nil info for %s", p)
			continue
		}
		got := info.Sys().(*syscall.Stat_t).Uid
		if got != uint32(uid) {
			t.Errorf("uid(%s) = %d, want %d", p, got, uint32(uid))
		}
	}
}
