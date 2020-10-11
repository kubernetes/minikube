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
The script releases current kic base image as stable, ie:
  - strips current version suffix starting from '-' in pkg/drivers/kic/types.go => release version
    (eg, 'v0.0.13-snapshot1' -> 'v0.0.13')
  - makes sure current kic base image exists locally, tries to pull one if not
  - tags current kic base image with the release version, and
  - pushes it to all relevant container registries

The script requires following credentials as env variables (injected by Jenkins credential provider):
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
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const (
	// default context timeout
	cxTimeout = 600 * time.Second
)

var (
	kicFile      = "../../pkg/drivers/kic/types.go"
	kicVersionRE = `Version = "(.*)"`

	// keep list of registries in sync with those in kicFile
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

	run = func(cmd *exec.Cmd, stdin io.Reader) error {
		cmd.Stdin = stdin
		var out bytes.Buffer
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return errors.Errorf("%s: %s", err.Error(), out.String())
		}
		return nil
	}
)

// container registry name, image path, credentials, and updated flag
type registry struct {
	name     string
	image    string
	username string
	password string
	updated  bool
}

func (r *registry) setUpdated(updated bool) {
	r.updated = updated
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// write log statements to stderr instead of to files
	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("Error setting 'logtostderr' glog flag: %v", err)
	}
	flag.Parse()
	defer glog.Flush()

	// determine current kic base image version
	vCurrent, err := getKICVersion()
	if err != nil {
		glog.Fatalf("failed getting current kic base image version: %v", err)
	}
	if len(vCurrent) == 0 {
		glog.Fatalf("cannot determine current kic base image version")
	}
	glog.Infof("current kic base image version: %s", vCurrent)

	// determine release kic base image version
	vRelease := strings.Split(vCurrent, "-")[0]
	glog.Infof("release kic base image version: %s", vRelease)

	// prepare local kic base image
	image, err := prepareImage(ctx, vCurrent, vRelease)
	if err != nil {
		glog.Fatalf("failed preparing local kic base reference image: %v", err)
	}
	glog.Infof("local kic base reference image: %s", image)

	// update registries
	if updated := updateRegistries(ctx, image, vRelease); !updated {
		glog.Fatalf("failed updating all registries")
	}
}

// updateRegistries tags image with release version, pushes it to registries, and returns if any registry got updated
func updateRegistries(ctx context.Context, image, release string) (updated bool) {
	for _, reg := range registries {
		login := exec.CommandContext(ctx, "docker", "login", "--username", reg.username, "--password-stdin", reg.image)
		if err := run(login, strings.NewReader(reg.password)); err != nil {
			glog.Errorf("failed logging in to %s: %v", reg.name, err)
			continue
		}
		glog.Infof("successfully logged in to %s", reg.name)

		tag := exec.CommandContext(ctx, "docker", "tag", image+":"+release, reg.image+":"+release)
		if err := run(tag, nil); err != nil {
			glog.Errorf("failed tagging %s for %s: %v", reg.image+":"+release, reg.name, err)
			continue
		}
		glog.Infof("successfully tagged %s for %s", reg.image+":"+release, reg.name)

		push := exec.CommandContext(ctx, "docker", "push", reg.image+":"+release)
		if err := run(push, nil); err != nil {
			glog.Errorf("failed pushing %s to %s: %v", reg.image+":"+release, reg.name, err)
			continue
		}
		glog.Infof("successfully pushed %s to %s", reg.image+":"+release, reg.name)

		reg.setUpdated(true)
		glog.Infof("successfully updated %s", reg.name)
		updated = true
	}
	return updated
}

// prepareImage checks if current image exists locally, tries to pull it if not,
// tags it with release version, returns reference image url and any error
func prepareImage(ctx context.Context, current, release string) (image string, err error) {
	// check if image exists locally
	for _, reg := range registries {
		inspect := exec.CommandContext(ctx, "docker", "inspect", reg.image+":"+current, "--format", "{{.Id}}")
		if err := run(inspect, nil); err != nil {
			continue
		}
		image = reg.image
		break
	}
	if image == "" {
		// try to pull image locally
		for _, reg := range registries {
			pull := exec.CommandContext(ctx, "docker", "pull", reg.image+":"+current)
			if err := run(pull, nil); err != nil {
				continue
			}
			image = reg.image
			break
		}
	}
	if image == "" {
		return "", errors.Errorf("cannot find current image version tag %s locally nor in any registry", current)
	}
	// tag current image with release version
	tag := exec.CommandContext(ctx, "docker", "tag", image+":"+current, image+":"+release)
	if err := run(tag, nil); err != nil {
		return "", err
	}
	return image, nil
}

// getKICVersion returns current kic base image version and any error
func getKICVersion() (string, error) {
	blob, err := ioutil.ReadFile(kicFile)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(kicVersionRE)
	ver := re.FindSubmatch(blob)
	if ver == nil {
		return "", nil
	}
	return string(ver[1]), nil
}
