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
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/kapi"
)

func TestGvisor(t *testing.T) {
	if NoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}
	MaybeParallel(t)


	profile := Profile("gvisor")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock", "--wait=false"}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("kubernetes client: %v", client)
	}

	// TODO: Re-examine if we should be pulling in an image which users don't normally use
	rr, err = RunCmd(ctx, t, Target(), "-p", profile, "cache", "add", "gcr.io/k8s-minikube/gvisor-addon:latest")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	// NOTE: addons are global, but the addon must assert that the runtime is containerd
	rr, err = RunCmd(ctx, t, Target(), "-p", profile, "addons", "enable", "gvisor")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	// mostly because addons are persistent across profiles :(
	defer func() {
		rr, err := RunCmd(context.Background(), t, Target(), "-p", profile, "addons", "disable", "gvisor")
		if err != nil {
			t.Logf("%s failed: %v", rr.Args, err)
		}
	}()

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"kubernetes.io/minikube-addons": "gvisor"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}
	log.Infof("gvisor controller is up")

	// Create an untrusted workload
	validateUntrustedWorkload(ctx, t, client, profile)

	/*
	// TODO(tstromberg): Investigate whether or not it's beneficial to Kill minikube and start over again
	args := append([]string{"delete", "-p", profile)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	args := append([]string{"start", "-p", profile, "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock", "--wait=false"}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	// Re-create the untrusted workload
	validateUntrustedWorkload(ctx, t, client, profile)

	// Create gvisor workload
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	selector = labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx", "runtime": "gvisor"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		describePod(ctx, t, profile, "nginx-gvisor")
		t.Fatalf("waiting for nginx pods: %v", err)
	}
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "delete", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	*/
}

// for debugging purposes
func debugDescribePod(ctx context.Context, t *testing.T, profile string, pod string) {
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "describe", "pod", pod)
	if err != nil {
		t.Logf("unable to describe nginx-untrusted: %v", err)
		return
	}
	t.Logf("%s pod status:\n%s", pod, rr.Stdout)
}

// validate untrusted workloads (this step is run twice)
func validateUntrustedWorkload(ctx context.Context, t *testing.T, client kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"run": "nginx", "untrusted": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		debugDescribePod(ctx, t, profile, "nginx-untrusted")
		t.Fatalf("waiting for nginx-untrusted: %v", err)
	}

	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "delete", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
}