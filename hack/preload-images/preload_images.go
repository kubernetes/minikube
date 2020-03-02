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
	"path"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/vmpath"
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
	tarballFilename = fmt.Sprintf("preloaded-images-k8s-%s.tar.lz4", kubernetesVersion)
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
		return err
	}
	if err := driver.Create(); err != nil {
		return errors.Wrap(err, "creating kic driver")
	}

	// Now, get images to pull
	imgs, err := images.Kubeadm("", kubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "kubeadm images")
	}

	for _, img := range append(imgs, kic.OverlayImage) {
		cmd := exec.Command("docker", "exec", profile, "docker", "pull", img)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "downloading %s", img)
		}
	}
	// Transfer in binaries
	var g errgroup.Group
	dir := filepath.Join(vmpath.GuestPersistentDir, "binaries")
	mkdirCmd := exec.Command("docker", "exec", profile, "mkdir", "-p", dir)
	if err := mkdirCmd.Run(); err != nil {
		return err
	}

	for _, name := range constants.KubernetesReleaseBinaries {
		name := name
		g.Go(func() error {
			src, err := machine.CacheBinary(name, kubernetesVersion, "linux", runtime.GOARCH)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", name)
			}

			dst := path.Join(dir, name)
			copyCmd := exec.Command("docker", "cp", src, fmt.Sprintf("%s:%s", profile, dst))
			copyCmd.Stdout = os.Stdout
			copyCmd.Stderr = os.Stderr
			fmt.Println(copyCmd.Args)
			if err := copyCmd.Run(); err != nil {
				return errors.Wrapf(err, "copybinary %s -> %s", src, dst)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap(err, "pulling binaries")
	}

	// Delete /var/lib/docker/network
	if err := deleteDirInMinikube("/var/lib/docker/network"); err != nil {
		return errors.Wrap(err, "deleting dir")
	}

	// Create image tarball
	if err := createImageTarball(); err != nil {
		return err
	}
	return copyTarballToHost()
}

func startMinikube() error {
	cmd := exec.Command(minikubePath, "start", "-p", profile, "--memory", "4000", "--kubernetes-version", kubernetesVersion, "--vm-driver=docker")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func createImageTarball() error {
	cmd := exec.Command("docker", "exec", profile, "sudo", "tar", "-I", "lz4", "-C", "/var/lib/docker", "-cvf", tarballFilename, "./")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "creating image tarball")
	}
	return nil
}

func deleteDirInMinikube(dir string) error {
	cmd := exec.Command("docker", "exec", profile, "sudo", "rm", "-rf", dir)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "deleting %s", dir)
	}
	return nil
}

func copyTarballToHost() error {
	dest := filepath.Join("out/", tarballFilename)
	cmd := exec.Command("docker", "cp", fmt.Sprintf("%s:/%s", profile, tarballFilename), dest)
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
