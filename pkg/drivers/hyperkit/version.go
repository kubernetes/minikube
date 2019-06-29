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

package hyperkit

// The current version of the docker-machine-driver-hyperkit

// version is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/drivers/hyperkit.version=vX.Y.Z"
var version = "v0.0.0-unset"

// gitCommitID is a private field and should be set when compiling with --ldflags="-X k8s.io/minikube/pkg/drivers/hyperkit.gitCommitID=<commit-id>"
var gitCommitID = ""

// GetVersion returns the current docker-machine-driver-hyperkit version
func GetVersion() string {
	return version
}

// GetGitCommitID returns the git commit id from which it is being built
func GetGitCommitID() string {
	return gitCommitID
}
