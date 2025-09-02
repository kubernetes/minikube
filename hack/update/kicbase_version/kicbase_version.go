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

/*
Script promotes current KIC base image as stable, ie:
  - strips current version suffix starting from '-' in pkg/drivers/kic/types.go => release version
    (eg, 'v0.0.13-snapshot1' -> 'v0.0.13')
  - makes sure current KIC base image exists locally, tries to pull one if not
  - tags current KIC base image with the release version, and
  - pushes it to all relevant container registries

Script also requires following credentials as env variables (injected by Jenkins credential provider):
  @GCR (ref: https://cloud.google.com/container-registry/docs/advanced-authentication):
  - GCR_USERNAME=<string>: GCR username, eg:
	= "oauth2accesstoken" if Access Token is used for GCR_TOKEN, or
	= "_json_key" if JSON Key File is used for GCR_TOKEN
  - GCR_TOKEN=<string>: GCR JSON token

  @Docker (ref: https://docs.docker.com/docker-hub/access-tokens/)
  - DOCKER_USERNAME=<string>: Docker username
  - DOCKER_TOKEN=<string>: Docker personal access token or password

  @GitHub (ref: https://docs.github.com/en/free-pro-team@latest/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages)
  - GITHUB_USERNAME=<string>: GitHub username
  - GITHUB_TOKEN=<string>: GitHub [personal] access token
*/

package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 10 * time.Minute
)

var (
	kicFile   = "pkg/drivers/kic/types.go"
	versionRE = `Version = "(.*)"`

	schema = map[string]update.Item{
		kicFile: {
			Replace: map[string]string{
				`Version = ".*"`: `Version = "{{.StableVersion}}"`,
			},
		},
	}
)

// Data holds current and stable KIC base image versions
type Data struct {
	CurrentVersion string
	StableVersion  string
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// determine current and stable KIC base image versions
	current, stable, err := KICVersions()
	if err != nil {
		klog.Fatalf("Unable to get KIC base image versions: %v", err)
	}
	if len(current) == 0 || len(stable) == 0 {
		klog.Fatalf("Unable to determine KIC base image versions")
	}
	data := Data{CurrentVersion: current, StableVersion: stable}
	klog.Infof("KIC base image versions: 'current' is %s and 'stable' would be %s", data.CurrentVersion, data.StableVersion)

	// prepare local KIC base image
	image, err := prepareImage(ctx, data)
	if err != nil {
		klog.Fatalf("Unable to prepare local KIC base reference image: %v", err)
	}
	klog.Infof("Local KIC base reference image: %s", image)

	// update registries
	if updated := update.CRUpdateAll(ctx, image, data.StableVersion); !updated {
		klog.Fatalf("Unable to update any registry")
	}

	update.Apply(schema, data)
}

// KICVersions returns current and stable KIC base image versions and any error occurred.
func KICVersions() (current, stable string, err error) {
	blob, err := os.ReadFile(filepath.Join(update.FSRoot, kicFile))
	if err != nil {
		return "", "", err
	}
	re := regexp.MustCompile(versionRE)
	ver := re.FindSubmatch(blob)
	if ver == nil {
		return "", "", nil
	}
	current = string(ver[1])
	stable = strings.Split(current, "-")[0]
	return current, stable, nil
}

// prepareImage checks if current image exists locally, tries to pull it if not, tags it with release version.
// Returns reference image url and any error occurred.
func prepareImage(ctx context.Context, data Data) (image string, err error) {
	image, err = update.PullImage(ctx, data.CurrentVersion)
	if err != nil {
		return "", err
	}
	if err := update.TagImage(ctx, image, data.CurrentVersion, data.StableVersion); err != nil {
		return "", err
	}
	return image, nil
}
