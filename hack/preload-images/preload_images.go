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

	"github.com/pkg/errors"

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
	k8sVersions         []string
	k8sVersion          = flag.String("kubernetes-version", "", "desired Kubernetes version, for example `v1.17.2`")
	noUpload            = flag.Bool("no-upload", false, "Do not upload tarballs to GCS")
	force               = flag.Bool("force", false, "Generate the preload tarball even if it's already exists")
	limit               = flag.Int("limit", 0, "Limit the number of tarballs to generate")
)

type preloadCfg struct {
	k8sVer  string
	runtime string
}

func (p preloadCfg) String() string {
	return fmt.Sprintf("%q/%q", p.runtime, p.k8sVer)
}

func main() {
	flag.Parse()

	// used by pkg/minikube/download.PreloadExists()
	viper.Set("preload", "true")

	if *k8sVersion != "" {
		k8sVersions = []string{*k8sVersion}
	}

	if err := deleteMinikube(); err != nil {
		fmt.Printf("error cleaning up minikube at start up: %v \n", err)
	}

	k8sVersions, err := collectK8sVers()
	if err != nil {
		exit("Unable to get recent k8s versions: %v\n", err)
	}

	var toGenerate []preloadCfg
	var i int

out:
	for _, kv := range k8sVersions {
		for _, cr := range containerRuntimes {
			if *limit > 0 && i >= *limit {
				break out
			}
			// Since none/mock are the only exceptions, it does not matter what driver we choose.
			if !download.PreloadExists(kv, cr, "docker") {
				toGenerate = append(toGenerate, preloadCfg{kv, cr})
				i++
				fmt.Printf("[%d] A preloaded tarball for k8s version %s - runtime %q does not exist.\n", i, kv, cr)
			} else if *force {
				// the tarball already exists, but '--force' is passed. we need to overwrite the file
				toGenerate = append(toGenerate, preloadCfg{kv, cr})
				i++
				fmt.Printf("[%d] A preloaded tarball for k8s version %s - runtime %q already exists. Going to overwrite it.\n", i, kv, cr)
			} else {
				fmt.Printf("A preloaded tarball for k8s version %s - runtime %q already exists, skipping generation.\n", kv, cr)
			}
		}
	}

	fmt.Printf("Going to generate preloads for %v\n", toGenerate)

	for _, cfg := range toGenerate {
		if err := makePreload(cfg); err != nil {
			exit(err.Error(), err)
		}
	}
}

func collectK8sVers() ([]string, error) {
	if k8sVersions == nil {
		recent, err := recentK8sVersions()
		if err != nil {
			return nil, err
		}
		k8sVersions = recent
	}
	return append([]string{
		constants.DefaultKubernetesVersion,
		constants.NewestKubernetesVersion,
		constants.OldestKubernetesVersion,
	}, k8sVersions...), nil
}

func makePreload(cfg preloadCfg) error {
	kv, cr := cfg.k8sVer, cfg.runtime

	fmt.Printf("A preloaded tarball for k8s version %s - runtime %q doesn't exist, generating now...\n", kv, cr)
	tf := download.TarballName(kv, cr)

	defer func() {
		if err := deleteMinikube(); err != nil {
			fmt.Printf("error cleaning up minikube before finishing up: %v\n", err)
		}
	}()

	if err := generateTarball(kv, cr, tf); err != nil {
		return errors.Wrap(err, fmt.Sprintf("generating tarball for k8s version %s with %s", kv, cr))
	}

	if *noUpload {
		fmt.Printf("skip upload of %q\n", tf)
		return nil
	}
	if err := uploadTarball(tf); err != nil {
		return errors.Wrap(err, fmt.Sprintf("uploading tarball for k8s version %s with %s", kv, cr))
	}
	return nil
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
