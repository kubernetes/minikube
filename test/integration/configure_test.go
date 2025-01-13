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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/constants"
)

// TestConfigure tests addons configure command to ensure that configurations are loaded/set correctly
func TestAddonsConfigure(t *testing.T) {
	profile := UniqueProfileName("addons-configure")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	setupSucceeded := t.Run("Setup", func(t *testing.T) {
		// Set an env var to point to our dummy credentials file
		// don't use t.Setenv because we sometimes manually unset the env var later manually
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(*testdataDir, "gcp-creds.json"))
		t.Cleanup(func() {
			os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		})
		if err != nil {
			t.Fatalf("Failed setting GOOGLE_APPLICATION_CREDENTIALS env var: %v", err)
		}

		err = os.Setenv("GOOGLE_CLOUD_PROJECT", "this_is_fake")
		t.Cleanup(func() {
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
		})
		if err != nil {
			t.Fatalf("Failed setting GOOGLE_CLOUD_PROJECT env var: %v", err)
		}

		// MOCK_GOOGLE_TOKEN forces the gcp-auth webhook to use a mock token instead of trying to get a valid one from the credentials.
		os.Setenv("MOCK_GOOGLE_TOKEN", "true")

		// for some reason, (Docker_Cloud_Shell) sets 'MINIKUBE_FORCE_SYSTEMD=true' while having cgroupfs set in docker (and probably os itself), which might make it unstable and occasionally fail:
		// - I1226 15:05:24.834294   11286 out.go:177]   - MINIKUBE_FORCE_SYSTEMD=true
		// - I1226 15:05:25.070037   11286 info.go:266] docker info: {... CgroupDriver:cgroupfs ...}
		// ref: https://storage.googleapis.com/minikube-builds/logs/15463/27154/Docker_Cloud_Shell.html
		// so we override that here to let minikube auto-detect appropriate cgroup driver
		os.Setenv(constants.MinikubeForceSystemdEnv, "")

		// Add more addons as tests are added
		args := append([]string{"start", "-p", profile, "--wait=true", "--memory=4000", "--alsologtostderr", "--addons=registry-creds"}, StartArgs()...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil {
			t.Fatalf("%s failed: %v", rr.Command(), err)
		}

	})

	if !setupSucceeded {
		t.Fatalf("Failed setup for addon configure tests")
	}

	type TestCase = struct {
		name      string
		validator validateFunc
	}
	// Run tests in serial to avoid collision

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []TestCase{
			{"RegistryCreds", validateRegistryCredsAddon},
		}

		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})
}

// validateRegistryCredsAddon tests the registry-creds addon
func validateRegistryCredsAddon(ctx context.Context, t *testing.T, profile string) {
	defer disableAddon(t, "registry-creds", profile)
	defer PostMortemLogs(t, profile)

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("failed to get Kubernetes client for %s : %v", profile, err)
	}

	start := time.Now()
	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "registry-creds", Minutes(6)); err != nil {
		t.Errorf("failed waiting for registry-creds deployment to stabilize: %v", err)
	}
	t.Logf("registry-creds stabilized in %s", time.Since(start))

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "addons", "configure", "registry-creds", "-f", "./testdata/addons_testconfig.json", "-p", profile))
	if err != nil {
		t.Errorf("failed to configure addon. args %q : %v", rr.Command(), err)
	}
}
