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

	"golang.org/x/mod/semver"

	"k8s.io/minikube/hack/update"

	"k8s.io/klog/v2"
)

const (
	cxTimeout = 5 * time.Minute

	// dockerTagPrefix is the prefix moby uses for Docker Engine release tags since v29 (ie, 'docker-v29.9.3').
	dockerTagPrefix = "docker-"
)

var schema = map[string]update.Item{
	"deploy/iso/minikube-iso/arch/aarch64/package/docker-bin-aarch64/docker-bin.mk": {
		Replace: map[string]string{
			`DOCKER_BIN_AARCH64_VERSION = .*`: `DOCKER_BIN_AARCH64_VERSION = {{.Version}}`,
		},
	},
	"deploy/iso/minikube-iso/arch/x86_64/package/docker-bin/docker-bin.mk": {
		Replace: map[string]string{
			`DOCKER_BIN_VERSION = .*`: `DOCKER_BIN_VERSION = {{.Version}}`,
		},
	},
}

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// List only 'docker-v*' tags: moby also has 'api/v*', 'client/v*' and legacy 'v*' tags,
	// and 'docker-v*' tags sort too late to be reached by update.GHReleases.
	refs, _, err := update.GHClient().Git.ListMatchingRefs(ctx, "moby", "moby", "tags/"+dockerTagPrefix+"v")
	if err != nil {
		klog.Fatalf("Unable to list docker tags: %v", err)
	}
	tags := make([]string, 0, len(refs))
	for _, ref := range refs {
		tags = append(tags, strings.TrimPrefix(ref.GetRef(), "refs/tags/"))
	}

	version := latestStable(tags)
	if version == "" {
		klog.Fatalf("Unable to find docker stable version in %d %q tags", len(tags), dockerTagPrefix+"v")
	}

	data := Data{Version: version}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}

	if err := updateHashFile(data.Version, "aarch64", "-aarch64"); err != nil {
		klog.Fatalf("failed to update hash file: %v", err)
	}
	if err := updateHashFile(data.Version, "x86_64", ""); err != nil {
		klog.Fatalf("failed to update hash file: %v", err)
	}
}

// latestStable returns the greatest stable version (without 'v') among 'docker-v*' tags,
// or an empty string if there is none. Pre-releases and other tags are ignored.
func latestStable(tags []string) string {
	stable := ""
	for _, tag := range tags {
		ver, ok := strings.CutPrefix(tag, dockerTagPrefix)
		if !ok || !semver.IsValid(ver) || semver.Prerelease(ver) != "" {
			continue
		}
		if stable == "" || semver.Compare(ver, stable) == 1 {
			stable = ver
		}
	}
	return strings.TrimPrefix(stable, "v")
}

func updateHashFile(version, arch, folderSuffix string) error {
	r, err := http.Get(fmt.Sprintf("https://download.docker.com/linux/static/stable/%s/docker-%s.tgz", arch, version))
	if err != nil {
		return fmt.Errorf("failed to download source code: %v", err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	sum := sha256.Sum256(b)
	filePath := fmt.Sprintf("../deploy/iso/minikube-iso/arch/%s/package/docker-bin%s/docker-bin.hash", arch, folderSuffix)
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
	if _, err := fmt.Fprintf(f, "sha256 %x  docker-%s.tgz\n", sum, version); err != nil {
		return fmt.Errorf("failed to write to hash file: %v", err)
	}
	return nil
}
