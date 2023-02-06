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

package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute
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
			Content: update.Loadf("templates/v1beta3/containerd-api-port.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/containerd-pod-network-cidr.yaml": {
			Content: update.Loadf("templates/v1beta3/containerd-pod-network-cidr.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/containerd.yaml": {
			Content: update.Loadf("templates/v1beta3/containerd.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/crio-options-gates.yaml": {
			Content: update.Loadf("templates/v1beta3/crio-options-gates.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/crio.yaml": {
			Content: update.Loadf("templates/v1beta3/crio.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/default.yaml": {
			Content: update.Loadf("templates/v1beta3/default.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/dns.yaml": {
			Content: update.Loadf("templates/v1beta3/dns.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/image-repository.yaml": {
			Content: update.Loadf("templates/v1beta3/image-repository.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
		"pkg/minikube/bootstrapper/bsutil/testdata/{{.LatestVersionMM}}/options.yaml": {
			Content: update.Loadf("templates/v1beta3/options.yaml"),
			Replace: map[string]string{
				`kubernetesVersion:.*`: `kubernetesVersion: {{.LatestVersionP0}}`,
			},
		},
	}
)

// Data holds greatest current stable release and greatest latest rc or beta pre-release Kubernetes versions
type Data struct {
	StableVersion   string
	LatestVersion   string
	LatestVersionMM string // LatestVersion in <major>.<minor> format
	// for testdata: if StableVersion greater than 'LatestVersionMM.0' exists, LatestVersionP0 is 'LatestVersionMM.0', otherwise LatestVersionP0 is LatestVersion.
	LatestVersionP0 string
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

	// Print PR title for GitHub action.
	fmt.Printf("Bump Kubernetes version default: %s and latest: %s\n", data.StableVersion, data.LatestVersion)

	update.Apply(schema, data)
}

// k8sVersions returns Kubernetes versions.
func k8sVersions(ctx context.Context, owner, repo string) (stable, latest, latestMM, latestP0 string, err error) {
	// get Kubernetes versions from GitHub Releases
	stableRls, latestRls, _, err := update.GHReleases(ctx, owner, repo)
	stable = stableRls.Tag
	latest = latestRls.Tag
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
