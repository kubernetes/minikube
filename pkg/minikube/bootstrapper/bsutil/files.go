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

import (
	"path"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// KubeadmYamlPath is the path to the kubeadm configuration
var KubeadmYamlPath = path.Join(vmpath.GuestEphemeralDir, "kubeadm.yaml")

const (
	DefaultCNIConfigPath = "/etc/cni/net.d/k8s.conf"
	KubeletServiceFile   = "/lib/systemd/system/kubelet.service"
	// enum to differentiate kubeadm command line parameters from kubeadm config file parameters (see the
	// KubeadmExtraArgsWhitelist variable for more info)
	KubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
)

// ConfigFileAssets returns configuration file assets
func ConfigFileAssets(cfg config.KubernetesConfig, kubeadm []byte, kubelet []byte, kubeletSvc []byte, defaultCNIConfig []byte) []assets.CopyableFile {
	fs := []assets.CopyableFile{
		assets.NewMemoryAssetTarget(kubeadm, KubeadmYamlPath, "0640"),
		assets.NewMemoryAssetTarget(kubelet, KubeletSystemdConfFile, "0644"),
		assets.NewMemoryAssetTarget(kubeletSvc, KubeletServiceFile, "0644"),
	}
	// Copy the default CNI config (k8s.conf), so that kubelet can successfully
	// start a Pod in the case a user hasn't manually installed any CNI plugin
	// and minikube was started with "--extra-config=kubelet.network-plugin=cni".
	if defaultCNIConfig != nil {
		fs = append(fs, assets.NewMemoryAssetTarget(defaultCNIConfig, DefaultCNIConfigPath, "0644"))
	}
	return fs
}
