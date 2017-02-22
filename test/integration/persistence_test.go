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
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/kubernetes/pkg/api"
	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func TestPersistence(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{BinaryPath: *binaryPath, T: t}
	minikubeRunner.EnsureRunning()

	kubectlRunner := util.NewKubectlRunner(t)
	podName := "busybox"
	podPath, _ := filepath.Abs("testdata/busybox.yaml")

	podNamespace := kubectlRunner.CreateRandomNamespace()
	defer kubectlRunner.DeleteNamespace(podNamespace)

	// Create a pod and wait for it to be running.
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath, "--namespace=" + podNamespace}); err != nil {
		t.Fatalf("Error creating test pod: %s", err)
	}

	checkPod := func() error {
		p, err := kubectlRunner.GetPod(podName, podNamespace)
		if err != nil {
			return &commonutil.RetriableError{Err: err}
		}
		if kubectlRunner.IsPodReady(p) {
			return nil
		}
		return &commonutil.RetriableError{Err: fmt.Errorf("Pod %s is not ready yet.", podName)}
	}

	if err := commonutil.RetryAfter(20, checkPod, 6*time.Second); err != nil {
		t.Fatalf("Error checking the status of pod %s. Err: %s", podName, err)
	}

	checkDashboard := func() error {
		pods := api.PodList{}
		cmd := []string{"get", "pods", "--namespace=kube-system", "--selector=app=kubernetes-dashboard"}
		if err := kubectlRunner.RunCommandParseOutput(cmd, &pods); err != nil {
			return err
		}
		if len(pods.Items) < 1 {
			return &commonutil.RetriableError{Err: fmt.Errorf("No pods found matching query: %v", cmd)}
		}
		db := pods.Items[0]
		if kubectlRunner.IsPodReady(&db) {
			return nil
		}
		return &commonutil.RetriableError{Err: fmt.Errorf("Dashboard pod is not ready yet.")}
	}

	// Make sure the dashboard is running before we stop the VM.
	// On slow networks it can take several minutes to pull the addon-manager then the dashboard image.
	if err := commonutil.RetryAfter(25, checkDashboard, 20*time.Second); err != nil {
		t.Fatalf("Dashboard pod is not healthy: %s", err)
	}

	// Now restart minikube and make sure the pod is still there.
	// minikubeRunner.RunCommand("stop", true)
	// minikubeRunner.CheckStatus("Stopped")
	checkStop := func() error {
		minikubeRunner.RunCommand("stop", true)
		return minikubeRunner.CheckStatusNoFail(state.Stopped.String())
	}

	if err := commonutil.RetryAfter(6, checkStop, 5*time.Second); err != nil {
		t.Fatalf("timed out while checking stopped status: %s", err)
	}

	minikubeRunner.Start()
	minikubeRunner.CheckStatus(state.Running.String())

	if err := commonutil.RetryAfter(5, checkPod, 3*time.Second); err != nil {
		t.Fatalf("Error checking the status of pod %s. Err: %s", podName, err)
	}

	// Now make sure it's still running after.
	if err := commonutil.RetryAfter(5, checkDashboard, 3*time.Second); err != nil {
		t.Fatalf("Dashboard pod is not healthy: %s", err)
	}
}
