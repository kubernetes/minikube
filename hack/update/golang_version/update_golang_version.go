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
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var (
	workflowReplace = update.Item{
		Replace: map[string]string{
			`GO_VERSION: .*`: `GO_VERSION: '{{.StableVersion}}'`,
		},
	}

	schema = map[string]update.Item{
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
		"hack/jenkins/installers/check_install_golang.ps1": {
			Replace: map[string]string{
				`GoVersion = ".*"`: `GoVersion = "{{.StableVersion}}"`,
			},
		},
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`golang:[0-9.]+`: `golang:{{.StableVersion}}`,
			},
		},
		"deploy/addons/auto-pause/Dockerfile": {
			Replace: map[string]string{
				`golang:.* AS`: `golang:{{.StableVersion}} AS`,
			},
		},
		"deploy/gvisor/Dockerfile": {
			Replace: map[string]string{
				`golang:.* AS`: `golang:{{.StableVersion}} AS`,
			},
		},
	}
)

// Data holds stable Golang version - in full and in <major>.<minor> format
type Data struct {
	StableVersion string
	K8SVersion    string // as of v1.23.0 Kubernetes uses k8s version in golang image name because: https://github.com/kubernetes/kubernetes/pull/103692#issuecomment-908659826

}

func main() {
	addGitHubWorkflowFiles()

	// get Golang stable version
	stable, k8sVersion, err := goVersions()
	if err != nil || stable == "" {
		klog.Fatalf("Unable to get Golang stable version: %v", err)
	}
	// skip rc versions
	if strings.Contains(stable, "rc") {
		klog.Warningf("Golang stable version is a release candidate, skipping: %s", stable)
		return
	}
	data := Data{StableVersion: stable, K8SVersion: k8sVersion}
	klog.Infof("Golang stable version: %s", data.StableVersion)

	update.Apply(schema, data)

	if err := updateGoHashFile(stable); err != nil {
		klog.Fatalf("failed to update go hash file: %v", err)
	}
}

// goVersions returns Golang stable version.
func goVersions() (stable, k8sVersion string, err error) {
	// will update to the same image that kubernetes project uses
	resp, err := http.Get("https://raw.githubusercontent.com/kubernetes/kubernetes/master/build/build-image/cross/VERSION")
	if err != nil {
		return "", "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	// example response: v1.23.0-go1.17-buster.0
	stable = string(body)
	k8sVersion = strings.Split(stable, "-")[0]
	stable = strings.Split(stable, "-")[1]
	stable = strings.Replace(stable, "go", "", 1)
	return stable, k8sVersion, nil
}

func updateGoHashFile(version string) error {
	hashFilePath := "../../../deploy/iso/minikube-iso/go.hash"
	b, err := os.ReadFile(hashFilePath)
	if err != nil {
		return fmt.Errorf("failed to read hash file: %v", err)
	}
	if strings.Contains(string(b), version) {
		klog.Infof("hash file already contains %q", version)
		return nil
	}
	r, err := http.Get(fmt.Sprintf("https://dl.google.com/go/go%s.src.tar.gz.sha256", version))
	if err != nil {
		return fmt.Errorf("failed to download golang sha256 file: %v", err)
	}
	defer r.Body.Close()
	sha, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	f, err := os.OpenFile(hashFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open go.hash file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("sha256  %s  go%s.src.tar.gz\n", sha, version)); err != nil {
		return fmt.Errorf("failed to write to go.hash file: %v", err)
	}
	return nil
}

func addGitHubWorkflowFiles() {
	files, err := os.ReadDir("../../../.github/workflows")
	if err != nil {
		klog.Fatalf("failed to read workflows dir: %v", err)
	}
	for _, f := range files {
		filename := ".github/workflows/" + f.Name()
		schema[filename] = workflowReplace
	}
}
