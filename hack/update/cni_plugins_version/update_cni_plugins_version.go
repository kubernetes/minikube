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

const cxTimeout = 5 * time.Minute

var (
	schema = map[string]update.Item{
		"deploy/iso/minikube-iso/arch/aarch64/package/cni-plugins-latest-aarch64/cni-plugins-latest.mk": {
			Replace: map[string]string{
				`CNI_PLUGINS_LATEST_AARCH64_VERSION = .*`: `CNI_PLUGINS_LATEST_AARCH64_VERSION = {{.Version}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/cni-plugins-latest/cni-plugins-latest.mk": {
			Replace: map[string]string{
				`CNI_PLUGINS_LATEST_VERSION = .*`: `CNI_PLUGINS_LATEST_VERSION = {{.Version}}`,
			},
		},
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`CNI_PLUGINS_LATEST_VERSION=.*`: `CNI_PLUGINS_LATEST_VERSION="{{.Version}}"`,
			},
		},
	}
)

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "containernetworking", "plugins")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	data := Data{Version: stable.Tag}

	update.Apply(schema, data)

	if err := updateHashFile(data.Version, "arm64", "aarch64/package/cni-plugins-latest-aarch64"); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}
	if err := updateHashFile(data.Version, "amd64", "x86_64/package/cni-plugins-latest"); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}
}

func updateHashFile(version, arch, packagePath string) error {
	r, err := http.Get(fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", version, arch, version))
	if err != nil {
		return fmt.Errorf("failed to download source code: %v", err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	sum := sha256.Sum256(b)
	filePath := fmt.Sprintf("../../../deploy/iso/minikube-iso/arch/%s/cni-plugins-latest.hash", packagePath)
	b, err = os.ReadFile(filePath)
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
	if _, err := f.WriteString(fmt.Sprintf("sha256 %x  cni-plugins-linux-%s-%s.tgz\n", sum, arch, version)); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
