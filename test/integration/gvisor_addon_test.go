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
)

func TestGvisorAddon(t *testing.T) {
	if NoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}
	MaybeParallel(t)

	profile := Profile("gvisor")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer func() {
		CleanupWithLogs(t, profile, cancel)
	}()

	args := append([]string{"start", "-p", profile, "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock", "--wait=false"}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	/*
		// TODO: Re-examine if we should be pulling in an image which users don't normally use
		rr, err = RunCmd(ctx, t, Target(), "-p", profile, "cache", "add", "gcr.io/k8s-minikube/gvisor-addon:latest")
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}
	*/
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

	if err := WaitForPods(ctx, t, profile, "kube-system", "kubernetes.io/minikube-addons=gvisor", 2*time.Minute); err != nil {
		t.Fatalf("waiting for gvisor controller to be up: %v", err)
	}

	// Create an untrusted workload
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	// Create gvisor workload
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	if err := WaitForPods(ctx, t, profile, "kube-system", "run=nginx,untrusted=true", 2*time.Minute); err != nil {
		t.Fatalf("nginx: %v", err)
	}
	if err := WaitForPods(ctx, t, profile, "kube-system", "run=nginx,runtime=gvisor", 2*time.Minute); err != nil {
		t.Fatalf("nginx: %v", err)
	}

	/*
		// TODO(tstromberg): Investigate whether or not it's beneficial to Kill minikube and start over again
		rr, err = RunCmd(ctx, t, Target(), "delete", "-p", profile)
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}
		args = append([]string{"start", "-p", profile, "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock", "--wait=false"}, StartArgs()...)
		rr, err = RunCmd(ctx, t, Target(), args...)
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Args, err)
		}

		// Create an untrusted workload (again???)
		rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Args, err)
		}
		if err := WaitForPods(ctx, t, profile, "kube-system", "run=nginx,untrusted=true", 2*time.Minute); err != nil {
			t.Fatalf("nginx: %v", err)
		}
		rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "delete", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml"))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Args, err)
		}

		// Create gvisor workload
		rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml"))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Args, err)
		}

		if err := WaitForPods(ctx, t, profile, "kube-system", "run=nginx,runtime=gvisor", 2*time.Minute); err != nil {
			t.Fatalf("nginx: %v", err)
		}

		rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "delete", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml"))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Args, err)
		}
	*/
}
