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
	"testing"
)

// TestIngressAddonLegacy tests ingress and ingress-dns addons with legacy k8s version <1.19
func TestIngressAddonLegacy(t *testing.T) {
	if NoneDriver() {
		t.Skipf("skipping: none driver does not support ingress")
	}

	profile := UniqueProfileName("ingress-addon-legacy")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	defer Cleanup(t, profile, cancel)

	t.Run("StartLegacyK8sCluster", func(t *testing.T) {
		args := append([]string{"start", "-p", profile, "--kubernetes-version=v1.18.20", "--memory=4096", "--wait=true", "--alsologtostderr", "-v=5"}, StartArgs()...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Errorf("failed to start minikube with args: %q : %v", rr.Command(), err)
		}
	})

	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"ValidateIngressAddonActivation", validateIngressAddonActivation},
			{"ValidateIngressDNSAddonActivation", validateIngressDNSAddonActivation},
			{"ValidateIngressAddons", validateIngressAddon},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})
}

// validateIngressAddonActivation tests ingress addon activation
func validateIngressAddonActivation(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if _, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "enable", "ingress", "--alsologtostderr", "-v=5")); err != nil {
		t.Errorf("failed to enable ingress addon: %v", err)
	}
}

// validateIngressDNSAddonActivation tests ingress-dns addon activation
func validateIngressDNSAddonActivation(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if _, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "enable", "ingress-dns", "--alsologtostderr", "-v=5")); err != nil {
		t.Errorf("failed to enable ingress-dns addon: %v", err)
	}
}
