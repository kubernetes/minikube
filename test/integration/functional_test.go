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
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/test/integration/util"
)

func TestFunctional(t *testing.T) {
	r := NewMinikubeRunner(t)
	r.EnsureRunning()
	// This one is not parallel, and ensures the cluster comes up
	// before we run any other tests.
	t.Run("Status", testClusterStatus)

	t.Run("DNS", testClusterDNS)
	t.Run("Logs", testClusterLogs)
	t.Run("Addons", testAddons)
	t.Run("Dashboard", testDashboard)
	t.Run("ServicesList", testServicesList)
	t.Run("Provisioning", testProvisioning)
	t.Run("Tunnel", testTunnel)

	if !usingNoneDriver(r) {
		t.Run("EnvVars", testClusterEnv)
		t.Run("SSH", testClusterSSH)
		t.Run("IngressController", testIngressController)
		t.Run("Mounting", testMounting)
	}
}

func TestFunctionalContainerd(t *testing.T) {
	r := NewMinikubeRunner(t)

	if usingNoneDriver(r) {
		t.Skip("Can't run containerd backend with none driver")
	}

	if r.GetStatus() != state.None.String() {
		r.RunCommand("delete", true)
	}

	// Build and push new gvisor image for testing.
	repo := fmt.Sprintf("gcr.io/k8s-minikube/test/gvisor-image/%d", time.Now().Unix())
	pushGvisorImage(t, repo)

	defer os.Setenv(constants.MinikubeGvisorAddonImageRepo, "")
	os.Setenv(constants.MinikubeGvisorAddonImageRepo, repo)

	r.Start("--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock")
	t.Run("Gvisor", testGvisor)
	t.Run("GvisorRestart", testGvisorRestart)
	r.RunCommand("delete", true)
}

func pushGvisorImage(t *testing.T, repo string) {
	cmd := exec.Command("make", "push-gvisor-addon-image")
	cmd.Env = append(os.Environ(), []string{fmt.Sprintf("REGISTRY=%s", repo)}...)
	cmd.Dir = "../../"
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error running command: %s %v. Output: %s", cmd.Args, err, string(stdout))
	}
}

// usingNoneDriver returns true if using the none driver
func usingNoneDriver(r util.MinikubeRunner) bool {
	return strings.Contains(r.StartArgs, "--vm-driver=none")
}
