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

package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
)

const (
	profile      = "generate-preloaded-images-tar"
	minikubePath = "out/minikube"
)

var (
	dockerStorageDriver     = "overlay2"
	preloadedTarballVersion = "v1"
	containerRuntimes       = []string{"docker"}
)

func main() {
	if err := verifyDockerStorage(); err != nil {
		exit.WithError("Docker storage type is incompatible: %v\n", err)
	}
	kubernetesVersions, err := RecentK8sVersions()
	if err != nil {
		exit.WithError("Unable to get recent k8s versions: %v\n", err)
	}

	for _, kubernetesVersion := range kubernetesVersions {
		for _, cr := range containerRuntimes {
			tf := tarballFilename(kubernetesVersion, cr)
			if tarballExists(tf) {
				fmt.Printf("A preloaded tarball for k8s version %s already exists, skipping generation.\n", kubernetesVersion)
				continue
			}
			fmt.Printf("A preloaded tarball for k8s version %s doesn't exist, generating now...\n", kubernetesVersion)
			if err := generateTarball(kubernetesVersion, tf); err != nil {
				exit.WithError(fmt.Sprintf("generating tarball for k8s version %s with %s", kubernetesVersion, cr), err)
			}
			if err := uploadTarball(tf); err != nil {
				exit.WithError(fmt.Sprintf("uploading tarball for k8s version %s with %s", kubernetesVersion, cr), err)
			}
		}
	}
}

func tarballFilename(kubernetesVersion string, containerRuntime string) string {
	return fmt.Sprintf("preloaded-images-k8s-%s-%s-%s-%s.tar.lz4", preloadedTarballVersion, kubernetesVersion, containerRuntime, dockerStorageDriver)
}

func tarballExists(tarballFilename string) bool {
	fmt.Println("Checking if tarball already exists...")
	gcsPath := fmt.Sprintf("gs://%s/%s", download.PreloadBucket, tarballFilename)
	cmd := exec.Command("gsutil", "stat", gcsPath)
	return cmd.Run() == nil
}

func verifyDockerStorage() error {
	cmd := exec.Command("docker", "info", "-f", "{{.Info.Driver}}")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%v: %v:\n%s", cmd.Args, err, stderr.String())
	}
	driver := strings.Trim(string(output), " \n")
	if driver != dockerStorageDriver {
		return fmt.Errorf("docker storage driver %s does not match requested %s", driver, dockerStorageDriver)
	}
	return nil
}
