/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
		"deploy/iso/minikube-iso/package/crun-latest/crun-latest.mk": {
			Replace: map[string]string{
				`CRUN_LATEST_VERSION = .*`: `CRUN_LATEST_VERSION = {{.Version}}`,
				`CRUN_LATEST_COMMIT = .*`:  `CRUN_LATEST_COMMIT = {{.Commit}}`,
			},
		},
	}
)

type Data struct {
	Version string
	Commit  string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "containers", "crun")
	if err != nil {
		klog.Fatalf("Unable to get crun stable version: %v", err)
	}

	// Hack: Strip 'v' prefix added in 'update.GHReleases' for this package
	data := Data{Version: strings.Trim(stable.Tag, "v"), Commit: stable.Commit}

	update.Apply(schema, data)

	if err := updateHashFiles(data.Version); err != nil {
		klog.Fatalf("failed to update hash files: %v", err)
	}
}

func updateHashFiles(version string) error {
	r, err := http.Get(fmt.Sprintf("https://github.com/containers/crun/releases/download/%s/crun-%s.tar.gz", version, version))
	if err != nil {
		return fmt.Errorf("failed to download source code: %v", err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	sum := sha256.Sum256(b)
	filePath := "../../../deploy/iso/minikube-iso/package/crun-latest/crun-latest.hash"
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
	if _, err := f.WriteString(fmt.Sprintf("sha256 %x crun-%s.tar.gz\n", sum, version)); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
