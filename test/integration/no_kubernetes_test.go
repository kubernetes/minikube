//go:build integration

/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
)

// validateStartNoK8S starts a minikube cluster without kubernetes started/configured
func validateStartNoK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// Modified to include --cpus=1 to verify fix for #22152
	args := append([]string{"start", "-p", profile, "--no-kubernetes", "--cpus=1", "--memory=3072", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}
