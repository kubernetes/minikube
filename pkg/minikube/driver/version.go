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

package driver

import (
	"github.com/blang/semver"
	"github.com/golang/glog"
)

// minHyperkitVersion is the minimum version of the minikube hyperkit driver compatible with the current minikube code
var minHyperkitVersion *semver.Version

const minHyperkitVersionStr = "1.11.0"

func init() {
	v, err := semver.New(minHyperkitVersionStr)
	if err != nil {
		glog.Errorf("Failed to parse the hyperkit driver version: %v", err)
	} else {
		minHyperkitVersion = v
	}
}

// minAcceptableDriverVersion is the minimum version of driver supported by current version of minikube
func minAcceptableDriverVersion(driver string, mkVer semver.Version) semver.Version {
	switch driver {
	case HyperKit:
		if minHyperkitVersion != nil {
			return *minHyperkitVersion
		}
		return mkVer
	case KVM2:
		return mkVer
	default:
		glog.Warningf("Unexpected driver: %v", driver)
		return mkVer
	}
}
