//go:build integration

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
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestOffline makes sure minikube works without internet, once the user has cached the necessary images.
// This test has to run after TestDownloadOnly.
func TestOffline(t *testing.T) {
	MaybeParallel(t)
	rt := ContainerRuntime()
	if rt != "docker" && arm64Platform() {
		t.Skipf("skipping %s - only docker runtime supported on arm64. See https://github.com/kubernetes/minikube/issues/10144", t.Name())
	}

	if rt != "docker" && NoneDriver() {
		t.Skipf("skipping %s - incompatible with none driver", t.Name())
	}

	profile := UniqueProfileName(fmt.Sprintf("offline-%s", rt))
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--alsologtostderr", "-v=1", "--memory=3072", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)
	c := exec.CommandContext(ctx, Target(), startArgs...)
	env := os.Environ()
	// RFC1918 address that unlikely to host working a proxy server
	env = append(env, "HTTP_PROXY=172.16.1.1:1")
	env = append(env, "HTTP_PROXYS=172.16.1.1:1")

	c.Env = env
	rr, err := Run(t, c)
	if err != nil {
		// Fatal so that we may collect logs before stop/delete steps
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}
}
