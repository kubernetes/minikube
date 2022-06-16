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

	"github.com/blang/semver/v4"
)

// VersionPrefix is the prefix of the git tag for a version
const VersionPrefix = "v"

// version is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.version=vX.Y.Z"
var version = "v0.0.0-unset"

// gitCommitID is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.gitCommitID=<commit-id>"
var gitCommitID = ""

// isoVersion is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.isoVersion=vX.Y.Z"
var isoVersion = "v0.0.0-unset"

// storageProvisionerVersion is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/version.storageProvisionerVersion=<storage-provisioner-version>"
var storageProvisionerVersion = ""

// GetVersion returns the current minikube version
func GetVersion() string {
	return version
}

// GetGitCommitID returns the git commit id from which it is being built
func GetGitCommitID() string {
	return gitCommitID
}

// GetISOVersion returns the current minikube.iso version
func GetISOVersion() string {
	return isoVersion
}

// GetSemverVersion returns the current minikube semantic version (semver)
func GetSemverVersion() (semver.Version, error) {
	return semver.Make(strings.TrimPrefix(GetVersion(), VersionPrefix))
}

// GetStorageProvisionerVersion returns the storage provisioner version
func GetStorageProvisionerVersion() string {
	return storageProvisionerVersion
}
