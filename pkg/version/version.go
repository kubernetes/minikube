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

package version

import (
	"strings"

	"github.com/blang/semver"
)

// VersionPrefix is the prefix of the git tag for a version
const VersionPrefix = "v"

// The current version of the minikube

// version is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.version=vX.Y.Z"
var version = "v0.0.0-unset"

// isoVersion is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.isoVersion=vX.Y.Z"
var isoVersion = "v0.0.0-unset"

var isoPath = "minikube/iso"

// GetVersion returns the current minikube version
func GetVersion() string {
	return version
}

// GetISOVersion returns the current minikube.iso version
func GetISOVersion() string {
	return isoVersion
}

// GetISOPath returns the remote path to the minikube.iso
func GetISOPath() string {
	return isoPath
}

// GetSemverVersion returns the current minikube semantic version (semver)
func GetSemverVersion() (semver.Version, error) {
	return semver.Make(strings.TrimPrefix(GetVersion(), VersionPrefix))
}
