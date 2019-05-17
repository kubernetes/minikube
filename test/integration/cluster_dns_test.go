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
	"k8s.io/client-go/kubernetes"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testClusterDNS(t *testing.T) {
	t.Parallel()
	client, err := pkgutil.GetClient()
	if err != nil {
		t.Fatalf("Error getting kubernetes client %v", err)
	}

	kr := util.NewKubectlRunner(t)
	busybox := busyBoxPod(t, client, kr)
	defer func() {
		if _, err := kr.RunCommand([]string{"delete", "po", busybox}); err != nil {
			t.Errorf("delete failed: %v", err)
		}
	}()

	out := []byte{}

	nslookup := func() error {
		out, err = kr.RunCommand([]string{"exec", busybox, "nslookup", "kubernetes.default"})
		return err
	}
	if err := util.Retry(t, nslookup, 3*time.Second, 60); err != nil {
		t.Fatalf(err.Error())
	}

	clusterIP := []byte("10.96.0.1")
	if !bytes.Contains(out, clusterIP) {
		t.Errorf("output did not contain expected IP:\n%s", out)
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
