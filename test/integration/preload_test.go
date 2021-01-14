// +build integration

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

func TestPreload(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	if arm64Platform() {
		t.Skipf("skipping %s - not yet supported on arm64", t.Name())
	}

	profile := UniqueProfileName("test-preload")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "--wait=true", "--preload=false"}
	startArgs = append(startArgs, StartArgs()...)
	k8sVersion := "v1.17.0"
	startArgs = append(startArgs, fmt.Sprintf("--kubernetes-version=%s", k8sVersion))

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// Now, pull the busybox image into the VMs docker daemon
	image := "busybox"
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "--", "docker", "pull", image))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// Restart minikube with v1.17.3, which has a preloaded tarball
	startArgs = []string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "-v=1", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)
	k8sVersion = "v1.17.3"
	startArgs = append(startArgs, fmt.Sprintf("--kubernetes-version=%s", k8sVersion))
	rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "--", "docker", "images"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Output(), image) {
		t.Fatalf("Expected to find %s in output of `docker images`, instead got %s", image, rr.Output())
	}
}
