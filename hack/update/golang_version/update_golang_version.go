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
	"io/ioutil"
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
		".github/workflows/iso.yml": {
			Replace: map[string]string{
				`go-version: '.*`: `go-version: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/kic_image.yml": {
			Replace: map[string]string{
				`go-version: '.*`: `go-version: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/master.yml": {
			Replace: map[string]string{
				`go-version: '.*`: `go-version: '{{.StableVersion}}'`,
			},
		},
		".github/workflows/pr.yml": {
			Replace: map[string]string{
				`go-version: '.*`: `go-version: '{{.StableVersion}}'`,
			},
		},
		".travis.yml": {
			Replace: map[string]string{
				`go:\n  - .*`: `go:{{printf "\n  - %s" .StableVersion}}`,
				`go: .*`:      `go: {{.StableVersion}}`,
			},
		},
		"go.mod": {
			Replace: map[string]string{
				`(?m)^go .*`: `go {{.StableVersionMM}}`,
			},
		},
		"hack/jenkins/common.sh": {
			Replace: map[string]string{
				`sudo \.\/installers\/check_install_golang\.sh \".*\" \"\/usr\/local\"`: `sudo ./installers/check_install_golang.sh "{{.StableVersion}}" "/usr/local"`,
			},
		},
		"Makefile": {
			Replace: map[string]string{
				`GO_VERSION \?= .*`: `GO_VERSION ?= {{.StableVersion}}`,
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
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Golang stable version
	stable, stableMM, err := goVersions()
	if err != nil || stable == "" || stableMM == "" {
		klog.Fatalf("Unable to get Golang stable version: %v", err)
	}
	data := Data{StableVersion: stable, StableVersionMM: stableMM}
	klog.Infof("Golang stable version: %s", data.StableVersion)

	update.Apply(ctx, schema, data, prBranchPrefix, prTitle, prIssue)
}

// goVersion returns Golang stable version.
func goVersions() (stable, stableMM string, err error) {
	resp, err := http.Get("https://golang.org/VERSION?m=text")
	if err != nil {
		return "", "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	stable = strings.TrimPrefix(string(body), "go")
	mmp := strings.SplitN(stable, ".", 3)
	stableMM = strings.Join(mmp[0:2], ".") // <major>.<minor> version
	return stable, stableMM, nil
}
