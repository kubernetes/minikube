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
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
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
			{"StartNoK8sWithVersion", validateStartNoK8sWithVersion},
			{"StartWithK8s", validateStartWithK8S},
			{"StartWithStopK8s", validateStartWithStopK8s},
			{"Start", validateStartNoK8S},
			{"VerifyNok8sNoK8sDownloads", VerifyNoK8sDownloadCache},
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

// VerifyNoK8sDownloadCache verifies that starting minikube with --no-kubernetes does not create a download cache.
func VerifyNoK8sDownloadCache(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	cachePath := filepath.Join(localpath.MiniPath(), "cache", runtime.GOOS, runtime.GOARCH, constants.NoKubernetesVersion)

	t.Logf("Checking cache directory: %s", cachePath)
	files, err := filepath.Glob(filepath.Join(cachePath, "*"))
	if err != nil {
		t.Errorf("Error reading cache directory: %v", err)
		return
	}

	if len(files) > 0 {
		t.Logf("Files found in cache directory: %v", files)
		t.Errorf("Cache directory should not contain files when using --no-kubernetes")
	}
}

// validateStartNoK8sWithVersion expect an error when starting a minikube cluster without kubernetes and with a kubernetes version.
func validateStartNoK8sWithVersion(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: start minikube with no kubernetes.
	args := append([]string{"start", "-p", profile, "--no-kubernetes", "--kubernetes-version=" + constants.OldestKubernetesVersion}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err == nil {
		t.Fatalf("expected an error but none was thrown with args: %q", rr.Command())
	}
}

// validateStartWithK8S starts a minikube cluster with Kubernetes started/configured.
func validateStartWithK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: start minikube with Kubernetes.
	args := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	// docs: return an error if Kubernetes is not running.
	if k8sStatus := getK8sStatus(ctx, t, profile); k8sStatus != "Running" {
		t.Errorf("Kubernetes status, got: %s, want: Running", k8sStatus)
	}
}

// validateStartWithStopK8s starts a minikube cluster while stopping Kubernetes.
func validateStartWithStopK8s(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: start minikube with no Kubernetes.
	args := append([]string{"start", "-p", profile, "--no-kubernetes", "--memory=3072", "--alsologtostderr", "-v=5"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}

	// docs: return an error if Kubernetes is not stopped.
	if k8sStatus := getK8sStatus(ctx, t, profile); k8sStatus != "Stopped" {
		t.Errorf("Kubernetes status, got: %s, want: Stopped", k8sStatus)
	}

	// docs: delete minikube profile.
	args = []string{"delete", "-p", profile}
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to delete minikube profile with args: %q : %v", rr.Command(), err)
	}
}

// validateStartNoK8S starts a minikube cluster without kubernetes started/configured
func validateStartNoK8S(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: start minikube with no Kubernetes.
	// Use --cpus=1 to verify that the 2 CPU minimum is not enforced when --no-kubernetes is set.
	args := append([]string{"start", "-p", profile, "--no-kubernetes", "--cpus=1", "--memory=3072", "--alsologtostderr", "-v=5"}, StartArgs()...)
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

// validateStartNoArgs validates that minikube start with no args works.
func validateStartNoArgs(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	args := append([]string{"start", "-p", profile}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("failed to start minikube with args: %q : %v", rr.Command(), err)
	}
}

// getK8sStatus returns whether Kubernetes is running.
func getK8sStatus(ctx context.Context, t *testing.T, profile string) string {
	// Run `minikube status` as JSON output.
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-o", "json"))
	// We expect Kubernetes config to come back as configured, since we started Kubernetes in a previous test.
	if err != nil && rr.ExitCode != 2 {
		t.Errorf("failed to run minikube status with json output. args %q : %v", rr.Command(), err)
	}

	// Unmarshal JSON output.
	var jsonObject map[string]interface{}
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
	if err != nil {
		t.Errorf("failed to decode json from minikube status. args %q. %v", rr.Command(), err)
	}
	return fmt.Sprintf("%s", jsonObject["Kubelet"])
}
