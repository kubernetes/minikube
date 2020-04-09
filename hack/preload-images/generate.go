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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

func generateTarball(kubernetesVersion, containerRuntime, tarballFilename string) error {
	defer func() {
		if err := deleteMinikube(); err != nil {
			fmt.Println(err)
		}
	}()

	driver := kic.NewDriver(kic.Config{
		KubernetesVersion: kubernetesVersion,
		ContainerRuntime:  driver.Docker,
		OCIBinary:         oci.Docker,
		MachineName:       profile,
		ImageDigest:       kic.BaseImage,
		StorePath:         localpath.MiniPath(),
		CPU:               2,
		Memory:            4000,
		APIServerPort:     8080,
	})

	baseDir := filepath.Dir(driver.GetSSHKeyPath())
	defer os.Remove(baseDir)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return errors.Wrap(err, "mkdir")
	}
	if err := driver.Create(); err != nil {
		return errors.Wrap(err, "creating kic driver")
	}

	// Now, get images to pull
	imgs, err := images.Kubeadm("", kubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	if containerRuntime != "docker" { // kic overlay image is only needed by containerd and cri-o https://github.com/kubernetes/minikube/issues/7428
		imgs = append(imgs, kic.OverlayImage)
	}

	for _, img := range imgs {
		cmd := exec.Command("docker", "exec", profile, "docker", "pull", img)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "downloading %s", img)
		}
	}

	// Transfer in k8s binaries
	kcfg := config.KubernetesConfig{
		KubernetesVersion: kubernetesVersion,
	}
	runner := command.NewKICRunner(profile, driver.OCIBinary)
	sm := sysinit.New(runner)

	if err := bsutil.TransferBinaries(kcfg, runner, sm); err != nil {
		return errors.Wrap(err, "transferring k8s binaries")
	}
	// Create image tarball
	if err := createImageTarball(tarballFilename); err != nil {
		return errors.Wrap(err, "create tarball")
	}
	return copyTarballToHost(tarballFilename)
}

func createImageTarball(tarballFilename string) error {
	// directories to save into tarball
	dirs := []string{
		fmt.Sprintf("./lib/docker/%s", dockerStorageDriver),
		"./lib/docker/image",
		"./lib/minikube/binaries",
	}
	args := []string{"exec", profile, "sudo", "tar", "-I", "lz4", "-C", "/var", "-cvf", tarballFilename}
	args = append(args, dirs...)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "tarball cmd: %s", cmd.Args)
	}
	return nil
}

func copyTarballToHost(tarballFilename string) error {
	dest := filepath.Join("out/", tarballFilename)
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:/%s", profile, tarballFilename), dest)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "cp cmd: %s", cmd.Args)
	}
	return nil
}

func deleteMinikube() error {
	cmd := exec.Command(minikubePath, "delete", "-p", profile)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
