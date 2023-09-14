//go:build integration

/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// TestPreload verifies the preload tarballs get pulled in properly by minikube
func TestPreload(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	profile := UniqueProfileName("test-preload")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "--wait=true", "--preload=false"}
	startArgs = append(startArgs, StartArgs()...)
	k8sVersion := "v1.24.4"
	startArgs = append(startArgs, fmt.Sprintf("--kubernetes-version=%s", k8sVersion))

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// Now, pull the busybox image into minikube
	image := "gcr.io/k8s-minikube/busybox"
	cmd := exec.CommandContext(ctx, Target(), "-p", profile, "image", "pull", image)
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// stop the cluster
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// re-start the cluster and check if image is preserved
	startArgs = []string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "-v=1", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	cmd = exec.CommandContext(ctx, Target(), "-p", profile, "image", "list")
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), image) {
		t.Fatalf("Expected to find %s in image list output, instead got %s", image, rr.Output())
	}
}
