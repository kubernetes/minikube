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
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGvisorAddon(t *testing.T) {
	if NoneDriver() {
		t.Skip("Can't run containerd backend with none driver")
	}
	if !*enableGvisor {
		t.Skip("skipping test because --gvisor=false")
	}

	MaybeParallel(t)
	profile := UniqueProfileName("gvisor")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(60))
	defer func() {
		if t.Failed() {
			rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "gvisor", "-n", "kube-system"))
			if err != nil {
				t.Logf("failed to get gvisor post-mortem logs: %v", err)
			}
			t.Logf("gvisor post-mortem: %s:\n%s\n", rr.Command(), rr.Output())
		}
		CleanupWithLogs(t, profile, cancel)
	}()

	startArgs := append([]string{"start", "-p", profile, "--memory=" + Megabytes(2200), "--container-runtime=containerd", "--docker-opt", "containerd=/var/run/containerd/containerd.sock"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: args %q: %v", rr.Command(), err)
	}

	// If it exists, include a locally built gvisor image
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "add", "gcr.io/k8s-minikube/gvisor-addon:2"))
	if err != nil {
		t.Logf("%s failed: %v (won't test local image)", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "enable", "gvisor"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "kube-system", "kubernetes.io/minikube-addons=gvisor", Minutes(4)); err != nil {
		t.Fatalf("failed waiting for 'gvisor controller' pod: %v", err)
	}

	// Create an untrusted workload
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-untrusted.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	// Create gvisor workload
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "nginx-gvisor.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "run=nginx,untrusted=true", Minutes(4)); err != nil {
		t.Errorf("failed waiting for nginx pod: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "run=nginx,runtime=gvisor", Minutes(4)); err != nil {
		t.Errorf("failed waitinf for gvisor pod: %v", err)
	}

	// Ensure that workloads survive a restart
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Fatalf("faild stopping minikube. args %q : %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed starting minikube after a stop. args %q, %v", rr.Command(), err)
	}
	if _, err := PodWait(ctx, t, profile, "kube-system", "kubernetes.io/minikube-addons=gvisor", Minutes(4)); err != nil {
		t.Errorf("failed waiting for 'gvisor controller' pod : %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "run=nginx,untrusted=true", Minutes(4)); err != nil {
		t.Errorf("failed waiting for 'nginx' pod : %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "run=nginx,runtime=gvisor", Minutes(4)); err != nil {
		t.Errorf("failed waiting for 'gvisor' pod : %v", err)
	}
}
