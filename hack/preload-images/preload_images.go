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
	"strings"

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
	cmd := exec.Command(minikubePath, "start", "-p", profile, "--memory", "4000", "--kubernetes-version", kubernetesVersion, "--wait=false")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func createImageTarball() error {
	cmd := exec.Command(minikubePath, "ssh", "-p", profile, "--", "sudo", "tar", "cvf", tarballFilename, "/var/lib/docker", "/var/lib/minikube/binaries")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func copyTarballToHost() error {
	sshKey, err := runCmd([]string{minikubePath, "ssh-key", "-p", profile})
	if err != nil {
		return errors.Wrap(err, "getting ssh-key")
	}

	ip, err := runCmd([]string{minikubePath, "ip", "-p", profile})
	if err != nil {
		return errors.Wrap(err, "getting ip")
	}

	dest := filepath.Join("out/", tarballFilename)
	args := []string{"scp", "-o", "StrictHostKeyChecking=no", "-i", sshKey, fmt.Sprintf("docker@%s:/home/docker/%s", ip, tarballFilename), dest}
	_, err = runCmd(args)
	return err
}

func deleteMinikube() error {
	cmd := exec.Command(minikubePath, "delete", "-p", profile)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func runCmd(command []string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	output, err := cmd.Output()
	return strings.Trim(string(output), "\n "), err
}
