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
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/test/integration/util"
)

func TestPersistence(t *testing.T) {
	minikubeRunner := NewMinikubeRunner(t)
	if strings.Contains(minikubeRunner.StartArgs, "--vm-driver=none") {
		t.Skip("skipping test as none driver does not support persistence")
	}
	minikubeRunner.EnsureRunning()

	kubectlRunner := util.NewKubectlRunner(t)
	podPath, _ := filepath.Abs("testdata/busybox.yaml")

	// Create a pod and wait for it to be running.
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath}); err != nil {
		t.Fatalf("Error creating test pod: %s", err)
	}

	verify := func(t *testing.T) {
		if err := util.WaitForDashboardRunning(t); err != nil {
			t.Fatalf("waiting for dashboard to be up: %s", err)
		}

		if err := util.WaitForBusyboxRunning(t, "default"); err != nil {
			t.Fatalf("waiting for busybox to be up: %s", err)
		}

	}

	// Make sure everything is up before we stop.
	verify(t)

	// Now restart minikube and make sure the pod is still there.
	// minikubeRunner.RunCommand("stop", true)
	// minikubeRunner.CheckStatus("Stopped")
	checkStop := func() error {
		minikubeRunner.RunCommand("stop", true)
		return minikubeRunner.CheckStatusNoFail(state.Stopped.String())
	}

	if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
		t.Fatalf("timed out while checking stopped status: %s", err)
	}

	minikubeRunner.Start()
	minikubeRunner.CheckStatus(state.Running.String())

	// Make sure the same things come up after we've restarted.
	verify(t)
}
