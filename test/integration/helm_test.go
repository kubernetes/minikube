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
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"
	"k8s.io/minikube/pkg/minikube/run"
)

// minExpectedHelmVersion is the minimum helm version we expect to be
// installed. We check ">= minimum" rather than "== latest" because CDN
// propagation delays during Helm releases can make the install script and
// our version query see different versions, causing flaky failures.
// See https://github.com/kubernetes/minikube/issues/23323
//
// Bump this occasionally to catch the CDN serving something stale or
// broken. It doesn't need to track every Helm release.
var minExpectedHelmVersion = semver.Version{Major: 3, Minor: 20}

// TestHelmInstall verifies that InstallHelm can install, upgrade, and
// re-install helm inside a live node.
func TestHelmInstall(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("helm")
	// TODO: update the timeout based on data from recent runs (3*p95 once data is available)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	t.Cleanup(func() { CleanupWithLogs(t, profile, cancel) })

	t.Logf("starting minikube profile %s", profile)
	startArgs := append([]string{"start", "-p", profile, "--memory=3072", "--alsologtostderr"}, StartArgs()...)
	_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: %v", err)
	}

	runner := runnerForProfile(t, profile)

	// Skip if the node has no internet — helm install downloads from
	// get.helm.sh which requires outbound connectivity.
	t.Log("checking network connectivity")
	_, err = runner.RunCmd(exec.Command("curl", "-fsSL", "--max-time", "10", "-o", "/dev/null", "https://get.helm.sh/helm3-latest-version"))
	if err != nil {
		t.Skip("skipping: guest VM/container has no outbound internet access (this is required to download helm): https://github.com/kubernetes/minikube/issues/23275")
	}

	// Verify that InstallHelm installs helm when it is not installed. This
	// happens when enabling the first helm-based addon. Since /usr/bin is not
	// persisted, this happens again after restarting a stopped cluster.
	t.Run("Install", func(t *testing.T) {
		t.Log("removing helm to simulate missing binary")
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Fatalf("failed to remove helm: %v", err)
		}

		t.Log("checking helm version with missing binary")
		_, err = addons.HelmVersion(runner)
		if !errors.Is(err, addons.ErrHelmNotInstalled) {
			t.Fatalf("expected ErrHelmNotInstalled, got: %v", err)
		}
		t.Logf("helm version error: %v", err)

		t.Log("installing helm latest version")
		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatal(err)
		}

		version, err := addons.HelmVersion(runner)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("installed helm version: %s", version)
		if version.LT(minExpectedHelmVersion) {
			t.Fatalf("installed helm version %q is older than minimum expected %q", version, minExpectedHelmVersion)
		}
	})

	// Verify that InstallHelm upgrades an older helm to the latest
	// version. This functionality is not used yet by minikube.
	t.Run("Upgrade", func(t *testing.T) {
		t.Logf("installing helm %s", minExpectedHelmVersion)
		err := addons.InstallHelm(runner, addons.HelmOptions{Version: &minExpectedHelmVersion})
		if err != nil {
			t.Fatal(err)
		}

		t.Log("checking installed helm version")
		versionOld, err := addons.HelmVersion(runner)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("installed helm version: %s", versionOld)
		if versionOld.NE(minExpectedHelmVersion) {
			t.Fatalf("helm version mismatch: expected %q, got %q", minExpectedHelmVersion, versionOld)
		}

		t.Log("upgrading helm to latest version")
		err = addons.InstallHelm(runner, addons.HelmOptions{})
		if err != nil {
			t.Fatal(err)
		}

		t.Log("checking upgraded helm version")
		versionNew, err := addons.HelmVersion(runner)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("upgraded helm version: %s (from %s)", versionNew, versionOld)
		if versionNew.LTE(minExpectedHelmVersion) {
			t.Fatalf("upgraded version %q not newer than older version %q", versionNew, minExpectedHelmVersion)
		}
	})

	// Verify that InstallHelm with a pinned version is idempotent — running
	// it twice does not modify the binary. This is not used yet by minikube.
	t.Run("NoChange", func(t *testing.T) {
		t.Logf("installing helm %s", minExpectedHelmVersion)
		err := addons.InstallHelm(runner, addons.HelmOptions{Version: &minExpectedHelmVersion})
		if err != nil {
			t.Fatal(err)
		}

		t.Log("checking first helm version")
		firstVersion, err := addons.HelmVersion(runner)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("first helm version: %s", firstVersion)

		t.Logf("installing helm %s again", minExpectedHelmVersion)
		err = addons.InstallHelm(runner, addons.HelmOptions{Version: &minExpectedHelmVersion})
		if err != nil {
			t.Fatal(err)
		}

		t.Log("checking second helm version")
		secondVersion, err := addons.HelmVersion(runner)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("second helm version: %s", secondVersion)
		if firstVersion.NE(secondVersion) {
			t.Fatalf("helm version changed: first %q, second %q", firstVersion, secondVersion)
		}
	})
}

// runnerForProfile returns a command runner for the control plane node of the given profile.
func runnerForProfile(t *testing.T, profile string) command.Runner {
	t.Helper()

	cc, err := config.Load(profile)
	if err != nil {
		t.Fatalf("failed to load config for %s: %v", profile, err)
	}

	api, err := machine.NewAPIClient(&run.CommandOptions{NonInteractive: true})
	if err != nil {
		t.Fatalf("failed to create api client: %v", err)
	}
	t.Cleanup(func() { api.Close() })

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		t.Fatalf("failed to get control plane node for %s: %v", profile, err)
	}

	host, err := machine.LoadHost(api, config.MachineName(*cc, cp))
	if err != nil {
		t.Fatalf("failed to load host for %s: %v", profile, err)
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		t.Fatalf("failed to get command runner for %s: %v", profile, err)
	}

	return runner
}
