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
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
)

const (
	profile      = "generate-preloaded-images-tar"
	minikubePath = "out/minikube"
)

var (
	dockerStorageDriver = "overlay2"
	containerRuntimes   = []string{"docker"}
	k8sVersion          string
	k8sVersions         []string
)

func init() {
	flag.StringVar(&k8sVersion, "kubernetes-version", "", "desired kubernetes version, for example `v1.17.2`")
	flag.Parse()
	if k8sVersion != "" {
		k8sVersions = append(k8sVersions, k8sVersion)
	}
	viper.Set("preload", "true")
}

func main() {
	if err := verifyDockerStorage(); err != nil {
		exit.WithError("Docker storage type is incompatible: %v\n", err)
	}
	if k8sVersions == nil {
		var err error
		k8sVersions, err = RecentK8sVersions()
		if err != nil {
			exit.WithError("Unable to get recent k8s versions: %v\n", err)
		}
	}

	for _, kv := range k8sVersions {
		for _, cr := range containerRuntimes {
			tf := download.TarballName(kv, cr)
			if download.PreloadExists(kv, cr) {
				fmt.Printf("A preloaded tarball for k8s version %s already exists, skipping generation.\n", kv)
				continue
			}
			fmt.Printf("A preloaded tarball for k8s version %s doesn't exist, generating now...\n", kv)
			if err := generateTarball(kv, tf); err != nil {
				exit.WithError(fmt.Sprintf("generating tarball for k8s version %s with %s", kv, cr), err)
			}
			if err := uploadTarball(tf); err != nil {
				exit.WithError(fmt.Sprintf("uploading tarball for k8s version %s with %s", kv, cr), err)
			}
		}
	}
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
