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
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

// TestHelmInstall integration test verifies helm installation, upgrade, and no-change behavior inside a live guest VM.
// We test this flow because installing the latest helm allows self-healing and updating the cluster when Helm is missing, outdated, or broken.
func TestHelmInstall(t *testing.T) {
	MaybeParallel(t)

	profile := UniqueProfileName("helm")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	startArgs := append([]string{"start", "-p", profile, "--memory=2048", "--alsologtostderr", "-v=1"}, StartArgs()...)
	_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed to start minikube: %v", err)
	}

	cc, err := config.Load(profile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	api, err := machine.NewAPIClient(nil)
	if err != nil {
		t.Fatalf("failed to create api client: %v", err)
	}
	defer api.Close()

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		t.Fatalf("failed to get control plane node: %v", err)
	}

	host, err := machine.LoadHost(api, config.MachineName(*cc, cp))
	if err != nil {
		t.Fatalf("failed to load host: %v", err)
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		t.Fatalf("failed to get command runner: %v", err)
	}

	// Backup the original helm binaries if they exist in the guest VM/container to keep tests isolated.
	_, errHasHelm := runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
	if errHasHelm == nil {
		if _, err := runner.RunCmd(exec.Command("sudo", "mv", "/usr/bin/helm", "/usr/bin/helm.bak")); err != nil {
			t.Fatalf("failed to backup original /usr/bin/helm: %v", err)
		}
	}
	_, errHasLocalHelm := runner.RunCmd(exec.Command("test", "-f", "/usr/local/bin/helm"))
	if errHasLocalHelm == nil {
		if _, err := runner.RunCmd(exec.Command("sudo", "mv", "/usr/local/bin/helm", "/usr/local/bin/helm.bak")); err != nil {
			t.Fatalf("failed to backup original /usr/local/bin/helm: %v", err)
		}
	}

	t.Cleanup(func() {
		// Clean up any test files we created and restore original helm binaries
		runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm", "/usr/local/bin/helm"))
		if errHasHelm == nil {
			runner.RunCmd(exec.Command("sudo", "mv", "/usr/bin/helm.bak", "/usr/bin/helm"))
		}
		if errHasLocalHelm == nil {
			runner.RunCmd(exec.Command("sudo", "mv", "/usr/local/bin/helm.bak", "/usr/local/bin/helm"))
		}
	})

	// 1. Install test
	// This test ensures that if helm is not installed at /usr/bin/helm,
	// InstallHelm will successfully download and install it.
	t.Run("Install", func(t *testing.T) {
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm", "/usr/local/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete helm inside guest: %v", err)
		}

		err = addons.InstallHelm(nil, runner)
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		rr, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version"))
		if err != nil {
			t.Errorf("helm binary at /usr/bin/helm is not executable or failed: %v", err)
		}
		if !strings.Contains(rr.Stdout.String(), "Version") {
			t.Errorf("unexpected helm version output: %s", rr.Stdout.String())
		}
	})

	// 2. Upgrade test
	// This test ensures that if /usr/bin/helm is missing, but an older version of Helm is
	// installed at /usr/local/bin/helm, InstallHelm will correctly download and install
	// the latest version of Helm directly to /usr/bin/helm.
	t.Run("Upgrade", func(t *testing.T) {
		// Ensure /usr/bin/helm is deleted so the installer triggers
		_, err := runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/bin/helm"))
		if err != nil {
			t.Fatalf("failed to delete /usr/bin/helm: %v", err)
		}

		// Install an older Helm version (e.g. v3.12.0) to /usr/local/bin/helm
		script := `
			curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
			chmod 700 get_helm.sh
			HELM_INSTALL_DIR=/usr/local/bin DESIRED_VERSION=v3.12.0 ./get_helm.sh
		`
		_, err = runner.RunCmd(exec.Command("sudo", "bash", "-o", "errexit", "-c", script))
		if err != nil {
			t.Fatalf("failed to install older helm: %v", err)
		}

		// Verify the older helm was installed and is executable
		rrOld, err := runner.RunCmd(exec.Command("/usr/local/bin/helm", "version"))
		if err != nil {
			t.Fatalf("older helm at /usr/local/bin/helm version check failed: %v", err)
		}
		if !strings.Contains(rrOld.Stdout.String(), "v3.12.0") {
			t.Fatalf("older helm version mismatch: stdout: %s", rrOld.Stdout.String())
		}

		// Run InstallHelm, which should download and install the latest helm into /usr/bin/helm
		err = addons.InstallHelm(nil, runner)
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Verify that a newer version of helm has been installed in /usr/bin/helm
		rrNew, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version"))
		if err != nil {
			t.Errorf("helm binary not found or failed at /usr/bin/helm after upgrade: %v", err)
		}
		if strings.Contains(rrNew.Stdout.String(), "v3.12.0") {
			t.Errorf("helm was not successfully upgraded/replaced: stdout: %s", rrNew.Stdout.String())
		}

		// Clean up the temporary older helm
		_, err = runner.RunCmd(exec.Command("sudo", "rm", "-f", "/usr/local/bin/helm"))
		if err != nil {
			t.Logf("warning: failed to delete temporary older helm: %v", err)
		}
	})

	// 3. No Change test
	// This test verifies that if Helm is already installed in /usr/bin/helm,
	// running InstallHelm again is a no-op and does not modify or reinstall it.
	t.Run("NoChange", func(t *testing.T) {
		// Run InstallHelm to ensure latest helm is present
		err = addons.InstallHelm(nil, runner)
		if err != nil {
			t.Fatalf("InstallHelm failed: %v", err)
		}

		// Retrieve current helm version details
		rrFirst, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version"))
		if err != nil {
			t.Fatalf("failed to check helm version: %v", err)
		}
		firstVersion := rrFirst.Stdout.String()

		// Run InstallHelm again
		err = addons.InstallHelm(nil, runner)
		if err != nil {
			t.Fatalf("InstallHelm second call failed: %v", err)
		}

		// Retrieve helm version details again
		rrSecond, err := runner.RunCmd(exec.Command("/usr/bin/helm", "version"))
		if err != nil {
			t.Fatalf("failed to check helm version after second call: %v", err)
		}
		secondVersion := rrSecond.Stdout.String()

		// Verify version details did not change
		if firstVersion != secondVersion {
			t.Errorf("helm version changed after second InstallHelm call:\nfirst:  %s\nsecond: %s", firstVersion, secondVersion)
		}
	})
}
