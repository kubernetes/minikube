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
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/test/integration/util"
)

func TestPersistence(t *testing.T) {
	if isTestNoneDriver() {
		t.Skip("skipping test as none driver does not support persistence")
	}
	t.Parallel()
	p := profile(t) // profile name
	mk := NewMinikubeRunner(t, p)
	defer mk.TearDown(t)
	mk.EnsureRunning()

	kr := util.NewKubectlRunner(t, p)

	if _, err := kr.RunCommand([]string{"create", "-f", filepath.Join(*testdataDir, "busybox.yaml")}); err != nil {
		t.Fatalf("creating busybox pod: %s", err)
	}

	verifyBusybox := func(t *testing.T) {
		if err := util.WaitForBusyboxRunning(t, "default", p); err != nil {
			t.Fatalf("waiting for busybox to be up: %v", err)
		}

	}
	// Make sure everything is up before we stop.
	verifyBusybox(t)

	// Now restart minikube and make sure the pod is still there.
	checkStop := func() error {
		mk.RunCommand("stop", false)
		return mk.CheckStatusNoFail(state.Stopped.String())
	}
	if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
		t.Fatalf("TestPersistence Failed to stop minikube : %v", err)
	}

	stdout, stderr, err := mk.StartWithStds(15 * time.Minute)
	if err != nil {
		t.Fatalf("%s minikube start failed : %v\nstdout: %s\nstderr: %s", t.Name(), err, stdout, stderr)
	}
	mk.CheckStatus(state.Running.String())

	// Make sure the same things come up after we've restarted.
	verifyBusybox(t)
}
