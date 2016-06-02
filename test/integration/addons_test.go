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
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	commonutil "k8s.io/minikube/pkg/util"

	"k8s.io/minikube/test/integration/util"
)

var (
	addonManagerCmd = []string{"get", "pod", "kube-addon-manager-127.0.0.1", "--namespace=kube-system"}
	dashboardCmd    = []string{"get", "rc", "kubernetes-dashboard", "--namespace=kube-system"}
)

func TestAddons(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{BinaryPath: *binaryPath, T: t}
	minikubeRunner.RunCommand("start", true)
	minikubeRunner.CheckStatus("Running")
	kubectlRunner := util.NewKubectlRunner(t)

	checkAddon := func() error {
		p := api.Pod{}
		if err := kubectlRunner.RunCommandParseOutput(addonManagerCmd, &p); err != nil {
			return err
		}

		if p.Status.Phase != "Running" {
			return fmt.Errorf("Pod is not Running. Status: %s", p.Status.Phase)
		}
		return nil
	}

	if err := commonutil.RetryAfter(10, checkAddon, 5*time.Second); err != nil {
		t.Fatalf("Addon Manager pod is unhealthy: %s", err)
	}
}

func TestDashboard(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{BinaryPath: *binaryPath, T: t}
	minikubeRunner.RunCommand("start", true)
	minikubeRunner.CheckStatus("Running")
	kubectlRunner := util.NewKubectlRunner(t)

	checkDashboard := func() error {
		rc := api.ReplicationController{}
		if err := kubectlRunner.RunCommandParseOutput(dashboardCmd, &rc); err != nil {
			return err
		}

		if rc.Status.Replicas != rc.Status.FullyLabeledReplicas {
			return fmt.Errorf("Not enough pods running. Expected %s, got %s.", rc.Status.Replicas, rc.Status.FullyLabeledReplicas)
		}
		return nil
	}

	if err := commonutil.RetryAfter(10, checkDashboard, 5*time.Second); err != nil {
		t.Fatalf("Dashboard is unhealthy: %s", err)
	}
}
