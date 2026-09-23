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
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/google/go-github/v85/github"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits.
	// (ref: https://pkg.go.dev/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ghListPerPage.
	ghSearchLimit = 300
)

// versionRegex matches ROCm release tags, which use 4 components (e.g. v1.31.0.11)
// and are not valid semver, so update.GHReleases cannot be used.
var versionRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)\.(\d+)$`)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`rocm/k8s-device-plugin:.*`: `rocm/k8s-device-plugin:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stable, err := latestStableVersion(ctx)
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	klog.Infof("Stable version: %s", stable)
	sha, err := update.GetImageSHA(fmt.Sprintf("rocm/k8s-device-plugin:%s", stable))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: stable, SHA: sha}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
}

// latestStableVersion returns the greatest non pre-release version (e.g. 1.31.0.11)
// from ROCm/k8s-device-plugin GitHub releases. Docker Hub image tags use the same
// version without the "v" prefix.
func latestStableVersion(ctx context.Context) (string, error) {
	ghc := update.GHClient()
	var latest []int
	latestTag := ""

	// walk through the paginated list of up to ghSearchLimit newest releases
	opts := &github.ListOptions{PerPage: ghListPerPage}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
		rls, resp, err := ghc.Repositories.ListReleases(ctx, "ROCm", "k8s-device-plugin", opts)
		if err != nil {
			return "", err
		}
		for _, rl := range rls {
			if rl.GetDraft() || rl.GetPrerelease() {
				continue
			}
			ver, ok := parseVersion(rl.GetTagName())
			if !ok {
				continue
			}
			if compareVersions(ver, latest) > 0 {
				latest = ver
				latestTag = rl.GetTagName()[1:]
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if latestTag == "" {
		return "", fmt.Errorf("no stable release found")
	}
	return latestTag, nil
}

// parseVersion parses a tag like v1.31.0.11 into its numeric components.
func parseVersion(tag string) ([]int, bool) {
	m := versionRegex.FindStringSubmatch(tag)
	if m == nil {
		return nil, false
	}
	ver := make([]int, 0, 4)
	for _, s := range m[1:] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, false
		}
		ver = append(ver, n)
	}
	return ver, true
}

// compareVersions returns 1 if a > b, -1 if a < b and 0 if equal.
// A nil version is lower than any other version.
func compareVersions(a, b []int) int {
	if b == nil {
		if a == nil {
			return 0
		}
		return 1
	}
	for i := range a {
		if a[i] != b[i] {
			if a[i] > b[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
