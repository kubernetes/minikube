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
	"testing"
)

func TestFunctional(t *testing.T) {
	p := profileName(t)
	mk := NewMinikubeRunner(t, p)
	stdout, stderr, err := mk.Start()
	if err != nil {
		t.Fatalf("failed to start minikube failed : %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !isTestNoneDriver(t) { // none driver doesn't need to be deleted
		defer mk.TearDown(t)
	}

	// group is needed to make sure tear down runs after parallel runs
	// https://github.com/golang/go/issues/17791#issuecomment-258476786
	t.Run("group", func(t *testing.T) {
		// This one is not parallel, and ensures the cluster comes up
		// before we run any other tests.
		t.Run("Status", testClusterStatus)
		t.Run("ProfileList", testProfileList)
		t.Run("DNS", testClusterDNS)
		t.Run("Logs", testClusterLogs)
		t.Run("Addons", testAddons)
		t.Run("Registry", testRegistry)
		t.Run("Dashboard", testDashboard)
		t.Run("ServicesList", testServicesList)
		t.Run("Provisioning", testProvisioning)
		t.Run("Tunnel", testTunnel)
		t.Run("kubecontext", testKubeConfigCurrentCtx)
		t.Run("config", testConfig)

		if !isTestNoneDriver(t) {
			t.Run("EnvVars", testClusterEnv)
			t.Run("SSH", testClusterSSH)
			t.Run("IngressController", testIngressController)
			t.Run("Mounting", testMounting)
		}

	})

}
