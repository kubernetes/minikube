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
	"context"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/kapi"
)

func TestGvisor(t *testing.T) {
	if NoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}
	MaybeParallel(t)

	profile := Profile("gvisor")
	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock", "--wait=false"}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	// TODO: Re-examine if we should be pulling in an image which users don't normally use
	rr, err = RunCmd(ctx, t, Target(), "-p", profile, "cache", "add", "gcr.io/k8s-minikube/gvisor-addon:latest")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	rr, err = RunCmd(ctx, t, Target(), "addons", "enable", "gvisor")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	// mostly because addons are persistent across profiles :(
	defer func() {
		rr, err := RunCmd(ctx, t, Target(), "addons", "disable", "gvisor")
		if err != nil {
			t.Logf("%s failed: %v", rr.Args, err)
		}
	}()

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"kubernetes.io/minikube-addons": "gvisor"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Errorf("waiting for gvisor controller pod to stabilize: %v", err)
	}

	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	selector = labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Errorf("waiting for nginx pods: %v", err)
	}
}
