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
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 300 * time.Second
)

var (
	schema = map[string]update.Item{
		".github/workflows/build.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/master.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/pr.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/docs.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/time-to-k8s.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/leaderboard.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/translations.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/update-k8s-versions.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/update-golang-version.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/update-golint-version.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},

		".github/workflows/time-to-k8s-public-chart.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/functional_verified.yml": {
			Replace: map[string]string{
				`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
			},
		},
		"go.mod": {
			Replace: map[string]string{
				`(?m)^go .*`: `go {{.StableVersionMM}}`,
			},
		},
		"Makefile": {
			Replace: map[string]string{
				// searching for 1.* so it does NOT match "KVM_GO_VERSION ?= $(GO_VERSION:.0=)" in the Makefile
				`GO_VERSION \?= 1.*`:             `GO_VERSION ?= {{.StableVersion}}`,
				`GO_K8S_VERSION_PREFIX \?= v1.*`: `GO_K8S_VERSION_PREFIX ?= {{.K8SVersion}}`,
			},
		},
		"hack/jenkins/installers/check_install_golang.sh": {
			Replace: map[string]string{
				`VERSION_TO_INSTALL=.*`: `VERSION_TO_INSTALL={{.StableVersion}}`,
			},
		},
	}

	// PR data
	prBranchPrefix = "update-golang-version_" // will be appended with first 7 characters of the PR commit SHA
	prTitle        = `update_golang_version: {stable: "{{.StableVersion}}"}`
	prIssue        = 9264
)

// Data holds stable Golang version - in full and in <major>.<minor> format
type Data struct {
	StableVersion   string `json:"stableVersion"`
	StableVersionMM string `json:"stableVersionMM"` // go.mod wants go version in <major>.<minor> format
	K8SVersion      string `json:"k8sVersion"`      // as of v1.23.0 Kubernetes uses k8s version in golang image name because: https://github.com/kubernetes/kubernetes/pull/103692#issuecomment-908659826

}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Golang stable version
	stable, stableMM, k8sVersion, err := goVersions()
	if err != nil || stable == "" || stableMM == "" {
		klog.Fatalf("Unable to get Golang stable version: %v", err)
	}
	data := Data{StableVersion: stable, StableVersionMM: stableMM, K8SVersion: k8sVersion}
	klog.Infof("Golang stable version: %s", data.StableVersion)

	update.Apply(ctx, schema, data, prBranchPrefix, prTitle, prIssue)
}

// goVersion returns Golang stable version.
func goVersions() (stable, stableMM, k8sVersion string, err error) {
	// will update to the same image that kubernetes project uses
	resp, err := http.Get("https://raw.githubusercontent.com/kubernetes/kubernetes/master/build/build-image/cross/VERSION")
	if err != nil {
		return "", "", "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}
	// example response: v1.23.0-go1.17-buster.0
	stable = string(body)
	k8sVersion = strings.Split(stable, "-")[0]
	stable = strings.Split(stable, "-")[1]
	stable = strings.Replace(stable, "go", "", 1)
	mmp := strings.SplitN(stable, ".", 3)
	stableMM = strings.Join(mmp[0:2], ".") // <major>.<minor> version
	return stable, stableMM, k8sVersion, nil
}
