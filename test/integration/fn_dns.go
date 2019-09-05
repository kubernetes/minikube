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
	"context"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/kapi"
)

func validateClusterDNS(ctx context.Context, t *testing.T, profile string) {
	MaybeParallel(t)

	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Cmd.Args, err)
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Errorf("client failed: %v", err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Errorf("wait failed: %v", err)
	}
	pods, err := client.CoreV1().Pods("default").List(metav1.ListOptions{LabelSelector: "integration-test=busybox"})
	if err != nil {
		t.Errorf("list error: %v", err)
	}
	pod := pods.Items[0].Name

	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "exec", pod, "nslookup", "kubernetes.default")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Cmd.Args, err)
	}

	want := []byte("10.96.0.1")
	if !bytes.Contains(rr.Stdout.Bytes(), want) {
		t.Errorf("nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
	}
}

