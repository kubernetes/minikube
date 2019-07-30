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
	p := t.Name()
	if isTestNoneDriver() {
		p = "minikube"
	} else {
		t.Parallel()
	}

	mk := NewMinikubeRunner(t, p)
	if isTestNoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}

	if mk.GetStatus() != state.None.String() {
		mk.RunCommand("delete", true)
	}
	mk.Start("--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock")
	t.Run("Gvisor", testGvisor)
	t.Run("GvisorRestart", testGvisorRestart)
	mk.RunCommand("delete", true)
}

func testGvisor(t *testing.T) {
	p := "TestContainerd"
	mk := NewMinikubeRunner(t, p, "--wait=false")
	mk.RunCommand("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t, p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t, p)

	t.Log("making sure untrusted workload is Running")
	if err := util.WaitForUntrustedNginxRunning(p); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}

	t.Log("disabling gvisor addon")
	mk.RunCommand("addons disable gvisor", true)
	t.Log("waiting for gvisor controller pod to be deleted")
	if err := util.WaitForGvisorControllerDeleted(p); err != nil {
		t.Fatalf("waiting for gvisor controller to be deleted: %v", err)
	}

	createUntrustedWorkload(t, p)

	t.Log("waiting for FailedCreatePodSandBox event")
	if err := util.WaitForFailedCreatePodSandBoxEvent(p); err != nil {
		t.Fatalf("waiting for FailedCreatePodSandBox event: %v", err)
	}
	deleteUntrustedWorkload(t, p)
}

func testGvisorRestart(t *testing.T) {
	p := "TestContainerd"
	mk := NewMinikubeRunner(t, p, "--wait=false")
	mk.EnsureRunning()
	mk.RunCommand("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := util.WaitForGvisorControllerRunning(t, p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	// TODO: @priyawadhwa to add test for stop as well
	mk.RunCommand("delete", false)
	mk.CheckStatus(state.None.String())
	stdout, stderr, err := mk.StartWithStds(15 * time.Minute)
	if err != nil {
		t.Fatalf("%s minikube start failed : %v\nstdout: %s\nstderr: %s", t.Name() err, stdout, stderr)
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
