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
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

func generateTarball(kubernetesVersion, containerRuntime, tarballFilename string) error {
	driver := kic.NewDriver(kic.Config{
		KubernetesVersion: kubernetesVersion,
		ContainerRuntime:  containerRuntime,
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

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return errors.Wrap(err, "mkdir")
	}
	if err := driver.Create(); err != nil {
		return errors.Wrap(err, "creating kic driver")
	}

	if err := verifyStorage(containerRuntime); err != nil {
		return errors.Wrap(err, "verifying storage")
	}

	// Now, get images to pull
	imgs, err := images.Kubeadm("", kubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	if containerRuntime != "docker" { // kic overlay image is only needed by containerd and cri-o https://github.com/kubernetes/minikube/issues/7428
		imgs = append(imgs, images.KindNet(""))
	}

	runner := command.NewKICRunner(profile, driver.OCIBinary)

	// will need to do this to enable the container run-time service
	sv, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "Failed to parse Kubernetes version")
	}

	co := cruntime.Config{
		Type:              containerRuntime,
		Runner:            runner,
		ImageRepository:   "",
		KubernetesVersion: sv, //  this is just to satisfy cruntime and shouldnt matter what version.
	}
	cr, err := cruntime.New(co)
	if err != nil {
		return errors.Wrap(err, "failed create new runtime")
	}
	if err := cr.Enable(true, false); err != nil {
		return errors.Wrap(err, "enable container runtime")
	}

	for _, img := range imgs {
		pull := func() error {
			cmd := imagePullCommand(containerRuntime, img)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				time.Sleep(time.Second) // to avoid error: : exec: already started
				return errors.Wrapf(err, "pulling image %s", img)
			}
			return nil
		}
		// retry up to 5 times if network is bad
		if err = retry.Expo(pull, time.Microsecond, time.Minute, 5); err != nil {
			return errors.Wrapf(err, "pull image %s", img)
		}

	}

	// Transfer in k8s binaries
	kcfg := config.KubernetesConfig{
		KubernetesVersion: kubernetesVersion,
	}

	sm := sysinit.New(runner)

	if err := bsutil.TransferBinaries(kcfg, runner, sm); err != nil {
		return errors.Wrap(err, "transferring k8s binaries")
	}
	// Create image tarball
	if err := createImageTarball(tarballFilename, containerRuntime); err != nil {
		return errors.Wrap(err, "create tarball")
	}

	return copyTarballToHost(tarballFilename)
}

func verifyStorage(containerRuntime string) error {
	if containerRuntime == "docker" || containerRuntime == "containerd" {
		if err := verifyDockerStorage(); err != nil {
			return errors.Wrap(err, "Docker storage type is incompatible")
		}
	}
	if containerRuntime == "cri-o" {
		if err := verifyPodmanStorage(); err != nil {
			return errors.Wrap(err, "Podman storage type is incompatible")
		}
	}
	return nil
}

// returns the right command to pull image for a specific runtime
func imagePullCommand(containerRuntime, img string) *exec.Cmd {
	if containerRuntime == "docker" {
		return exec.Command("docker", "exec", profile, "docker", "pull", img)
	}

	if containerRuntime == "containerd" {
		return exec.Command("docker", "exec", profile, "sudo", "crictl", "pull", img)
	}

	if containerRuntime == "cri-o" {
		return exec.Command("docker", "exec", profile, "sudo", "crictl", "pull", img)
	}
	return nil
}

func createImageTarball(tarballFilename, containerRuntime string) error {
	// directories to save into tarball
	dirs := []string{
		"./lib/minikube/binaries",
	}

	if containerRuntime == "docker" {
		dirs = append(dirs, fmt.Sprintf("./lib/docker/%s", dockerStorageDriver), "./lib/docker/image")
	}

	if containerRuntime == "containerd" {
		dirs = append(dirs, "./lib/containerd")
	}

	if containerRuntime == "cri-o" {
		dirs = append(dirs, "./lib/containers")
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
	cmd := exec.Command(minikubePath, "delete", "-p", profile) // to avoid https://github.com/kubernetes/minikube/issues/7814
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
