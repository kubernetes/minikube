/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

// Package addons provides helpers to install and verify Helm-based addons in minikube.
package addons

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// runs a helm install within the minikube vm or container based on the contents of chart *assets.HelmChart
func installHelmChart(ctx context.Context, chart *assets.HelmChart) *exec.Cmd {
	args := []string{
		fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"helm", "upgrade", "--install", chart.Name, chart.Repo, "--create-namespace",
	}
	if chart.Namespace != "" {
		args = append(args, "--namespace", chart.Namespace)
	}

	if chart.Values != nil {
		for _, value := range chart.Values {
			args = append(args, "--set", value)
		}
	}

	if chart.ValueFiles != nil {
		for _, value := range chart.ValueFiles {
			args = append(args, "--values", value)
		}
	}

	return exec.CommandContext(ctx, "sudo", args...)
}

// runs a helm uninstall based on the contents of chart *assets.HelmChart
func uninstalllHelmChart(ctx context.Context, chart *assets.HelmChart) *exec.Cmd {
	args := []string{
		fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"helm", "uninstall", chart.Name,
	}
	if chart.Namespace != "" {
		args = append(args, "--namespace", chart.Namespace)
	}
	return exec.CommandContext(ctx, "sudo", args...)
}

// based on enable will execute installHelmChart or uninstallHelmChart
func helmUninstallOrInstall(ctx context.Context, chart *assets.HelmChart, enable bool) *exec.Cmd {
	if enable {
		return installHelmChart(ctx, chart)
	}
	return uninstalllHelmChart(ctx, chart)
}

// InstallHelmVersion downloads and installs a specific version of Helm to the target directory inside the guest VM/container.
// If version is "latest" or empty, it installs the latest stable version of Helm.
func InstallHelmVersion(runner command.Runner, version string, installDir string) error {
	if version == "" {
		version = "latest"
	}
	if installDir == "" {
		installDir = "/usr/bin"
	}

	var env string
	if version != "latest" {
		env = fmt.Sprintf("HELM_INSTALL_DIR=%s DESIRED_VERSION=%s", installDir, version)
	} else {
		env = fmt.Sprintf("HELM_INSTALL_DIR=%s", installDir)
	}

	script := fmt.Sprintf(`
			curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
			chmod 700 get_helm.sh
			%s ./get_helm.sh
		`, env)

	_, err := runner.RunCmd(exec.Command("sudo", "bash", "-o", "errexit", "-c", script))
	if err != nil {
		return fmt.Errorf("downloading and installing helm version %s to %s: %w", version, installDir, err)
	}
	return nil
}

// InstallHelm checks if helm is installed at /usr/bin/helm.
// Use /usr/bin/helm which is always in the PATH, unlike /usr/local/bin.
func InstallHelm(_ *assets.Addon, runner command.Runner, version string) error {
	_, err := runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
	if err != nil {
		return InstallHelmVersion(runner, version, "/usr/bin")
	}
	return nil
}
