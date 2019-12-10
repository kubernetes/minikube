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

package images

import (
	"fmt"
	"path"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

// Kubeadm returns a list of images necessary to bootstrap kubeadm
func Kubeadm(mirror string, version string) ([]string, error) {
	v, err := semver.Make(strings.TrimPrefix(version, "v"))
	if err != nil {
		return nil, errors.Wrap(err, "semver")
	}

	imgs := []string{
		componentImage("kube-proxy", v, mirror),
		componentImage("kube-scheduler", v, mirror),
		componentImage("kube-controller-manager", v, mirror),
		componentImage("kube-apiserver", v, mirror),
		coreDNS(v, mirror),
		etcd(v, mirror),
		Pause(mirror),
	}
	imgs = append(imgs, Auxiliary(mirror)...)
	return imgs, nil
}

// componentImage returns a Kubernetes component image to pull
func componentImage(name string, v semver.Version, mirror string) string {
	needsArchSuffix := false
	ancient := semver.MustParseRange("<1.12.0")
	if ancient(v) {
		needsArchSuffix = true
	}

	return fmt.Sprintf("%sv%s", path.Join(KubernetesRepo(mirror), name+ArchTag(needsArchSuffix)), v)
}

// coreDNS returns the images used for CoreDNS
func coreDNS(v semver.Version, mirror string) string {
	// Should match `CoreDNSVersion` in
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	cv := "1.6.5"
	switch v.Minor {
	case 16:
		cv = "1.6.2"
	case 15, 14:
		cv = "1.3.1"
	case 13:
		cv = "1.2.6"
	case 12:
		cv = "1.2.2"
	case 11:
		cv = "1.1.3"
	}
	return path.Join(KubernetesRepo(mirror), "coredns"+ArchTag(false)+cv)
}

// etcd returns the image used for etcd
func etcd(v semver.Version, mirror string) string {
	needsArchSuffix := false
	ancient := semver.MustParseRange("<1.12.0")
	if ancient(v) {
		needsArchSuffix = true
	}

	// Should match `DefaultEtcdVersion` in:
	// https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/constants/constants.go
	ev := "3.4.3-0"
	switch v.Minor {
	case 16:
		ev = "3.3.15-0"
	case 14, 15:
		ev = "3.3.10"
	case 12, 13:
		ev = "3.2.24"
	case 11:
		ev = "3.2.18"
	}
	return path.Join(KubernetesRepo(mirror), "etcd"+ArchTag(needsArchSuffix)+ev)
}
