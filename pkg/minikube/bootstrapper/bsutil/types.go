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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

// SkipAdditionalPreflights are additional preflights we skip depending on the runtime in use.
var SkipAdditionalPreflights = map[string][]string{}

// expectedRemoteArtifacts remote artifacts that must exist for minikube to function properly. The sign of a previously working installation.
// NOTE: /etc is not persistent across restarts, so don't bother checking there
var expectedRemoteArtifacts = []string{
	"/var/lib/kubelet/kubeadm-flags.env",
	"/var/lib/kubelet/config.yaml",
	EtcdDataDir(),
}
