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
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute
)

var (
	schema = map[string]update.Item{
		"hack/jenkins/installers/check_install_gh.sh": {
			Replace: map[string]string{
				`GH_VERSION=".*"`: `GH_VERSION="{{.StableVersion}}"`,
			},
		},
	}
)

// Data holds stable gh version in semver format.
type Data struct {
	StableVersion string
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get gh stable version
	stable, err := ghVersion(ctx, "cli", "cli")
	if err != nil {
		klog.Fatalf("Unable to get gh stable version: %v", err)
	}
	data := Data{StableVersion: stable}
	klog.Infof("gh stable version: %s", data.StableVersion)

	update.Apply(schema, data)
}

// ghVersion returns stable version in semver format.
func ghVersion(ctx context.Context, owner, repo string) (stable string, err error) {
	// get gh version from GitHub Releases
	stable, _, _, err = update.GHReleases(ctx, owner, repo)
	if err != nil || !semver.IsValid(stable) {
		return "", err
	}
	return strings.TrimPrefix(stable, "v"), nil
}
