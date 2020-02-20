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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	profile      = "generate-preloaded-images-tar"
	minikubePath = "out/minikube"
)

var (
	kubernetesVersion = ""
	tarballFilename   = ""
)

func init() {
	flag.StringVar(&kubernetesVersion, "kubernetes-version", "", "desired kubernetes version, for example `v1.17.2`")
	flag.Parse()
	tarballFilename = fmt.Sprintf("preloaded-images-k8s-%s.tar", kubernetesVersion)
}

func main() {
	if err := executePreloadImages(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executePreloadImages() error {
	defer func() {
		if err := deleteMinikube(); err != nil {
			fmt.Println(err)
		}
	}()
	if err := startMinikube(); err != nil {
		return err
	}
	if err := createImageTarball(); err != nil {
		return err
	}
	return copyTarballToHost()
}

func startMinikube() error {
	cmd := exec.Command(minikubePath, "start", "-p", profile, "--memory", "4000", "--kubernetes-version", kubernetesVersion, "--wait=false", "--vm-driver=docker")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func createImageTarball() error {
	cmd := exec.Command(minikubePath, "ssh", "-p", profile, "--", "cd", "/var", "&&", "sudo", "tar", "cvf", tarballFilename, "./lib/docker", "./lib/minikube/binaries")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "creating image tarball")
	}
	return nil
}

func copyTarballToHost() error {
	dest := filepath.Join("out/", tarballFilename)
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:/var/%s", profile, tarballFilename), dest)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "copying tarball to host")
	}
	return nil
}

func deleteMinikube() error {
	cmd := exec.Command(minikubePath, "delete", "-p", profile)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
