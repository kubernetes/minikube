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
	"bytes"
	"path/filepath"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	commonutil "k8s.io/minikube/pkg/util"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testClusterDNS(t *testing.T) {
	t.Parallel()
	client, err := pkgutil.GetClient()
	if err != nil {
		t.Fatalf("Error getting kubernetes client %v", err)
	}
	waitForDNS(t, client)

	kr := util.NewKubectlRunner(t)
	busybox := busyBoxPod(t, client, kr)
	defer kr.RunCommand([]string{"delete", "po", busybox})

	// The query result is not as important as service reachability
	out, err := kr.RunCommand([]string{"exec", busybox, "nslookup", "localhost"})
	if err != nil {
		t.Errorf("nslookup within busybox failed: %v", err)
	}
	clusterIP := []byte("10.96.0.1")
	if !bytes.Contains(out, clusterIP) {
		t.Errorf("nslookup did not mention %s:\n%s", clusterIP, out)
	}
}

func waitForDNS(t *testing.T, c kubernetes.Interface) {
	// Implementation note: both kube-dns and coredns have k8s-app=kube-dns labels.
	sel := labels.SelectorFromSet(labels.Set(map[string]string{"k8s-app": "kube-dns"}))
	if err := commonutil.WaitForPodsWithLabelRunning(c, "kube-system", sel); err != nil {
		t.Fatalf("Waited too long for k8s-app=kube-dns pods")
	}
	if err := commonutil.WaitForDeploymentToStabilize(c, "kube-system", "kube-dns", time.Minute*2); err != nil {
		t.Fatalf("kube-dns deployment failed to stabilize within 2 minutes")
	}
}

func busyBoxPod(t *testing.T, c kubernetes.Interface, kr *util.KubectlRunner) string {
	if _, err := kr.RunCommand([]string{"create", "-f", filepath.Join(*testdataDir, "busybox.yaml")}); err != nil {
		t.Fatalf("creating busybox pod: %s", err)
	}
	// TODO(tstromberg): Refactor WaitForBusyboxRunning to return name of pod.
	if err := util.WaitForBusyboxRunning(t, "default"); err != nil {
		t.Fatalf("Waiting for busybox pod to be up: %v", err)
	}

	pods, err := c.CoreV1().Pods("default").List(metav1.ListOptions{LabelSelector: "integration-test=busybox"})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(pods.Items) == 0 {
		t.Fatal("Expected a busybox pod to be running")
	}
	return pods.Items[0].Name
}
