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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgutil "k8s.io/minikube/pkg/util"

	"k8s.io/minikube/test/integration/util"
)

func testClusterDNS(t *testing.T) {
	t.Parallel()

	kubectlRunner := util.NewKubectlRunner(t)
	podPath := filepath.Join(*testdataDir, "busybox.yaml")

	client, err := pkgutil.GetClient()
	if err != nil {
		t.Fatalf("Error getting kubernetes client %s", err)
	}

	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", podPath}); err != nil {
		t.Fatalf("creating busybox pod: %s", err)
	}

	if err := util.WaitForBusyboxRunning(t, "default"); err != nil {
		t.Fatalf("Waiting for busybox pod to be up: %s", err)
	}
	listOpts := metav1.ListOptions{LabelSelector: "integration-test=busybox"}
	pods, err := client.CoreV1().Pods("default").List(listOpts)
	if len(pods.Items) == 0 {
		t.Fatal("Expected a busybox pod to be running")
	}

	podName := pods.Items[0].Name
	defer kubectlRunner.RunCommand([]string{"delete", "po", podName})

	dnsByteArr, err := kubectlRunner.RunCommand([]string{"exec", podName,
		"nslookup", "kubernetes"})
	if err != nil {
		t.Fatalf("running nslookup in pod:%s", err)
	}
	dnsOutput := string(dnsByteArr)
	if !strings.Contains(dnsOutput, "10.96.0.1") || !strings.Contains(dnsOutput, "10.96.0.10") {
		t.Errorf("DNS lookup failed, could not find both 10.96.0.1 and 10.96.0.10.  Output: %s", dnsOutput)
	}
}
