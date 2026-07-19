//go:build integration

/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/util/retry"
)

// TestTraefikAddon tests the Traefik addon.
//
// Why this test exists:
// This tests the basic functionality of the Traefik addon to ensure it starts up
// correctly and can route traffic.
//
// What we test:
// Path-based HTTP ingress from the host (or via SSH in the guest VM if port forwarding is needed).
//
// Requires:
// Outbound internet access from the guest VM/container (needed to download helm/charts).
// The test will be skipped if there is no outbound connectivity.
func TestTraefikAddon(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("traefik")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(15))
	t.Cleanup(func() { CleanupWithLogs(t, profile, cancel) })

	startArgs := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: args %q: %v", rr.Command(), err)
	}

	// Verify guest VM has outbound internet access (needed to download helm/charts).
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "curl -fsSL --max-time 10 -o /dev/null https://get.helm.sh/helm3-latest-version"))
	if err != nil {
		t.Skip("skipping: guest VM/container has no outbound internet access (required to download helm/charts): https://github.com/kubernetes/minikube/issues/23275")
	}

	// Enable the traefik addon
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "enable", "traefik", "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Fatalf("failed to enable traefik addon: args %q: %v", rr.Command(), err)
	}
	t.Cleanup(func() { disableAddon(t, "traefik", profile) })

	// Deploy the dedicated nginx pod, service and ingress via kustomize
	rr, err = Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "apply", "-k", filepath.Join(*testdataDir, "traefik")))
	if err != nil {
		t.Fatalf("failed to deploy traefik integration test resources: %s: %v", rr.Command(), err)
	}
	t.Cleanup(func() {
		rr, err := Run(t, exec.CommandContext(ctx, KubectlBinary(), "--context", profile, "delete", "-k", filepath.Join(*testdataDir, "traefik")))
		if err != nil {
			t.Logf("failed to delete traefik integration test resources: args %q. %v", rr.Command(), err)
		}
	})

	// Wait for Traefik test pod to be ready (takes up to 15s in CI)
	if _, err := PodWait(ctx, t, profile, "traefik-test", "run=traefik-nginx", Minutes(1)); err != nil {
		t.Fatalf("failed waiting for traefik-nginx pod: %v", err)
	}

	// Retrieve minikube IP
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ip"))
	if err != nil {
		t.Fatalf("failed to retrieve minikube ip. args %q : %v", rr.Command(), err)
	}
	minikubeIP := strings.TrimSpace(rr.Stdout.String())
	testURL := fmt.Sprintf("http://%s/test", minikubeIP)

	// Check that the ingress routes correctly through Traefik using the VM IP
	checkTraefikIngress := func() error {
		var body string
		if NeedsPortForward() {
			t.Logf("Getting %s via ssh curl...", testURL)
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("curl -s %s", testURL)))
			if err != nil {
				return err
			}
			body = strings.TrimSpace(rr.Stdout.String())
		} else {
			t.Logf("Getting %s via host http client...", testURL)
			resp, err := http.Get(testURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			body = strings.TrimSpace(string(b))
		}

		t.Logf("Got response %q", body)
		want := "it works"
		if body != want {
			return fmt.Errorf("response body = %q, want %q", body, want)
		}
		return nil
	}
	if err := retry.Expo(checkTraefikIngress, 500*time.Millisecond, Minutes(1)); err != nil {
		t.Fatalf("failed to get expected response from %s: %v", testURL, err)
	}
}
