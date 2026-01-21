//go:build integration

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
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/vmpath"
)

// TestISOImage verifies files and packages installed inside minikube ISO/Base image
func TestISOImage(t *testing.T) {
	if !VMDriver() {
		t.Skip("This test requires a VM driver")
	}

	MaybeParallel(t)

	profile := UniqueProfileName("guest")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(22))
	defer CleanupWithLogs(t, profile, cancel)

	t.Run("Setup", func(t *testing.T) {
		args := append([]string{"start", "-p", profile, "--no-kubernetes", "--memory=2500mb"}, StartArgs()...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Errorf("failed to start minikube: args %q: %v", rr.Command(), err)
		}
	})

	// Run as a group so that our defer doesn't happen as tests are runnings
	t.Run("Binaries", func(t *testing.T) {
		binaries := []string{
			"crictl",
			"curl",
			"docker",
			"git",
			"iptables",
			"podman",
			"rsync",
			"socat",
			"wget",
		}

		// virtualbox is not available in the arm64 iso.
		if runtime.GOARCH == "amd64" {
			binaries = append(binaries, "VBoxControl", "VBoxService")
		}

		for _, pkg := range binaries {
			pkg := pkg
			t.Run(pkg, func(t *testing.T) {
				t.Parallel()
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("which %s", pkg)))
				if err != nil {
					t.Errorf("failed to verify existence of %q binary : args %q: %v", pkg, rr.Command(), err)
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
			vmpath.GuestPersistentDir,
			"/var/lib/toolbox",
			"/var/lib/boot2docker",
		} {
			mount := mount
			t.Run(mount, func(t *testing.T) {
				t.Parallel()
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("df -t ext4 %s | grep %s", mount, mount)))
				if err != nil {
					t.Errorf("failed to verify existence of %q mount. args %q: %v", mount, rr.Command(), err)
				}
			})
		}
	})

	t.Run("VersionJSON", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "cat /version.json"))
		if err != nil {
			t.Fatalf("failed to read /version.json. args %q: %v", rr.Command(), err)
		}

		var data map[string]string
		if err := json.Unmarshal(rr.Stdout.Bytes(), &data); err != nil {
			t.Fatalf("failed to parse /version.json as JSON: %v. \nContent: %s", err, rr.Stdout)
		}

		t.Logf("Successfully parsed /version.json:")
		for k, v := range data {
			t.Logf("  %s: %s", k, v)
		}
	})

	t.Run("eBPFSupport", func(t *testing.T) {
		// Ensure that BTF type information is available (https://github.com/kubernetes/minikube/issues/21788)
		btfFile := "/sys/kernel/btf/vmlinux"
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("test -f %s && echo 'OK' || echo 'NOT FOUND'", btfFile)))
		if err != nil {
			t.Errorf("failed to verify existence of %q file: args %q: %v", btfFile, rr.Command(), err)
		}

		if !strings.Contains(rr.Stdout.String(), "OK") {
			t.Errorf("expected file %q to exist, but it does not. BTF types are required for CO-RE eBPF programs; set CONFIG_DEBUG_INFO_BTF in kernel configuration.", btfFile)
		}
	})

	t.Run("kmodNVMeTCP", func(t *testing.T) {
		kmod := "nvme-tcp"
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("modinfo %s", kmod)))
		if err != nil {
			t.Errorf("failed to get info for kernel module %s: args %q: %v", kmod, rr.Command(), err)
		}
	})
}
