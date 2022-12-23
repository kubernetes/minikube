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
	"github.com/blang/semver/v4"
)

// OldDefaultKubernetesRepo is the old default Kubernetes repository
const OldDefaultKubernetesRepo = "k8s.gcr.io"

// NewDefaultKubernetesRepo is the new default Kubernetes repository
const NewDefaultKubernetesRepo = "registry.k8s.io"

// kubernetesRepo returns the official Kubernetes repository, or an alternate
func kubernetesRepo(mirror string, v semver.Version) string {
	if mirror != "" {
		return mirror
	}
	return DefaultKubernetesRepo(v)
}

// rangeVersion returns the repository to use, if the version is within range
func rangeVersion(k8sRelease string, curVersion, newVersion semver.Version) (bool, string) {
	k8sVersion, err := semver.Make(k8sRelease + ".0")
	if err != nil {
		return false, ""
	}
	if k8sVersion.Major != curVersion.Major {
		return false, ""
	} else if k8sVersion.Minor != curVersion.Minor {
		return false, ""
	}

	if curVersion.GTE(newVersion) {
		return true, NewDefaultKubernetesRepo
	}
	return true, OldDefaultKubernetesRepo
}

func DefaultKubernetesRepo(kv semver.Version) string {
	// 1.21 and earlier
	if kv.LT(semver.MustParse("1.22.0-alpha.0")) {
		return OldDefaultKubernetesRepo
	}
	// 1.26 and later
	if kv.GTE(semver.MustParse("1.26.0-alpha.0")) {
		return NewDefaultKubernetesRepo
	}

	// various versions
	if ok, repo := rangeVersion("1.22", kv, semver.MustParse("1.22.18-rc.0")); ok {
		return repo
	}
	if ok, repo := rangeVersion("1.23", kv, semver.MustParse("1.23.15-rc.0")); ok {
		return repo
	}
	if ok, repo := rangeVersion("1.24", kv, semver.MustParse("1.24.9-rc.0")); ok {
		return repo
	}
	if ok, repo := rangeVersion("1.25", kv, semver.MustParse("1.25.0-alpha.1")); ok {
		return repo
	}
	// should not happen
	return NewDefaultKubernetesRepo
}
