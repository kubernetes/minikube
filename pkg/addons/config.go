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

package addons

import (
	"k8s.io/minikube/pkg/minikube/config"
)

type setFn func(*config.ClusterConfig, string, string) error

// Addon represents an addon
type Addon struct {
	name        string
	set         func(*config.ClusterConfig, string, string) error
	validations []setFn
	callbacks   []setFn
}

// addonPodLabels holds the pod label that will be used to verify if the addon is enabled
var addonPodLabels = map[string]string{
	"ingress":             "app.kubernetes.io/name=ingress-nginx",
	"registry":            "kubernetes.io/minikube-addons=registry",
	"gvisor":              "kubernetes.io/minikube-addons=gvisor",
	"gcp-auth":            "kubernetes.io/minikube-addons=gcp-auth",
	"csi-hostpath-driver": "kubernetes.io/minikube-addons=csi-hostpath-driver",
}

// Addons is a list of all addons
var Addons = []*Addon{
	{
		name:      "auto-pause",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, enableOrDisableAutoPause},
	},

	{
		name:      "dashboard",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},

	{
		name:      "default-storageclass",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "efk",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "freshpod",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:        "gvisor",
		set:         SetBool,
		validations: []setFn{isRuntimeContainerd},
		callbacks:   []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "ingress",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "ingress-dns",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "istio-provisioner",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "istio",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "inspektor-gadget",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "kong",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "kubevirt",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "logviewer",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "metrics-server",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:        "nvidia-driver-installer",
		set:         SetBool,
		validations: []setFn{isKVMDriverForNVIDIA},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		// The nvidia-gpu-device-plugin addon is deprecated and it's functionality is merged inside of nvidia-device-plugin addon.
		name:        "nvidia-gpu-device-plugin",
		set:         SetBool,
		validations: []setFn{isKVMDriverForNVIDIA},
		callbacks:   []setFn{EnableOrDisableAddon},
	},
	{
		name:      "amd-gpu-device-plugin",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "olm",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "registry",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "registry-creds",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "registry-aliases",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
		// TODO - add other settings
		// TODO check if registry addon is enabled
	},
	{
		name:      "storage-provisioner",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "storage-provisioner-gluster",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "storage-provisioner-rancher",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "metallb",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "ambassador",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "pod-security-policy",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "gcp-auth",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableGCPAuth, EnableOrDisableAddon, verifyGCPAuthAddon},
	},
	{
		name:      "volcano",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "volumesnapshots",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:        "csi-hostpath-driver",
		set:         SetBool,
		validations: []setFn{isVolumesnapshotsEnabled},
		callbacks:   []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "portainer",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "inaccel",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "headlamp",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "cloud-spanner",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "kubeflow",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "nvidia-device-plugin",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
	{
		name:      "yakd",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
	},
}
