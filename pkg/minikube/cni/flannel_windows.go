/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package cni

import _ "embed"

//go:embed flannel-windows.yaml
var flannelWindowsYaml string

//go:embed kube-proxy-windows.yaml
var kubeProxyWindowsYaml string

// FlannelWindowsManifest returns the Windows flannel DaemonSet manifest.
func FlannelWindowsManifest() string {
	return flannelWindowsYaml
}

// KubeProxyWindowsManifest returns the Windows kube-proxy DaemonSet manifest.
// The image is pinned to the latest sigwindowstools/kube-proxy flannel-hostprocess
// release; bump the tag in kube-proxy-windows.yaml when a newer version is published.
func KubeProxyWindowsManifest() string {
	return kubeProxyWindowsYaml
}
