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
	"path"

	"k8s.io/minikube/pkg/minikube/constants"

	"github.com/blang/semver/v4"

	"k8s.io/minikube/pkg/version"
)

// Pause returns the image name to pull for a given Kubernetes version
func Pause(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `PauseVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants_unix.go
	pv := "3.6"
	majorMinorVersion := fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	imageName := "pause"
	if pVersion, ok := constants.KubeadmImages[majorMinorVersion][imageName]; ok {
		pv = pVersion
	}

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

// coreDNS returns the images used for CoreDNS
func coreDNS(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `CoreDNSImageName` and `CoreDNSVersion` in
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go

	cv := "1.8.4"
	in := "coredns/coredns"
	if semver.MustParseRange("<1.21.0-alpha.1")(v) {
		in = "coredns"
	}

	majorMinorVersion := fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	if cVersion, ok := constants.KubeadmImages[majorMinorVersion][in]; ok {
		cv = cVersion
	}

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror), in), cv)
}

// etcd returns the image used for etcd
func etcd(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `DefaultEtcdVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	ev := "3.5.0-0"
	majorMinorVersion := fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	imageName := "etcd"
	if eVersion, ok := constants.KubeadmImages[majorMinorVersion][imageName]; ok {
		ev = eVersion
	}

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror), imageName), ev)
}

// auxiliary returns images that are helpful for running minikube
func auxiliary(mirror string) []string {
	// Note: changing this list requires bumping the preload version
	return []string{
		storageProvisioner(mirror),
		dashboardFrontend(mirror),
		dashboardMetrics(mirror),
		// NOTE: kindnet is also used when the Docker driver is used with a non-Docker runtime
	}
}

// storageProvisioner returns the minikube storage provisioner image
func storageProvisioner(mirror string) string {
	return path.Join(minikubeRepo(mirror), "storage-provisioner:"+version.GetStorageProvisionerVersion())
}

// dashboardFrontend returns the image used for the dashboard frontend
func dashboardFrontend(repo string) string {
	if repo == "" {
		repo = "docker.io"
	}
	// See 'kubernetes-dashboard' in deploy/addons/dashboard/dashboard-dp.yaml
	return path.Join(repo, "kubernetesui", "dashboard:v2.3.1")
}

// dashboardMetrics returns the image used for the dashboard metrics scraper
func dashboardMetrics(repo string) string {
	if repo == "" {
		repo = "docker.io"
	}
	// See 'dashboard-metrics-scraper' in deploy/addons/dashboard/dashboard-dp.yaml
	return path.Join(repo, "kubernetesui", "metrics-scraper:v1.0.7")
}

// KindNet returns the image used for kindnet
// ref: https://hub.docker.com/r/kindest/kindnetd/tags
// src: https://github.com/kubernetes-sigs/kind/tree/master/images/kindnetd
func KindNet(repo string) string {
	if repo == "" {
		repo = "kindest"
	}
	return path.Join(repo, "kindnetd:v20210326-1e038dc5")
}

// all calico images are from https://docs.projectcalico.org/manifests/calico.yaml
const calicoVersion = "v3.20.0"
const calicoRepo = "docker.io/calico"

// CalicoDaemonSet returns the image used for calicoDaemonSet
func CalicoDaemonSet(repo string) string {
	return calicoCommon(repo, "node")

}

// CalicoDeployment returns the image used for calicoDeployment
func CalicoDeployment(repo string) string {
	return calicoCommon(repo, "kube-controllers")
}

// CalicoFelixDriver returns image used for felix driver
func CalicoFelixDriver(repo string) string {
	return calicoCommon(repo, "pod2daemon-flexvol")
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
