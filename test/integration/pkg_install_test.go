// +build integration

/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var distros = []string{
	"debian:sid",
	"debian:latest",
	"debian:10",
	"debian:9",
	"ubuntu:latest",
	"ubuntu:20.10",
	"ubuntu:20.04",
	"ubuntu:18.04",
}

var timeout = Minutes(10)

// TestPackageInstall tests installation of .deb packages with minikube itself and with kvm2 driver
// on various debian/ubuntu docker images
func TestDebPackageInstall(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rr, err := Run(t, exec.CommandContext(ctx, "docker", "version"))
	if err != nil || rr.ExitCode != 0 {
		t.Skip("docker is not installed")
	}

	pkgDir, err := filepath.Abs(filepath.Dir(Target()))
	if err != nil {
		t.Errorf("failed to get minikube path: %v", err)
	}
	mkDebs, err := filepath.Glob(fmt.Sprintf("%s/minikube_*_%s.deb", pkgDir, runtime.GOARCH))
	if err != nil {
		t.Errorf("failed to find minikube deb in %q: %v", pkgDir, err)
	}
	kvmDebs, err := filepath.Glob(fmt.Sprintf("%s/docker-machine-driver-kvm2_*_%s.deb", pkgDir, runtime.GOARCH))
	if err != nil {
		t.Errorf("failed to find minikube deb in %q: %v", pkgDir, err)
	}

	for _, distro := range distros {
		distroImg := distro
		testName := fmt.Sprintf("install_%s_%s", runtime.GOARCH, distroImg)
		t.Run(testName, func(t *testing.T) {
			// apt-get update; dpkg -i minikube_${ver}_${arch}.deb
			t.Run("minikube", func(t *testing.T) {
				for _, mkDeb := range mkDebs {
					rr, err := dpkgInstall(ctx, t, distro, mkDeb)
					if err != nil || rr.ExitCode != 0 {
						t.Errorf("failed to install %q on %q: err=%v, exit=%d",
							mkDeb, distroImg, err, rr.ExitCode)
					}
				}
			})
			// apt-get update; apt-get install -y libvirt0; dpkg -i docker-machine-driver-kvm2_${ver}_${arch}.deb
			t.Run("kvm2-driver", func(t *testing.T) {
				for _, kvmDeb := range kvmDebs {
					rr, err := dpkgInstallDriver(ctx, t, distro, kvmDeb)
					if err != nil || rr.ExitCode != 0 {
						t.Errorf("failed to install %q on %q: err=%v, exit=%d",
							kvmDeb, distroImg, err, rr.ExitCode)
					}
				}
			})
		})
	}
}

func dpkgInstall(ctx context.Context, t *testing.T, image, deb string) (*RunResult, error) {
	return Run(t, exec.CommandContext(ctx,
		"docker", "run", "--rm", fmt.Sprintf("-v%s:/var/tmp", filepath.Dir(deb)),
		image,
		"sh", "-c", fmt.Sprintf("apt-get update; dpkg -i /var/tmp/%s", filepath.Base(deb))))
}

func dpkgInstallDriver(ctx context.Context, t *testing.T, image, deb string) (*RunResult, error) {
	return Run(t, exec.CommandContext(ctx,
		"docker", "run", "--rm", fmt.Sprintf("-v%s:/var/tmp", filepath.Dir(deb)),
		image,
		"sh", "-c", fmt.Sprintf("apt-get update; apt-get install -y libvirt0; dpkg -i /var/tmp/%s", filepath.Base(deb))))
}
