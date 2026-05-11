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

	"k8s.io/minikube/pkg/minikube/constants"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute

	// templateDirBase is the base template directory for kubeadm testdata files (k8s < v1.36)
	templateDirBase = "update/kubernetes_version/templates/v1beta4"
	// templateDirV136Plus is the template directory for k8s >= v1.36 (adds ExtendWebSocketsToKubelet=false feature-gate for cri-dockerd)
	templateDirV136Plus = "update/kubernetes_version/templates/v136plus"
)

var (
	// schemaBase holds schema items that are independent of the Kubernetes version.
	schemaBase = map[string]update.Item{
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
	}
)

// dockerShimTemplateDir returns the template directory to use for cri-dockerd testdata files
// based on the given Kubernetes minor version. For k8s >= v1.36, a separate template directory
// is used that includes the ExtendWebSocketsToKubelet=false feature-gate workaround.
// versionMM should be in major.minor format (e.g. "v1.36"), but full semver is also accepted.
func dockerShimTemplateDir(versionMM string) string {
	if semver.Compare(semver.MajorMinor(versionMM), "v1.36") >= 0 {
		return templateDirV136Plus
	}
	return templateDirBase
}

// testdataSchemaItems returns schema items for kubeadm testdata files for the given version.
// versionMMTmplKey and versionP0TmplKey are the template placeholder names (e.g. "LatestVersionMM"),
// and versionMM is the actual minor version value used to select the right template directory.
func testdataSchemaItems(versionMMTmplKey, versionP0TmplKey, versionMM string) map[string]update.Item {
	dir := dockerShimTemplateDir(versionMM)
	p0Replace := map[string]string{
		`kubernetesVersion:.*`: `kubernetesVersion: {{.` + versionP0TmplKey + `}}`,
	}
	prefix := "pkg/minikube/bootstrapper/bsutil/testdata/{{." + versionMMTmplKey + "}}/"
	return map[string]update.Item{
		prefix + "containerd-api-port.yaml": {
			Content: update.Loadf(templateDirBase + "/containerd-api-port.yaml"),
			Replace: p0Replace,
		},
		prefix + "containerd-pod-network-cidr.yaml": {
			Content: update.Loadf(templateDirBase + "/containerd-pod-network-cidr.yaml"),
			Replace: p0Replace,
		},
		prefix + "containerd.yaml": {
			Content: update.Loadf(templateDirBase + "/containerd.yaml"),
			Replace: p0Replace,
		},
		prefix + "crio-options-gates.yaml": {
			Content: update.Loadf(templateDirBase + "/crio-options-gates.yaml"),
			Replace: p0Replace,
		},
		prefix + "crio.yaml": {
			Content: update.Loadf(templateDirBase + "/crio.yaml"),
			Replace: p0Replace,
		},
		prefix + "default.yaml": {
			Content: update.Loadf(dir + "/default.yaml"),
			Replace: p0Replace,
		},
		prefix + "dns.yaml": {
			Content: update.Loadf(dir + "/dns.yaml"),
			Replace: p0Replace,
		},
		prefix + "image-repository.yaml": {
			Content: update.Loadf(dir + "/image-repository.yaml"),
			Replace: p0Replace,
		},
		prefix + "options.yaml": {
			Content: update.Loadf(dir + "/options.yaml"),
			Replace: p0Replace,
		},
	}
}

// Data holds greatest current stable release and greatest latest rc or beta pre-release Kubernetes versions
type Data struct {
	StableVersion   string
	LatestVersion   string
	LatestVersionMM string // LatestVersion in <major>.<minor> format
	// for testdata: if StableVersion greater than 'LatestVersionMM.0' exists, LatestVersionP0 is 'LatestVersionMM.0', otherwise LatestVersionP0 is LatestVersion.
	LatestVersionP0  string
	CurrentVersionMM string
	CurrentVersionP0 string
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

	// Build schema: start with version-independent items and add testdata items using the
	// appropriate template directory based on the latest Kubernetes minor version.
	schema := make(map[string]update.Item)
	for k, v := range schemaBase {
		schema[k] = v
	}
	for k, v := range testdataSchemaItems("LatestVersionMM", "LatestVersionP0", latestMM) {
		schema[k] = v
	}

	updateCurrentTestDataFiles(latestMM, &data, schema)

	// Print PR title for GitHub action.
	fmt.Printf("Bump Kubernetes version default: %s and latest: %s\n", data.StableVersion, data.LatestVersion)

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
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

// updateCurrentTestDataFiles the point of this function it to update the testdata files for the current `NewestKubernetesVersion`
// otherwise, we can run into an issue where there's a new Kubernetes minor version, so the existing testdata files are ignored
// but they're not ending in `.0` and are potentially ending with a prerelease tag such as `v1.29.0-rc.2`, resulting in the unit
// tests failing and requiring manual intervention to correct the testdata files.
func updateCurrentTestDataFiles(latestMM string, data *Data, schema map[string]update.Item) {
	currentMM := semver.MajorMinor(constants.NewestKubernetesVersion)

	// if we're still on the same version, skip and we're already going to update these testdata files
	if currentMM == latestMM {
		return
	}

	data.CurrentVersionMM = currentMM
	data.CurrentVersionP0 = currentMM + ".0"

	for k, v := range testdataSchemaItems("CurrentVersionMM", "CurrentVersionP0", currentMM) {
		schema[k] = v
	}
}
