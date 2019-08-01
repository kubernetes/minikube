// +build integration

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
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/test/integration/util"
)

func TestContainerd(t *testing.T) {
	p := profile(t)
	if isTestNoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}
	t.Parallel()
	mk := NewMinikubeRunner(t, p)
	defer mk.TearDown(t)

	stdout, stderr, err := mk.Start("--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock")
	if err != nil {
		t.Fatalf("failed to start minikube (for profile %s) failed : %v\nstdout: %s\nstderr: %s", p, err, stdout, stderr)
	}

	t.Run("group", func(t *testing.T) {
		t.Run("GvisorRestart", testGvisorRestart)
	})
}

func testGvisorRestart(t *testing.T) {
	p := profile(t)
	mk := NewMinikubeRunner(t, p, "--wait=false")

	mk.RunCommand("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t, p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	mk.RunCommand("stop", true)
	stdout, stderr, err := mk.Start()
	if err != nil {
		t.Fatalf("failed to start minikube (for profile %s) failed : %v \nstdout: %s \nstderr: %s", t.Name(), err, stdout, stderr)
	}
	mk.CheckStatus(state.Running.String())

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t, p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t, p)
	t.Log("making sure untrusted workload is Running")
	if err := util.WaitForUntrustedNginxRunning(p); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}
	deleteUntrustedWorkload(t, p)
}

func createUntrustedWorkload(t *testing.T, profile string) {
	kr := util.NewKubectlRunner(t, profile)
	untrustedPath := filepath.Join(*testdataDir, "nginx-untrusted.yaml")
	t.Log("creating pod with untrusted workload annotation")
	if _, err := kr.RunCommand([]string{"replace", "-f", untrustedPath, "--force"}); err != nil {
		t.Fatalf("creating untrusted nginx resource: %v", err)
	}
}

func deleteUntrustedWorkload(t *testing.T, profile string) {
	kr := util.NewKubectlRunner(t, profile)
	untrustedPath := filepath.Join(*testdataDir, "nginx-untrusted.yaml")
	if _, err := kr.RunCommand([]string{"delete", "-f", untrustedPath}); err != nil {
		t.Logf("error deleting untrusted nginx resource: %v", err)
	}
}
