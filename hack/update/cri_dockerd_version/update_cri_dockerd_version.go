/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

var (
	schema = map[string]update.Item{
		".github/workflows/master.yml": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="v{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=".*"`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
		".github/workflows/pr.yml": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="v{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=".*"`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
		"hack/jenkins/linux_integration_tests_none.sh": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=".*"`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
		"deploy/iso/minikube-iso/arch/aarch64/package/cri-dockerd-aarch64/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_AARCH64_VER = .*`:     `CRI_DOCKERD_AARCH64_VER = {{.Version}}`,
				`CRI_DOCKERD_AARCH64_VERSION = .*`: `CRI_DOCKERD_AARCH64_VERSION = {{.FullCommit}}`,
				`CRI_DOCKERD_AARCH64_REV = .*`:     `CRI_DOCKERD_AARCH64_REV = {{.ShortCommit}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/cri-dockerd/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_VER = .*`:     `CRI_DOCKERD_VER = {{.Version}}`,
				`CRI_DOCKERD_VERSION = .*`: `CRI_DOCKERD_VERSION = {{.FullCommit}}`,
				`CRI_DOCKERD_REV = .*`:     `CRI_DOCKERD_REV = {{.ShortCommit}}`,
			},
		},
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=.*`: `CRI_DOCKERD_VERSION="v{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=.*`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
	}
)

// Data holds stable cri-dockerd version in semver format.
type Data struct {
	Version     string
	FullCommit  string
	ShortCommit string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "Mirantis", "cri-dockerd")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	version := strings.TrimPrefix(stable.Tag, "v")

	data := Data{Version: version, FullCommit: stable.Commit, ShortCommit: stable.Commit[:7]}

	update.Apply(schema, data)

	if err := updateHashFiles(stable.Commit); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}
}

func updateHashFiles(commit string) error {
	r, err := http.Get(fmt.Sprintf("https://github.com/Mirantis/cri-dockerd/archive/%s.tar.gz", commit))
	if err != nil {
		return fmt.Errorf("failed to download source code: %v", err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	sum := sha256.Sum256(b)
	filePathBase := "../../../deploy/iso/minikube-iso/arch/"
	if err := updateHashFile(filePathBase+"aarch64/package/cri-dockerd-aarch64/cri-dockerd.hash", commit, sum); err != nil {
		return fmt.Errorf("aarch64: %v", err)
	}
	if err := updateHashFile(filePathBase+"x86_64/package/cri-dockerd/cri-dockerd.hash", commit, sum); err != nil {
		return fmt.Errorf("x86_64: %v", err)
	}
	return nil
}

func updateHashFile(filePath, commit string, shaSum [sha256.Size]byte) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read hash file: %v", err)
	}
	if strings.Contains(string(b), commit) {
		klog.Infof("hash file already contains %q", commit)
		return nil
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open hash file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("sha256 %x %s.tar.gz\n", shaSum, commit)); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
