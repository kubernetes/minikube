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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/constants"

	"github.com/blang/semver/v4"

	"k8s.io/minikube/pkg/version"
)

const (
	// builds a docker v2 repository API call in the format https://registry.k8s.io/v2/coredns/coredns/tags/list
	tagURLTemplate = "https://%s/v2/%s/tags/list"
)

// Pause returns the image name to pull for a given Kubernetes version
func Pause(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `PauseVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants_unix.go
	imageName := "pause"
	pv := imageVersion(v, mirror, imageName, "3.6")

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror, v), imageName), pv)
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
	return fmt.Sprintf("%s:v%s", path.Join(kubernetesRepo(mirror, v), name), v)
}

// fixes 13136 by getting the latest image version from the registry.k8s.io repository instead of hardcoded
func findLatestTagFromRepository(url string, lastKnownGood string) string {
	client := &http.Client{}
	errorMsg := fmt.Sprintf("Failed to get latest image version for %s, reverting to version %s.", url, lastKnownGood)

	resp, err := client.Get(url)

	if err != nil || resp.StatusCode != http.StatusOK {
		klog.Warningf("%s Error %v", errorMsg, err)
		return lastKnownGood
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Warningf("%s Error %v", errorMsg, err)
		return lastKnownGood
	}

	type TagsResponse struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	tags := TagsResponse{}
	err = json.Unmarshal(body, &tags)
	if err != nil || len(tags.Tags) < 1 {
		klog.Warningf("%s Error %v", errorMsg, err)
		return lastKnownGood
	}
	lastTagNum := len(tags.Tags) - 1
	return tags.Tags[lastTagNum]
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
	cv := imageVersion(v, mirror, imageName, "v1.8.6")

	if mirror == constants.AliyunMirror {
		imageName = "coredns"
	}

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror, v), imageName), cv)
}

// etcd returns the image used for etcd
func etcd(v semver.Version, mirror string) string {
	// Note: changing this logic requires bumping the preload version
	// Should match `DefaultEtcdVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	imageName := "etcd"
	ev := imageVersion(v, mirror, imageName, "3.5.0-0")

	return fmt.Sprintf("%s:%s", path.Join(kubernetesRepo(mirror, v), imageName), ev)
}

func imageVersion(v semver.Version, mirror, imageName, defaultVersion string) string {
	versionString := fmt.Sprintf("v%s", v.String())
	if ver, ok := constants.KubeadmImages[versionString][imageName]; ok {
		return ver
	}
	return findLatestTagFromRepository(fmt.Sprintf(tagURLTemplate, kubernetesRepo(mirror, v), imageName), defaultVersion)
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
		repo = "kindest"
	}
	return path.Join(repo, "kindnetd:v20221004-44d545d1")
}

// all calico images are from https://docs.projectcalico.org/manifests/calico.yaml
const calicoVersion = "v3.24.5"
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
