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
	"fmt"
	"os/exec"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/driver"
	"sigs.k8s.io/yaml"
)

// minHyperkitVersion is the minimum version of the minikube hyperkit driver compatible with the current minikube code
var minHyperkitVersion *semver.Version

const minHyperkitVersionStr = "1.11.0"

// Version is auxiliary driver version info.
type Version struct {
	// Version is driver version in vX.Y.Z format.
	Version string `json:"version"`
	// Commit is the git commit hash the driver was built from.
	Commit string `json:"commit"`
}

func init() {
	v, err := semver.New(minHyperkitVersionStr)
	if err != nil {
		klog.Errorf("Failed to parse the hyperkit driver version: %v", err)
	} else {
		minHyperkitVersion = v
	}
}

// minAcceptableDriverVersion is the minimum version of driver supported by current version of minikube
func minAcceptableDriverVersion(driverName string, mkVer semver.Version) semver.Version {
	switch driverName {
	case driver.HyperKit:
		if minHyperkitVersion != nil {
			return *minHyperkitVersion
		}
		return mkVer
	case driver.KVM2:
		return mkVer
	default:
		klog.Warningf("Unexpected driver: %v", driverName)
		return mkVer
	}
}

// driverVersion returns auxiliary driver version.
func driverVersion(path string) (Version, error) {
	v := Version{}

	cmd := exec.Command(path, "version")
	output, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = ee.Stderr
		}
		return v, fmt.Errorf("command %s failed: %v: %s", cmd, err, stderr)
	}

	if err := yaml.Unmarshal(output, &v); err != nil {
		return v, fmt.Errorf("invalid driver version: %q: %v", output, err)
	}

	// Version is required to validate the driver in runtime.
	if v.Version == "" {
		return v, fmt.Errorf("version not specified: %+v", v)
	}

	// Commit is required to validate the driver during tests.
	if v.Commit == "" {
		return v, fmt.Errorf("commit not specified: %+v", v)
	}

	return v, nil
}
