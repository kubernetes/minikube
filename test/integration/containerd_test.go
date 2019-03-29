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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/test/integration/util"
)

func TestContainerd(t *testing.T) {
	if isTestNoneDriver(t) {
		t.Skip("Can't run containerd backend with none driver")
	}
	if shouldRunInParallel(t) {
		t.Parallel()
	}

	t.Run("GvisorUntrustedWorkload", testGvisorUntrustedWorkload)
	t.Run("GvisorRuntimeClass", testGvisorRuntimeClass)
	t.Run("GvisorRestart", testGvisorRestart)
}

func testGvisorUntrustedWorkload(t *testing.T) {
	p := profileName(t)
	if shouldRunInParallel(t) {
		t.Parallel()
	}
	mk := NewMinikubeRunner(t, p, "--wait=false")
	defer mk.TearDown(t)

	mk.MustRun("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := waitForGvisorControllerRunning(p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t, p)
	t.Log("making sure untrusted workload is Running")
	if err := waitForUntrustedNginxRunning(p); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}
	deleteUntrustedWorkload(t, p)
}

func testGvisorRuntimeClass(t *testing.T) {
	p := profileName(t)
	if shouldRunInParallel(t) {
		t.Parallel()
	}
	mk := NewMinikubeRunner(t, p, "--wait=false")
	defer mk.TearDown(t)

	mk.MustRun("addons enable gvisor", true)

	t.Log("waiting for gvisor controller to come up")
	if err := waitForGvisorControllerRunning(p); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	createGvisorWorkload(t, p)
	t.Log("making sure gvisor workload is Running")
	if err := waitForGvisorNginxRunning(p); err != nil {
		t.Fatalf("waiting for nginx to be up: %v", err)
	}
	deleteGvisorWorkload(t, p)
}

func testGvisorRestart(t *testing.T) {
	p := profileName(t)
	if shouldRunInParallel(t) {
		t.Parallel()
	}
	mk := NewMinikubeRunner(t, p, "--wait=false")
	defer mk.TearDown(t)

	mk.MustStart("--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock")
	mk.MustRun("cache add gcr.io/k8s-minikube/gvisor-addon:latest")
	mk.MustRun("addons enable gvisor")

	t.Log("waiting for gvisor controller to come up")
	if err := waitForGvisorControllerRunning(p); err != nil {
		t.Errorf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t, p)
	t.Log("making sure untrusted workload is Running")
	if err := waitForUntrustedNginxRunning(p); err != nil {
		t.Errorf("waiting for nginx to be up: %v", err)
	}
	deleteUntrustedWorkload(t, p)

	mk.MustRun("delete")
	mk.MustStart("--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock")
	mk.CheckStatus(state.Running.String())

	t.Log("waiting for gvisor controller to come up")
	if err := waitForGvisorControllerRunning(p); err != nil {
		t.Errorf("waiting for gvisor controller to be up: %v", err)
	}

	createUntrustedWorkload(t, p)
	t.Log("making sure untrusted workload is Running")
	if err := waitForUntrustedNginxRunning(p); err != nil {
		t.Errorf("waiting for nginx to be up: %v", err)
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

func createGvisorWorkload(t *testing.T, profile string) {
	kr := util.NewKubectlRunner(t, profile)
	gvisorPath := filepath.Join(*testdataDir, "nginx-gvisor.yaml")
	t.Log("creating pod with gvisor workload annotation")
	if _, err := kr.RunCommand([]string{"replace", "-f", gvisorPath, "--force"}); err != nil {
		t.Fatalf("creating gvisor nginx resource: %v", err)
	}
}

func deleteGvisorWorkload(t *testing.T, profile string) {
	kr := util.NewKubectlRunner(t, profile)
	gvisorPath := filepath.Join(*testdataDir, "nginx-gvisor.yaml")
	if _, err := kr.RunCommand([]string{"delete", "-f", gvisorPath}); err != nil {
		t.Logf("error deleting gvisor nginx resource: %v", err)
	}
}

// waitForGvisorControllerRunning waits for the gvisor controller pod to be running.
func waitForGvisorControllerRunning(p string) error {
	client, err := kapi.Client(p)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"kubernetes.io/minikube-addons": "gvisor"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "waiting for gvisor controller pod to stabilize")
	}
	return nil
}

// waitForUntrustedNginxRunning waits for the untrusted nginx pod to start
// running.
func waitForUntrustedNginxRunning(miniProfile string) error {
	client, err := kapi.Client(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx", "untrusted": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}
	return nil
}

// waitForGvisorNginxRunning waits for the nginx pod with gvisor runtime class
// to start running.
func waitForGvisorNginxRunning(miniProfile string) error {
	client, err := kapi.Client(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx", "runtime": "gvisor"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		return errors.Wrap(err, "waiting for nginx pods")
	}
	return nil
}
