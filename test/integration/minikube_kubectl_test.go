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
	"os/exec"
	"testing"
	"time"
)

func TestMinikubeKubectlCmd(t *testing.T) {
	MaybeParallel(t)
	profile := UniqueProfileName("kubectl-command")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--alsologtostderr", "-v=3", "--wait=true"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	kubectlArgs := []string{"kubectl", "--", "get", "pods"}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), kubectlArgs...))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
}
