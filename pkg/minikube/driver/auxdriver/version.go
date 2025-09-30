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

package auxdriver

import (
	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/driver"
)

// minAcceptableDriverVersion is the minimum version of driver supported by current version of minikube
func minAcceptableDriverVersion(driverName string, mkVer semver.Version) semver.Version {
	switch driverName {
	case driver.KVM2:
		return mkVer
	default:
		klog.Warningf("Unexpected driver: %v", driverName)
		return mkVer
	}
}
