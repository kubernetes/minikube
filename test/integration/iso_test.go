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
	"fmt"
	"strings"
	"testing"
)

func TestISO(t *testing.T) {
	p := profile(t)
	if isTestNoneDriver() {
		p = "minikube"
	} else {
		t.Parallel()
	}

	mk := NewMinikubeRunner(t, p, "--wait=false")
	mk.RunCommand("delete", false)
	stdout, stderr, err := mk.Start()
	if err != nil {
		t.Fatalf("failed to start minikube (for profile %s) %s) failed : %v\nstdout: %s\nstderr: %s", t.Name(), err, stdout, stderr)
	}
	if !isTestNoneDriver() { // none driver doesn't need to be deleted
		defer mk.TearDown(t)
	}

	t.Run("permissions", testMountPermissions)
	t.Run("packages", testPackages)
	t.Run("persistence", testPersistence)

}

func testMountPermissions(t *testing.T) {
	p := profile(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")
	// test mount permissions
	mountPoints := []string{"/Users", "/hosthome"}
	perms := "drwxr-xr-x"
	foundMount := false

	for _, dir := range mountPoints {
		output, err := mk.SSH(fmt.Sprintf("ls -l %s", dir))
		if err != nil {
			continue
		}
		foundMount = true
		if !strings.Contains(output, perms) {
			t.Fatalf("Incorrect permissions. Expected %s, got %s.", perms, output)
		}
	}
	if !foundMount {
		t.Fatalf("No shared mount found. Checked %s", mountPoints)
	}
}

func testPackages(t *testing.T) {
	p := profile(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")

	packages := []string{
		"git",
		"rsync",
		"curl",
		"wget",
		"socat",
		"iptables",
		"VBoxControl",
		"VBoxService",
	}

	for _, pkg := range packages {
		if output, err := mk.SSH(fmt.Sprintf("which %s", pkg)); err != nil {
			t.Errorf("Error finding package: %s. Error: %v. Output: %s", pkg, err, output)
		}
	}

}

func testPersistence(t *testing.T) {
	p := profile(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")

	for _, dir := range []string{
		"/data",
		"/var/lib/docker",
		"/var/lib/cni",
		"/var/lib/kubelet",
		"/var/lib/minikube",
		"/var/lib/toolbox",
		"/var/lib/boot2docker",
	} {
		output, err := mk.SSH(fmt.Sprintf("df %s | tail -n 1 | awk '{print $1}'", dir))
		if err != nil {
			t.Errorf("Error checking device for %s. Error: %v", dir, err)
		}
		if !strings.Contains(output, "/dev/sda1") {
			t.Errorf("Path %s is not mounted persistently. %s", dir, output)
		}
	}
}
