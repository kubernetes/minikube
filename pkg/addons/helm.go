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

package addons

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// ParsedImageRepo holds the parsed components of an image repository string
type ParsedImageRepo struct {
	Registry string
	Org      string
	Tag      string
}

// parseImageRepository parses an image repository string in format "registry/org[:tag]"
// Examples:
//   - ghcr.io/volcano-sh:latest -> {ghcr.io, volcano-sh, latest}
//   - docker.io/volcanosh:v1.12.0 -> {docker.io, volcanosh, v1.12.0}
//   - my-mirror.cn/volcanosh -> {my-mirror.cn, volcanosh, ""} (tag defaults to "latest")
//
// Returns an error if the format is invalid (missing org part).
func parseImageRepository(imageRepo string) (*ParsedImageRepo, error) {
	if imageRepo == "" {
		return nil, nil
	}

	result := &ParsedImageRepo{Tag: "latest"}

	if idx := strings.LastIndex(imageRepo, ":"); idx > strings.LastIndex(imageRepo, "/") {
		result.Tag = imageRepo[idx+1:]
		imageRepo = imageRepo[:idx]
	}

	parts := strings.SplitN(imageRepo, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fmt.Errorf("invalid image repository format %q: must be registry/org[:tag] (e.g., ghcr.io/volcano-sh:latest)", imageRepo)
	}

	result.Registry = parts[0]
	result.Org = parts[1]

	return result, nil
}

// runs a helm install within the minikube vm or container based on the contents of chart *assets.HelmChart
func installHelmChart(ctx context.Context, chart *assets.HelmChart, imageRepository string) *exec.Cmd {
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

	// Parse image repository and apply settings
	if imageRepository != "" && chart.ImageNameKeys != nil {
		parsed, err := parseImageRepository(imageRepository)
		if err == nil && parsed != nil {
			// Set registry
			if chart.ImageSetKey != "" {
				args = append(args, "--set", fmt.Sprintf("%s=%s", chart.ImageSetKey, parsed.Registry))
			}

			// Set image names with org prefix
			for imageName, helmKey := range chart.ImageNameKeys {
				args = append(args, "--set", fmt.Sprintf("%s=%s/%s", helmKey, parsed.Org, imageName))
			}

			// Set tag
			if chart.TagSetKey != "" && parsed.Tag != "" {
				args = append(args, "--set", fmt.Sprintf("%s=%s", chart.TagSetKey, parsed.Tag))
			}
		}
	} else if imageRepository != "" && chart.ImageSetKey != "" {
		// Fallback for charts without ImageNameKeys (simple registry override)
		args = append(args, "--set", fmt.Sprintf("%s=%s", chart.ImageSetKey, imageRepository))
	}

	return exec.CommandContext(ctx, "sudo", args...)
}

// helmRepoAdd adds a helm repository before chart installation
func helmRepoAdd(ctx context.Context, chart *assets.HelmChart) *exec.Cmd {
	repoName := chart.Repo
	if parts := strings.Split(chart.Repo, "/"); len(parts) > 0 {
		repoName = parts[0]
	}
	if chart.RepoURL == "" {
		return nil
	}
	args := []string{
		fmt.Sprintf("KUBECONFIG=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"helm", "repo", "add", repoName, chart.RepoURL, "--force-update",
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
func helmUninstallOrInstall(ctx context.Context, chart *assets.HelmChart, enable bool, imageRepository string) *exec.Cmd {
	if enable {
		return installHelmChart(ctx, chart, imageRepository)
	}
	return uninstalllHelmChart(ctx, chart)
}

func helmInstallBinary(addon *assets.Addon, runner command.Runner) error {
	_, err := runner.RunCmd(exec.Command("test", "-f", "/usr/bin/helm"))
	if err != nil {
		_, err = runner.RunCmd(exec.Command("test", "-d", "/usr/local/bin"))
		if err != nil {
			_, err = runner.RunCmd(exec.Command("sudo", "mkdir", "-p", "/usr/local/bin"))
			if err != nil {
				return fmt.Errorf("creating /usr/local/bin: %w", err)
			}
		}

		installCmd := "curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && chmod 700 get_helm.sh && ./get_helm.sh"
		_, err = runner.RunCmd(exec.Command("sudo", "bash", "-c", installCmd))
		if err != nil {
			return fmt.Errorf("downloading helm: %w", err)
		}
		// we copy the binary from /usr/local/bin to /usr/bin because /usr/local/bin is not in PATH in both iso and kicbase
		_, err = runner.RunCmd(exec.Command("sudo", "mv", "/usr/local/bin/helm", "/usr/bin/helm"))
		if err != nil {
			return fmt.Errorf("installing helm: %w", err)
		}
	}
	return err
}
