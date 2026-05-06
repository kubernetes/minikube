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

	"k8s.io/minikube/hack/update"

	"k8s.io/klog/v2"
)

var (
	schema = map[string]update.Item{
		".github/workflows/functional_test.yml": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="v{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=".*"`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
		"deploy/iso/minikube-iso/arch/aarch64/package/cri-dockerd-aarch64/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_AARCH64_VERSION = .*`: `CRI_DOCKERD_AARCH64_VERSION = {{.Version}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/cri-dockerd/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION = .*`: `CRI_DOCKERD_VERSION = {{.Version}}`,
			},
		},
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=.*`: `CRI_DOCKERD_VERSION="v{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=.*`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
		"hack/jenkins/linux_integration_tests_none.sh": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="{{.Version}}"`,
				`CRI_DOCKERD_COMMIT=".*"`:  `CRI_DOCKERD_COMMIT="{{.FullCommit}}"`,
			},
		},
	}
)

type Data struct {
	Version    string
	FullCommit string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "Mirantis", "cri-dockerd")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	version := strings.TrimPrefix(stable.Tag, "v")
	data := Data{Version: version, FullCommit: stable.Commit}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}

	if err := updateHashFiles(version); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}

	if err := updateServiceFiles(version); err != nil {
		klog.Fatalf("failed to update systemd service files: %v", err)
	}
}

func updateHashFiles(version string) error {
	filePathBase := "../deploy/iso/minikube-iso/arch/"

	// amd64
	amd64URL := fmt.Sprintf("https://github.com/Mirantis/cri-dockerd/releases/download/v%s/cri-dockerd-%s.amd64.tgz", version, version)
	sumAmd64, err := getURLSHA256(amd64URL)
	if err != nil {
		return err
	}
	if err := updateHashFile(filePathBase+"x86_64/package/cri-dockerd/cri-dockerd.hash", fmt.Sprintf("cri-dockerd-%s.amd64.tgz", version), sumAmd64); err != nil {
		return err
	}

	// arm64
	arm64URL := fmt.Sprintf("https://github.com/Mirantis/cri-dockerd/releases/download/v%s/cri-dockerd-%s.arm64.tgz", version, version)
	sumArm64, err := getURLSHA256(arm64URL)
	if err != nil {
		return err
	}
	if err := updateHashFile(filePathBase+"aarch64/package/cri-dockerd-aarch64/cri-dockerd.hash", fmt.Sprintf("cri-dockerd-%s.arm64.tgz", version), sumArm64); err != nil {
		return err
	}

	return nil
}

func getURLSHA256(url string) ([32]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to read response body: %v", err)
	}
	return sha256.Sum256(b), nil
}

func updateHashFile(filePath, filename string, shaSum [sha256.Size]byte) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read hash file: %v", err)
	}
	if strings.Contains(string(b), filename) {
		klog.Infof("hash file already contains %q", filename)
		return nil
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open hash file: %v", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "sha256 %x %s\n", shaSum, filename); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}

func updateServiceFiles(version string) error {
	serviceURL := fmt.Sprintf("https://raw.githubusercontent.com/Mirantis/cri-dockerd/v%s/packaging/systemd/cri-docker.service", version)
	socketURL := fmt.Sprintf("https://raw.githubusercontent.com/Mirantis/cri-dockerd/v%s/packaging/systemd/cri-docker.socket", version)

	serviceBytes, err := getURLBytes(serviceURL)
	if err != nil {
		return err
	}
	socketBytes, err := getURLBytes(socketURL)
	if err != nil {
		return err
	}

	paths := []string{
		"../deploy/iso/minikube-iso/arch/x86_64/package/cri-dockerd/",
		"../deploy/iso/minikube-iso/arch/aarch64/package/cri-dockerd-aarch64/",
	}

	for _, path := range paths {
		if err := os.WriteFile(path+"cri-docker.service", serviceBytes, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(path+"cri-docker.socket", socketBytes, 0644); err != nil {
			return err
		}
	}

	return nil
}

func getURLBytes(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
