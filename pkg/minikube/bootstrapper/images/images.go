/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package images

import (
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
	minikubeVersion "k8s.io/minikube/pkg/version"
)

const (
	// DefaultImageRepo is the default repository for images
	DefaultImageRepo = "k8s.gcr.io"
	// DefaultMinikubeRepo is the default repository for minikube
	DefaultMinikubeRepo = "gcr.io/k8s-minikube"
)

// getImageRepositories returns either the k8s image registry on GCR or a mirror if specified
func getImageRepository(imageRepository string) string {
	if imageRepository == "" {
		imageRepository = DefaultImageRepo
	}
	if !strings.HasSuffix(imageRepository, "/") {
		imageRepository += "/"
	}

	return imageRepository
}

// getMinikubeRepository returns either the minikube image registry on GCR or a mirror if specified
func getMinikubeRepository(imageRepository string) string {
	minikubeRepository := imageRepository
	if minikubeRepository == "" {
		minikubeRepository = DefaultMinikubeRepo
	}
	if !strings.HasSuffix(minikubeRepository, "/") {
		minikubeRepository += "/"
	}

	return minikubeRepository
}

// CachedImages gets the images to cache for kubeadm for a version
func CachedImages(imageRepositoryStr string, kubernetesVersionStr string) []string {
	imageRepository := getImageRepository(imageRepositoryStr)
	minikubeRepository := getMinikubeRepository(imageRepositoryStr)

	v1_16plus := semver.MustParseRange(">=1.16.0")
	v1_14plus := semver.MustParseRange(">=1.14.0 <1.16.0")
	v1_13 := semver.MustParseRange(">=1.13.0 <1.14.0")
	v1_12 := semver.MustParseRange(">=1.12.0 <1.13.0")
	v1_11 := semver.MustParseRange(">=1.11.0 <1.12.0")
	v1_12plus := semver.MustParseRange(">=1.12.0")

	kubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesVersionStr, minikubeVersion.VersionPrefix))
	if err != nil {
		glog.Errorln("Error parsing version semver: ", err)
	}

	var images []string
	if v1_12plus(kubernetesVersion) {
		images = append(images, []string{
			imageRepository + "kube-proxy" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-scheduler" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-controller-manager" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-apiserver" + ArchTag(false) + kubernetesVersionStr,
		}...)
	} else {
		images = append(images, []string{
			imageRepository + "kube-proxy" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-scheduler" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-controller-manager" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-apiserver" + ArchTag(true) + kubernetesVersionStr,
		}...)
	}

	podInfraContainerImage := PauseImage(imageRepository, kubernetesVersionStr)
	if v1_16plus(kubernetesVersion) {
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.13",
			imageRepository + "etcd" + ArchTag(false) + "3.3.15-0",
			imageRepository + "coredns" + ArchTag(false) + "1.6.2",
		}...)

	} else if v1_14plus(kubernetesVersion) {
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.13",
			imageRepository + "etcd" + ArchTag(false) + "3.3.10",
			imageRepository + "coredns" + ArchTag(false) + "1.3.1",
		}...)

	} else if v1_13(kubernetesVersion) {
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(false) + "3.2.24",
			imageRepository + "coredns:1.2.6",
		}...)

	} else if v1_12(kubernetesVersion) {
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(false) + "3.2.24",
			imageRepository + "coredns:1.2.2",
		}...)

	} else if v1_11(kubernetesVersion) {
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(true) + "3.2.18",
			imageRepository + "coredns:1.1.3",
		}...)
	}

	images = append(images, []string{
		imageRepository + "kubernetes-dashboard" + ArchTag(true) + "v1.10.1",
		imageRepository + "kube-addon-manager" + ArchTag(false) + "v9.0",
		minikubeRepository + "storage-provisioner" + ArchTag(false) + "v1.8.1",
	}...)

	return images
}

// PauseImage returns the image name for pause image (for pod infra)
func PauseImage(imageRepositoryStr string, kubernetesVersionStr string) string {
	imageRepository := getImageRepository(imageRepositoryStr)

	v1_16plus := semver.MustParseRange(">=1.16.0")
	v1_14plus := semver.MustParseRange(">=1.14.0 <1.16.0")
	v1_13 := semver.MustParseRange(">=1.13.0 <1.14.0")
	v1_12 := semver.MustParseRange(">=1.12.0 <1.13.0")
	v1_11 := semver.MustParseRange(">=1.11.0 <1.12.0")

	kubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesVersionStr, minikubeVersion.VersionPrefix))
	if err != nil {
		glog.Errorln("Error parsing version semver: ", err)
	}

	var podInfraContainerImage string
	switch {
	case v1_16plus(kubernetesVersion):
		podInfraContainerImage = imageRepository + "pause:3.1"

	case v1_14plus(kubernetesVersion):
		podInfraContainerImage = imageRepository + "pause:3.1"

	case v1_13(kubernetesVersion):
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.1"

	case v1_12(kubernetesVersion):
		podInfraContainerImage = imageRepository + "pause:3.1"

	case v1_11(kubernetesVersion):
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.1"

	default:
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.0"
	}

	return podInfraContainerImage
}

// ArchTag returns the archtag for images
func ArchTag(hasTag bool) string {
	if runtime.GOARCH == "amd64" && !hasTag {
		return ":"
	}
	return "-" + runtime.GOARCH + ":"
}
