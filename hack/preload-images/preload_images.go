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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
)

const (
	profile      = "generate-preloaded-images-tar"
	minikubePath = "out/minikube"
)

var (
	dockerStorageDriver = "overlay2"
	podmanStorageDriver = "overlay"
	containerRuntimes   = []string{"docker", "containerd", "cri-o"}
	k8sVersion          string
	k8sVersions         []string
)

func init() {
	flag.StringVar(&k8sVersion, "kubernetes-version", "", "desired Kubernetes version, for example `v1.17.2`")
	flag.Parse()
	if k8sVersion != "" {
		k8sVersions = append(k8sVersions, k8sVersion)
	}
	viper.Set("preload", "true")
}

func main() {
	defer func() {
		if err := deleteMinikube(); err != nil {
			fmt.Printf("error cleaning up minikube: %v \n", err)
		}
	}()

	if err := deleteMinikube(); err != nil {
		fmt.Printf("error cleaning up minikube at start up: %v \n", err)
	}

	if k8sVersions == nil {
		var err error
		k8sVersions, err = RecentK8sVersions()
		if err != nil {
			exit("Unable to get recent k8s versions: %v\n", err)
		}
	}

	k8sVersions = append(k8sVersions, constants.DefaultKubernetesVersion, constants.NewestKubernetesVersion, constants.OldestKubernetesVersion)

	for _, kv := range k8sVersions {
		for _, cr := range containerRuntimes {
			tf := download.TarballName(kv, cr)
			if download.PreloadExists(kv, cr) {
				fmt.Printf("A preloaded tarball for k8s version %s - runtime %q already exists, skipping generation.\n", kv, cr)
				continue
			}
			fmt.Printf("A preloaded tarball for k8s version %s - runtime %q doesn't exist, generating now...\n", kv, cr)
			if err := generateTarball(kv, cr, tf); err != nil {
				exit(fmt.Sprintf("generating tarball for k8s version %s with %s", kv, cr), err)
			}
			if err := uploadTarball(tf); err != nil {
				exit(fmt.Sprintf("uploading tarball for k8s version %s with %s", kv, cr), err)
			}

			if err := deleteMinikube(); err != nil {
				fmt.Printf("error cleaning up minikube before finishing up: %v\n", err)
			}

		}
	}
}

func verifyDockerStorage() error {
	cmd := exec.Command("docker", "exec", profile, "docker", "info", "-f", "{{.Info.Driver}}")
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

func verifyPodmanStorage() error {
	cmd := exec.Command("docker", "exec", profile, "sudo", "podman", "info", "-f", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%v: %v:\n%s", cmd.Args, err, stderr.String())
	}
	var info map[string]map[string]interface{}
	err = json.Unmarshal(output, &info)
	if err != nil {
		return err
	}
	driver := info["store"]["graphDriverName"]
	if driver != podmanStorageDriver {
		return fmt.Errorf("podman storage driver %s does not match requested %s", driver, podmanStorageDriver)
	}
	return nil
}

// exit will exit and clean up minikube
func exit(msg string, err error) {
	fmt.Printf("WithError(%s)=%v called from:\n%s", msg, err, debug.Stack())
	if err := deleteMinikube(); err != nil {
		fmt.Printf("error cleaning up minikube at start up: %v\n", err)
	}
	os.Exit(60)
}
