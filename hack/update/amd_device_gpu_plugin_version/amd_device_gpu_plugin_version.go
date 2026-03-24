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
	"strings"
	"time"

	"k8s.io/minikube/hack/update"

	"k8s.io/klog/v2"
)

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

// findMatchingImageTag finds the Docker Hub image tag corresponding to a GitHub release tag.
// Docker Hub tags may include a numeric patch suffix not present in the GitHub release tag
// (e.g. GitHub "v1.25.2" may correspond to Docker Hub "v1.25.2.8").
// Tags should be provided newest-first (Docker Hub API default); the first match is returned.
func findMatchingImageTag(tags []string, ghTag string) (string, error) {
	for _, tag := range tags {
		if tag == ghTag || strings.HasPrefix(tag, ghTag+".") {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no Docker Hub image tag found matching GitHub release tag %s", ghTag)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "ROCm", "k8s-device-plugin")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	dockerTags, err := update.ImageTagsFromDockerHub("rocm/k8s-device-plugin")
	if err != nil {
		klog.Fatalf("failed to get Docker Hub tags for rocm/k8s-device-plugin: %v", err)
	}

	imageTag, err := findMatchingImageTag(dockerTags, stable.Tag)
	if err != nil {
		klog.Fatalf("failed to find Docker Hub image tag matching GitHub release tag %s: %v", stable.Tag, err)
	}

	sha, err := update.GetImageSHA(fmt.Sprintf("rocm/k8s-device-plugin:%s", imageTag))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: imageTag, SHA: sha}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
}
