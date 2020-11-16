/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

/*
Script expects the following env variables:
 - UPDATE_TARGET=<string>: optional - if unset/absent, default option is "fs"; valid options are:
   - "fs"  - update only local filesystem repo files [default]
   - "gh"  - update only remote GitHub repo files and create PR (if one does not exist already)
   - "all" - update local and remote repo files and create PR (if one does not exist already)
 - GITHUB_TOKEN=<string>: GitHub [personal] access token
   - note: GITHUB_TOKEN is required if UPDATE_TARGET is "gh" or "all"
*/

package main

import (
	"context"
	"time"

	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 300 * time.Second
)

var (
	schema = map[string]update.Item{
		"pkg/minikube/constants/constants.go": {
			Replace: map[string]string{
				`DefaultKubernetesVersion = ".*`: `DefaultKubernetesVersion = "{{.StableVersion}}"`,
				`NewestKubernetesVersion = ".*`:  `NewestKubernetesVersion = "{{.LatestVersion}}"`,
			},
		},
		"site/content/en/docs/commands/start.md": {
			Replace: map[string]string{
				`'stable' for .*,`:  `'stable' for {{.StableVersion}},`,
				`'latest' for .*\)`: `'latest' for {{.LatestVersion}})`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/containerd-api-port.yaml": {
			Content: update.Loadf("templates/v1beta2/containerd-api-port.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/containerd-pod-network-cidr.yaml": {
			Content: update.Loadf("templates/v1beta2/containerd-pod-network-cidr.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/containerd.yaml": {
			Content: update.Loadf("templates/v1beta2/containerd.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/crio-options-gates.yaml": {
			Content: update.Loadf("templates/v1beta2/crio-options-gates.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/crio.yaml": {
			Content: update.Loadf("templates/v1beta2/crio.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/default.yaml": {
			Content: update.Loadf("templates/v1beta2/default.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/dns.yaml": {
			Content: update.Loadf("templates/v1beta2/dns.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/image-repository.yaml": {
			Content: update.Loadf("templates/v1beta2/image-repository.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/options.yaml": {
			Content: update.Loadf("templates/v1beta2/options.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
	}

	// PR data
	prBranchPrefix = "update-kubernetes-version_" // will be appended with first 7 characters of the PR commit SHA
	prTitle        = `update_kubernetes_version: {stable: "{{.StableVersion}}", latest: "{{.LatestVersion}}"}`
	prIssue        = 4392
)

// Data holds greatest current stable release and greatest latest rc or beta pre-release Kubernetes versions
type Data struct {
	StableVersion   string `json:"StableVersion"`
	LatestVersion   string `json:"LatestVersion"`
	LatestVersionMM string `json:"LatestVersionMM"` // LatestVersion in <major>.<minor> format
	// for testdata: if StableVersion greater than 'LatestVersionMM.0' exists, LatestVersionP0 is 'LatestVersionMM.0', otherwise LatestVersionP0 is LatestVersion.
	LatestVersionP0 string `json:"LatestVersionP0"`
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Kubernetes versions from GitHub Releases
	stable, latest, latestMM, latestP0, err := k8sVersions(ctx, "kubernetes", "kubernetes")
	if err != nil || !semver.IsValid(stable) || !semver.IsValid(latest) || !semver.IsValid(latestMM) || !semver.IsValid(latestP0) {
		klog.Fatalf("Unable to get Kubernetes versions: %v", err)
	}
	data := Data{StableVersion: stable, LatestVersion: latest, LatestVersionMM: latestMM, LatestVersionP0: latestP0}
	klog.Infof("Kubernetes versions: 'stable' is %s and 'latest' is %s", data.StableVersion, data.LatestVersion)

	update.Apply(ctx, schema, data, prBranchPrefix, prTitle, prIssue)
}

// k8sVersion returns Kubernetes versions.
func k8sVersions(ctx context.Context, owner, repo string) (stable, latest, latestMM, latestP0 string, err error) {
	// get Kubernetes versions from GitHub Releases
	stable, latest, err = update.GHReleases(ctx, owner, repo)
	if err != nil || !semver.IsValid(stable) || !semver.IsValid(latest) {
		return "", "", "", "", err
	}
	latestMM = semver.MajorMinor(latest)
	latestP0 = latestMM + ".0"
	if semver.Compare(stable, latestP0) == -1 {
		latestP0 = latest
	}
	return stable, latest, latestMM, latestP0, nil
}
