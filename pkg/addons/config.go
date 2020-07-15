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
	"k8s.io/minikube/pkg/addons/gcpauth"
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
	"ingress":  "app.kubernetes.io/name=ingress-nginx",
	"registry": "kubernetes.io/minikube-addons=registry",
	"gvisor":   "kubernetes.io/minikube-addons=gvisor",
	"gcp-auth": "kubernetes.io/minikube-addons=gcp-auth",
}

// Addons is a list of all addons
var Addons = []*Addon{
	{
		name:      "dashboard",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},

	{
		name:      "default-storageclass",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "efk",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "freshpod",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:        "gvisor",
		set:         SetBool,
		validations: []setFn{IsRuntimeContainerd},
		callbacks:   []setFn{enableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "helm-tiller",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "ingress",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "ingress-dns",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "istio-provisioner",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "istio",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "kubevirt",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "logviewer",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "metrics-server",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "nvidia-driver-installer",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "nvidia-gpu-device-plugin",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "olm",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "registry",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "registry-creds",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "registry-aliases",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
		//TODO - add other settings
		//TODO check if registry addon is enabled
	},
	{
		name:      "storage-provisioner",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "storage-provisioner-gluster",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "metallb",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "ambassador",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "pod-security-policy",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableAddon},
	},
	{
		name:      "gcp-auth",
		set:       SetBool,
		callbacks: []setFn{gcpauth.EnableOrDisable, enableOrDisableAddon, verifyGCPAuthAddon, gcpauth.DisplayAddonMessage},
	},
}
