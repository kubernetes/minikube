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

var (
	schema = map[string]update.Item{
		"deploy/iso/minikube-iso/arch/aarch64/package/nerdctl-bin-aarch64/nerdctl-bin.mk": {
			Replace: map[string]string{
				`NERDCTL_BIN_AARCH64_VERSION = .*`: `NERDCTL_BIN_AARCH64_VERSION = {{.Version}}`,
				`NERDCTL_BIN_AARCH64_COMMIT = .*`:  `NERDCTL_BIN_AARCH64_COMMIT = {{.Commit}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/nerdctl-bin/nerdctl-bin.mk": {
			Replace: map[string]string{
				`NERDCTL_BIN_VERSION = .*`: `NERDCTL_BIN_VERSION = {{.Version}}`,
				`NERDCTL_BIN_COMMIT = .*`:  `NERDCTL_BIN_COMMIT = {{.Commit}}`,
			},
		},
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`NERDCTL_VERSION=.*`: `NERDCTL_VERSION="{{.Version}}"`,
			},
		},
	}
)

type Data struct {
	Version string
	Commit  string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "containerd", "nerdctl")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	version := strings.TrimPrefix(stable.Tag, "v")

	data := Data{Version: version, Commit: stable.Commit}

	update.Apply(schema, data)

	if err := updateHashFile(version, "arm64", "aarch64/package/nerdctl-bin-aarch64"); err != nil {
		klog.Fatalf("failed updating arm64 hash file: %v", err)
	}
	if err := updateHashFile(version, "amd64", "x86_64/package/nerdctl-bin"); err != nil {
		klog.Fatalf("failed updating amd64 hash file: %v", err)
	}
}

func updateHashFile(version, arch, packagePath string) error {
	r, err := http.Get(fmt.Sprintf("https://github.com/containerd/nerdctl/releases/download/v%s/nerdctl-%s-linux-%s.tar.gz", version, version, arch))
	if err != nil {
		return fmt.Errorf("failed to download source code: %v", err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	sum := sha256.Sum256(b)
	filePath := fmt.Sprintf("../../../deploy/iso/minikube-iso/arch/%s/nerdctl-bin.hash", packagePath)
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
	if _, err := f.WriteString(fmt.Sprintf("sha256 %x  nerdctl-%s-linux-%s.tar.gz\n", sum, version, arch)); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
