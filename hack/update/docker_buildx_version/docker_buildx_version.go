/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
		"deploy/iso/minikube-iso/arch/aarch64/package/docker-buildx-aarch64/docker-buildx.mk": {
			Replace: map[string]string{
				`DOCKER_BUILDX_AARCH64_VERSION = .*`: `DOCKER_BUILDX_AARCH64_VERSION = {{.Version}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/docker-buildx/docker-buildx.mk": {
			Replace: map[string]string{
				`DOCKER_BUILDX_VERSION = .*`: `DOCKER_BUILDX_VERSION = {{.Version}}`,
			},
		},
	}
)

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "docker", "buildx")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	data := Data{Version: stable.Tag}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}

	if err := updateHashFiles(data.Version); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}
}

func getHashesFromChecksums(version string) (map[string]string, error) {
	url := fmt.Sprintf("https://github.com/docker/buildx/releases/download/%s/checksums.txt", version)
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download checksums.txt: %v", err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download checksums.txt: status %s", r.Status)
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksums body: %v", err)
	}
	lines := strings.Split(string(b), "\n")
	hashes := make(map[string]string)
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		name := strings.TrimPrefix(parts[1], "*")
		if strings.Contains(name, "linux-amd64") || strings.Contains(name, "linux-arm64") {
			hashes[name] = hash
		}
	}
	return hashes, nil
}

func updateHashFiles(version string) error {
	hashes, err := getHashesFromChecksums(version)
	if err != nil {
		return err
	}

	amd64Binary := fmt.Sprintf("buildx-%s.linux-amd64", version)
	amd64Hash, ok := hashes[amd64Binary]
	if !ok {
		return fmt.Errorf("checksum for %s not found", amd64Binary)
	}

	arm64Binary := fmt.Sprintf("buildx-%s.linux-arm64", version)
	arm64Hash, ok := hashes[arm64Binary]
	if !ok {
		return fmt.Errorf("checksum for %s not found", arm64Binary)
	}

	if err := updateHashFile(version, "aarch64", "-aarch64", arm64Binary, arm64Hash); err != nil {
		return fmt.Errorf("aarch64: %v", err)
	}
	if err := updateHashFile(version, "x86_64", "", amd64Binary, amd64Hash); err != nil {
		return fmt.Errorf("x86_64: %v", err)
	}
	return nil
}

func updateHashFile(version, arch, folderSuffix, binaryName, shaSum string) error {
	filePath := fmt.Sprintf("../deploy/iso/minikube-iso/arch/%s/package/docker-buildx%s/docker-buildx.hash", arch, folderSuffix)
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read hash file: %v", err)
	}
	if strings.Contains(string(b), version) {
		klog.Infof("hash file already contains %q", version)
		return nil
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open hash file: %v", err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "sha256 %s  %s\n", shaSum, binaryName); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
