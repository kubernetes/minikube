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

package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

var (
	// list of registries - keep it in sync with those in "pkg/drivers/kic/types.go"
	registries = []registry{
		{
			name:     "Google Cloud Container Registry",
			image:    "gcr.io/k8s-minikube/kicbase",
			username: os.Getenv("GCR_USERNAME"),
			password: os.Getenv("GCR_TOKEN"),
		},
		{
			name:     "Docker Hub Container Registry",
			image:    "docker.io/kicbase/stable",
			username: os.Getenv("DOCKER_USERNAME"),
			password: os.Getenv("DOCKER_TOKEN"),
		},
		{
			name:     "GitHub Packages Registry",
			image:    "docker.pkg.github.com/kubernetes/minikube/kicbase",
			username: os.Getenv("GITHUB_USERNAME"),
			password: os.Getenv("GITHUB_TOKEN"),
		},
	}
)

// registry contains a container registry name, image path, and credentials.
type registry struct {
	name     string
	image    string
	username string
	password string
}

// CRUpdateAll updates all registries, and returns if at least one got updated.
func CRUpdateAll(ctx context.Context, image, version string) (updated bool) {
	for _, reg := range registries {
		if err := crUpdate(ctx, reg, image, version); err != nil {
			klog.Errorf("Unable to update %s", reg.name)
			continue
		}
		klog.Infof("Successfully updated %s", reg.name)
		updated = true
	}
	return updated
}

// crUpdate tags image with version, pushes it to container registry, and returns any error occurred.
func crUpdate(ctx context.Context, reg registry, image, version string) error {
	login := exec.CommandContext(ctx, "docker", "login", "--username", reg.username, "--password-stdin", reg.image)
	if err := RunWithRetryNotify(ctx, login, strings.NewReader(reg.password), 1*time.Minute, 10); err != nil {
		return fmt.Errorf("unable to login to %s: %w", reg.name, err)
	}
	klog.Infof("Successfully logged in to %s", reg.name)

	tag := exec.CommandContext(ctx, "docker", "tag", image+":"+version, reg.image+":"+version)
	if err := RunWithRetryNotify(ctx, tag, nil, 1*time.Minute, 10); err != nil {
		return fmt.Errorf("unable to tag %s for %s: %w", reg.image+":"+version, reg.name, err)
	}
	klog.Infof("Successfully tagged %s for %s", reg.image+":"+version, reg.name)

	push := exec.CommandContext(ctx, "docker", "push", reg.image+":"+version)
	if err := RunWithRetryNotify(ctx, push, nil, 2*time.Minute, 10); err != nil {
		return fmt.Errorf("unable to push %s to %s: %w", reg.image+":"+version, reg.name, err)
	}
	klog.Infof("Successfully pushed %s to %s", reg.image+":"+version, reg.name)

	return nil
}

// TagImage tags local image:current with stable version, and returns any error occurred.
func TagImage(ctx context.Context, image, current, stable string) error {
	tag := exec.CommandContext(ctx, "docker", "tag", image+":"+current, image+":"+stable)
	if err := RunWithRetryNotify(ctx, tag, nil, 1*time.Second, 10); err != nil {
		return err
	}
	return nil
}

// PullImage checks if current image exists locally, tries to pull it if not, and returns reference image url and any error occurred.
func PullImage(ctx context.Context, current, release string) (image string, err error) {
	// check if image exists locally
	for _, reg := range registries {
		inspect := exec.CommandContext(ctx, "docker", "inspect", reg.image+":"+current, "--format", "{{.Id}}")
		if err := RunWithRetryNotify(ctx, inspect, nil, 1*time.Second, 10); err != nil {
			continue
		}
		image = reg.image
		break
	}
	if image == "" {
		// try to pull image locally
		for _, reg := range registries {
			pull := exec.CommandContext(ctx, "docker", "pull", reg.image+":"+current)
			if err := RunWithRetryNotify(ctx, pull, nil, 2*time.Minute, 10); err != nil {
				continue
			}
			image = reg.image
			break
		}
	}
	if image == "" {
		return "", fmt.Errorf("unable to find current image version tag %s locally nor in any registry", current)
	}
	return image, nil
}
