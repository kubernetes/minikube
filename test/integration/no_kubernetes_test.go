//go:build integration
// +build integration

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
	"strings"
	"testing"
)

// TestNoKubernetes tests starting minikube without Kubernetes,
// for use cases where user only needs to use the container runtime (docker, containerd, crio) inside minikube
func TestNoKubernetes(t *testing.T) {
	MaybeParallel(t)

	if NoneDriver() {
		t.Skip("None driver does not need --no-kubernetes test")
	}
	type validateFunc func(context.Context, *testing.T, string)
	profile := UniqueProfileName("NoKubernetes")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer Cleanup(t, profile, cancel)

	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"Start", validateStartNoK8S},
			{"VerifyK8sNotRunning", validateK8SNotRunning},
			{"ProfileList", validateProfileListNoK8S},
			{"Stop", validateStopNoK8S},
			{"StartNoArgs", validateStartNoArgs},
			{"VerifyK8sNotRunningSecond", validateK8SNotRunning},
		}

		for _, tc := range tests {
			tc := tc

			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}

			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
				if t.Failed() && *postMortemLogs {
					PostMortemLogs(t, profile)
				}
			})
		}
	})
}

// validateStartNoK8S starts a minikube cluster without kubernetes started/configured
func validateStartNoK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := append([]string{"start", "-p", profile, "--no-kubernetes"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}

// validateK8SNotRunning validates that there is no kubernetes running inside minikube
func validateK8SNotRunning(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"ssh", "-p", profile, "sudo systemctl is-active --quiet service kubelet"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err == nil {
		t.Fatalf("Expected Kubelet not to be running and but it is running : %q : %v", rr.Command(), err)
	}
}

// validateStopNoK8S validates that minikube is stopped after a --no-kubernetes start
func validateStopNoK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"stop", "-p", profile}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("Failed to stop minikube %q : %v", rr.Command(), err)
	}
}

// validateProfileListNoK8S validates that profile list works with --no-kubernetes
func validateProfileListNoK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := []string{"profile", "list"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("Profile list failed : %q : %v", rr.Command(), err)
	}

	if !strings.Contains(rr.Output(), "N/A") {
		t.Fatalf("expected N/A in the profile list for kubernetes version but got : %q : %v", rr.Command(), rr.Output())
	}

	args = []string{"profile", "list", "--output=json"}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("Profile list --output=json failed : %q : %v", rr.Command(), err)
	}

}

// validateStartNoArgs valides that minikube start with no args works
func validateStartNoArgs(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := append([]string{"start", "-p", profile}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

}
