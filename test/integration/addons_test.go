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
	"strings"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/api"
	commonutil "k8s.io/minikube/pkg/util"

	"k8s.io/minikube/test/integration/util"
)

var (
	addonManagerCmd = []string{"get", "pods", "--namespace=kube-system"}
	dashboardRcCmd  = []string{"get", "rc", "kubernetes-dashboard", "--namespace=kube-system"}
	dashboardSvcCmd = []string{"get", "svc", "kubernetes-dashboard", "--namespace=kube-system"}
)

func TestAddons(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		BinaryPath: *binaryPath,
		Args:       *args,
		T:          t}

	minikubeRunner.EnsureRunning()
	kubectlRunner := util.NewKubectlRunner(t)

	checkAddon := func() error {
		pods := api.PodList{}
		if err := kubectlRunner.RunCommandParseOutput(addonManagerCmd, &pods); err != nil {
			return err
		}

		for _, p := range pods.Items {
			if strings.HasPrefix(p.ObjectMeta.Name, "kube-addon-manager-") {
				if p.Status.Phase == "Running" {
					return nil
				}
				return fmt.Errorf("Pod is not Running. Status: %s", p.Status.Phase)
			}
		}

		return fmt.Errorf("Addon manager not found. Found pods: %s", pods)
	}

	if err := commonutil.RetryAfter(20, checkAddon, 5*time.Second); err != nil {
		t.Fatalf("Addon Manager pod is unhealthy: %s", err)
	}
}

func TestDashboard(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		BinaryPath: *binaryPath,
		Args:       *args,
		T:          t}
	minikubeRunner.Start()
	minikubeRunner.CheckStatus("Running")
	kubectlRunner := util.NewKubectlRunner(t)

	checkDashboard := func() error {
		rc := api.ReplicationController{}
		svc := api.Service{}
		if err := kubectlRunner.RunCommandParseOutput(dashboardRcCmd, &rc); err != nil {
			return err
		}

		if err := kubectlRunner.RunCommandParseOutput(dashboardSvcCmd, &svc); err != nil {
			return err
		}

		if rc.Status.Replicas != rc.Status.FullyLabeledReplicas {
			return fmt.Errorf("Not enough pods running. Expected %s, got %s.", rc.Status.Replicas, rc.Status.FullyLabeledReplicas)
		}

		if svc.Spec.Ports[0].NodePort != 30000 {
			return fmt.Errorf("Dashboard is not exposed on port {}", svc.Spec.Ports[0].NodePort)
		}

		return nil
	}

	if err := commonutil.RetryAfter(10, checkDashboard, 5*time.Second); err != nil {
		t.Fatalf("Dashboard is unhealthy: %s", err)
	}
}
