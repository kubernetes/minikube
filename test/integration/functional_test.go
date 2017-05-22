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
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubernetes_versions"
	"k8s.io/minikube/test/integration/util"
)

func RunVersionedFunctional(t *testing.T, minikubeRunner util.MinikubeRunner) {
	minikubeRunner.EnsureRunning()

	// This one is not parallel, and ensures the cluster comes up
	// before we run any other tests.
	t.Run("Status", testClusterStatus)
	t.Run("DNS", testClusterDNS)
	t.Run("Addons", testAddons)
}

func TestFunctional(t *testing.T) {
	var minikubeRunner util.MinikubeRunner

	if *versioned {
		t.Logf("Running versioned integration tests")
		k8sVersions, err := kubernetes_versions.GetK8sVersionsFromURL(constants.KubernetesVersionGCSURL)
		if err != nil {
			t.Fatalf(err.Error())
		}
		for _, version := range k8sVersions {
			vArgs := fmt.Sprintf("%s --kubernetes-version %s", *args, version.Version)
			minikubeRunner = util.MinikubeRunner{
				BinaryPath: *binaryPath,
				Args:       vArgs,
				T:          t}
			RunVersionedFunctional(t, minikubeRunner)
		}
	}

	minikubeRunner = util.MinikubeRunner{
		BinaryPath: *binaryPath,
		Args:       *args,
		T:          t}

	RunVersionedFunctional(t, minikubeRunner)

	t.Run("EnvVars", testClusterEnv)
	t.Run("Logs", testClusterLogs)
	t.Run("SSH", testClusterSSH)
	t.Run("Systemd", testVMSystemd)
	t.Run("Dashboard", testDashboard)
	t.Run("ServicesList", testServicesList)
	t.Run("Provisioning", testProvisioning)

	if !strings.Contains(*args, "--vm-driver=none") {
		t.Run("EnvVars", testClusterEnv)
		t.Run("SSH", testClusterSSH)
		if runtime.GOOS != "windows" {
			t.Run("Systemd", testVMSystemd)

		}
		// t.Run("Mounting", testMounting)
	}
}
