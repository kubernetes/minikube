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
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
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

// AddonOverrides is a list of all addons
var AddonOverrides = []*Addon{
	{
		name:      "auto-pause",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, enableOrDisableAutoPause},
	},
	{
		name:      "default-storageclass",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:        "gvisor",
		set:         SetBool,
		validations: []setFn{IsRuntimeContainerd},
		callbacks:   []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "ingress",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "metrics-server",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "registry",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
	{
		name:      "registry-aliases",
		set:       SetBool,
		callbacks: []setFn{EnableOrDisableAddon},
		// TODO - add other settings
		// TODO check if registry addon is enabled
	},
	{
		name:      "storage-provisioner-gluster",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableStorageClasses},
	},
	{
		name:      "gcp-auth",
		set:       SetBool,
		callbacks: []setFn{enableOrDisableGCPAuth, EnableOrDisableAddon, verifyGCPAuthAddon},
	},
	{
		name:        "csi-hostpath-driver",
		set:         SetBool,
		validations: []setFn{IsVolumesnapshotsEnabled},
		callbacks:   []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
}

type CustomRegistry struct {
	Path    string
	Enabled bool
}

type Configuration struct {
	CustomRegistries []CustomRegistry
}

func LoadAddonsConfig() (*Configuration, error) {
	contents, err := os.ReadFile(localpath.AddonsConfigFile())
	if err != nil {
		// The file is optional
		if errors.Is(err, os.ErrNotExist) {
			return &Configuration{}, nil
		}

		return nil, err
	}

	var config Configuration
	err = json.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
