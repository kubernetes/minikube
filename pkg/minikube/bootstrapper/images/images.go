/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// Package images implements helpers for getting image names
package images

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"

	"github.com/blang/semver/v4"

	"k8s.io/minikube/pkg/version"
)

// Pause returns the image name to pull for a given Kubernetes version
func Pause(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `PauseVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants_unix.go
	imageName := "pause"
	pv := imageVersion(v, imageName, "3.9")

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror), imageName), pv)
}

// essentials returns images needed too bootstrap a Kubernetes
func essentials(mirror string, v semver.Version) []string {
	imgs := []string{
		// use the same order as: `kubeadm config images list`
		componentImage("kube-apiserver", v, mirror),
		componentImage("kube-controller-manager", v, mirror),
		componentImage("kube-scheduler", v, mirror),
		componentImage("kube-proxy", v, mirror),
		Pause(v, mirror),
		etcd(v, mirror),
		coreDNS(v, mirror),
	}
	return imgs
}

// componentImage returns a Kubernetes component image to pull
func componentImage(name string, v semver.Version, mirror string) string {
	return fmt.Sprintf("%s:v%s", path.Join(kubernetesRepo(mirror), name), v)
}

// tagFromKubeadm gets the image tag by running kubeadm image list command on the host machine (Linux only)
func tagFromKubeadm(v, name string) (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("can only get tag from kubeadm on Linux")
	}
	kubeadm, err := download.Binary("kubeadm", v, "linux", runtime.GOARCH, "")
	if err != nil {
		return "", fmt.Errorf("failed to download kubeadm binary: %v", err)
	}
	// TODO: Once kubeadm graduates the "-experimental-output" flag to non-experimental should use JSON output here
	b, err := exec.Command(kubeadm, "config", "images", "list").Output()
	if err != nil {
		return "", fmt.Errorf("failed getting kubeadm image list: %v", err)
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if !strings.Contains(line, name) {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("unexpected image format: %s", line)
		}
		return parts[1], nil
	}
	return "", fmt.Errorf("failed to find %q image in kubeadm image list", name)
}

// tagFromLastMinor finds the last matching minor version in the kubeadm images map and uses its image version
func tagFromLastMinor(v semver.Version, name, lastKnownGood string) string {
	majorMinor := fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	var latestMinorVer string
	for _, existingVer := range constants.ValidKubernetesVersions {
		if !strings.HasPrefix(existingVer, majorMinor) {
			continue
		}
		latestMinorVer = existingVer
		break
	}
	tag, ok := constants.KubeadmImages[latestMinorVer][name]
	if !ok {
		return lastKnownGood
	}
	return tag
}

// coreDNS returns the images used for CoreDNS
func coreDNS(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `CoreDNSImageName` and `CoreDNSVersion` in
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go

	imageName := "coredns/coredns"
	if semver.MustParseRange("<1.21.0-alpha.1")(v) {
		imageName = "coredns"
	}
	cv := imageVersion(v, imageName, "v1.10.1")

	if mirror == constants.AliyunMirror {
		imageName = "coredns"
	}

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror), imageName), cv)
}

// etcd returns the image used for etcd
func etcd(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `DefaultEtcdVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	imageName := "etcd"
	ev := imageVersion(v, imageName, "3.5.7-0")

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror), imageName), ev)
}

func imageVersion(v semver.Version, imageName, defaultVersion string) string {
	versionString := fmt.Sprintf("v%s", v.String())
	if ver, ok := constants.KubeadmImages[versionString][imageName]; ok {
		return ver
	}
	if ver, err := tagFromKubeadm(versionString, imageName); err == nil {
		return ver
	}
	return tagFromLastMinor(v, imageName, defaultVersion)
}

// auxiliary returns images that are helpful for running minikube
func auxiliary(mirror string) []string {
	// Note: changing this list requires bumping the preload version
	return []string{
		storageProvisioner(mirror),
		// NOTE: kindnet is also used when the Docker driver is used with a non-Docker runtime
	}
}

// storageProvisioner returns the minikube storage provisioner image
func storageProvisioner(mirror string) string {
	cv := version.GetStorageProvisionerVersion()
	in := "k8s-minikube/storage-provisioner:" + cv
	if mirror == "" {
		mirror = "gcr.io"
	} else if mirror == constants.AliyunMirror {
		in = "storage-provisioner:" + cv
	}
	return path.Join(mirror, in)
}

// KindNet returns the image used for kindnet
// ref: https://hub.docker.com/r/kindest/kindnetd/tags
// src: https://github.com/kubernetes-sigs/kind/tree/master/images/kindnetd
func KindNet(repo string) string {
	if repo == "" {
		repo = "docker.io/kindest"
	}
	return path.Join(repo, "kindnetd:v20241023-a345ebe4")
}

// all calico images are from https://github.com/projectcalico/calico/blob/master/manifests/calico.yaml
const calicoVersion = "v3.29.0"
const calicoRepo = "docker.io/calico"

// CalicoDaemonSet returns the image used for calicoDaemonSet
func CalicoDaemonSet(repo string) string {
	return calicoCommon(repo, "node")

}

// CalicoDeployment returns the image used for calicoDeployment
func CalicoDeployment(repo string) string {
	return calicoCommon(repo, "kube-controllers")
}

// CalicoBin returns image used for calico binary image
func CalicoBin(repo string) string {
	return calicoCommon(repo, "cni")
}

func calicoCommon(repo string, name string) string {
	if repo == "" {
		repo = calicoRepo
	}
	return path.Join(repo, fmt.Sprintf("%s:%s", name, calicoVersion))
}
