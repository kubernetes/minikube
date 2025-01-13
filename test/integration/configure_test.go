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
	"path/filepath"
	"regexp"
	"strings"
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

	// Check a few secrets exists that match our test data
	// In our test aws and gcp are set, docker and acr are disabled - so they will be set to "changeme"
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "-n", "kube-system", "get", "secret", "-o", "yaml"))
	if err != nil {
		t.Errorf("failed to get secrets. args %q : %v", rr.Command(), err)
	}

	expected := []string{
		"DOCKER_PRIVATE_REGISTRY_PASSWORD: Y2hhbmdlbWU=",
		"DOCKER_PRIVATE_REGISTRY_SERVER: Y2hhbmdlbWU=",
		"DOCKER_PRIVATE_REGISTRY_USER: Y2hhbmdlbWU=",

		"ACR_CLIENT_ID: Y2hhbmdlbWU=",
		"ACR_PASSWORD: Y2hhbmdlbWU=",
		"ACR_URL: Y2hhbmdlbWU=",

		"AWS_ACCESS_KEY_ID: dGVzdF9hd3NfYWNjZXNzaWQ=",
		"AWS_SECRET_ACCESS_KEY: dGVzdF9hd3NfYWNjZXNza2V5",
		"AWS_SESSION_TOKEN: dGVzdF9hd3Nfc2Vzc2lvbl90b2tlbg==",
		"aws-account: dGVzdF9hd3NfYWNjb3VudA==",
		"aws-assume-role: dGVzdF9hd3Nfcm9sZQ==",
		"aws-region: dGVzdF9hd3NfcmVnaW9u",

		"application_default_credentials.json: ewogICJjbGllbnRfaWQiOiAiaGFoYSIsCiAgImNsaWVudF9zZWNyZXQiOiAibmljZV90cnkiLAogICJxdW90YV9wcm9qZWN0X2lkIjogInRoaXNfaXNfZmFrZSIsCiAgInJlZnJlc2hfdG9rZW4iOiAibWF5YmVfbmV4dF90aW1lIiwKICAidHlwZSI6ICJhdXRob3JpemVkX3VzZXIiCn0K",
		"gcrurl: aHR0cHM6Ly9nY3IuaW8=",
	}

	rrout := strings.TrimSpace(rr.Stdout.String())
	for _, exp := range expected {
		re := regexp.MustCompile(fmt.Sprintf(".*%s.*", exp))
		secret := re.FindString(rrout)
		if secret == "" {
			t.Errorf("Did not find expected secret: '%s'", secret)
		}
	}
}
