// +build iso

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os/exec"
	"testing"
	"time"
)

func TestGuestEnvironment(t *testing.T) {
	MaybeParallel(t)
	WaitForStartSlot(t)

	profile := UniqueProfileName("guest")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--wait=false"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	// Run as a group so that our defer doesn't happen as tests are runnings
	t.Run("Binaries", func(t *testing.T) {
		for _, pkg := range []string{"git", "rsync", "curl", "wget", "socat", "iptables", "VBoxControl", "VBoxService"} {
			pkg := pkg
			t.Run(pkg, func(t *testing.T) {
				t.Parallel()
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("which %s", pkg)))
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}
			})
		}
	})

	t.Run("PersistentMounts", func(t *testing.T) {
		for _, mount := range []string{
			"/data",
			"/var/lib/docker",
			"/var/lib/cni",
			"/var/lib/kubelet",
			"/var/lib/minikube",
			"/var/lib/toolbox",
			"/var/lib/boot2docker",
		} {
			mount := mount
			t.Run(mount, func(t *testing.T) {
				t.Parallel()
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("df -t ext4 %s | grep %s", mount, mount)))
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}
			})
		}
	})
}
