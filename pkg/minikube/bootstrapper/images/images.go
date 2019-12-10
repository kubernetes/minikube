/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// Package images implements helpers for getting image names
package images

import (
	"path"
	"runtime"
)

// ArchTag returns a CPU architecture suffix for images
func ArchTag(hasTag bool) string {
	if runtime.GOARCH == "amd64" && !hasTag {
		return ":"
	}
	return "-" + runtime.GOARCH + ":"
}

// Auxiliary returns images that are helpful for running minikube
func Auxiliary(mirror string) []string {
	return []string{
		addonManager(mirror),
		storageProvisioner(mirror),
		dashboardFrontend(mirror),
		dashboardMetrics(mirror),
	}
}

// Pause returns the image name to pull for the pause image
func Pause(mirror string) string {
	// Should match `PauseVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	return path.Join(KubernetesRepo(mirror), "pause"+ArchTag(false)+"3.1")
}

// storageProvisioner returns the minikube storage provisioner image
func storageProvisioner(mirror string) string {
	return path.Join(minikubeRepo(mirror), "storage-provisioner"+ArchTag(false)+"v1.8.1")
}

// addonManager returns the Kubernetes addon manager image
func addonManager(mirror string) string {
	return path.Join(KubernetesRepo(mirror), "kube-addon-manager"+ArchTag(false)+"v9.0.2")
}

// dashboardFrontend returns the image used for the dashboard frontend
func dashboardFrontend(repo string) string {
	if repo == "" {
		repo = "kubernetesui"
	}
	// See 'kubernetes-dashboard' in deploy/addons/dashboard/dashboard-dp.yaml
	return path.Join(repo, "dashboard:v2.0.0-beta8")
}

// dashboardMetrics returns the image used for the dashboard metrics scraper
func dashboardMetrics(repo string) string {
	if repo == "" {
		repo = "kubernetesui"
	}
	// See 'dashboard-metrics-scraper' in deploy/addons/dashboard/dashboard-dp.yaml
	return path.Join(repo, "metrics-scraper:v1.0.2")
}
