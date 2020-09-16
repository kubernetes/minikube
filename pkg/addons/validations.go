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
	"fmt"
	"strconv"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/out"
)

const volumesnapshotsAddon = "volumesnapshots"

// containerdOnlyMsg is the message shown when a containerd-only addon is enabled
const containerdOnlyAddonMsg = `
This addon can only be enabled with the containerd runtime backend. To enable this backend, please first stop minikube with:

minikube stop

and then start minikube again with the following flags:

minikube start --container-runtime=containerd --docker-opt containerd=/var/run/containerd/containerd.sock`

// volumesnapshotsDisabledMsg is the message shown when csi-hostpath-driver addon is enabled without the volumesnapshots addon
const volumesnapshotsDisabledMsg = `[WARNING] For full functionality, the 'csi-hostpath-driver' addon requires the 'volumesnapshots' addon to be enabled.

You can enable 'volumesnapshots' addon by running: 'minikube addons enable volumesnapshots'
`

// IsRuntimeContainerd is a validator which returns an error if the current runtime is not containerd
func IsRuntimeContainerd(cc *config.ClusterConfig, _, _ string) error {
	r, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime})
	if err != nil {
		return err
	}
	_, ok := r.(*cruntime.Containerd)
	if !ok {
		return fmt.Errorf(containerdOnlyAddonMsg)
	}
	return nil
}

// IsVolumesnapshotsEnabled is a validator that prints out a warning if the volumesnapshots addon
// is disabled (does not return any errors!)
func IsVolumesnapshotsEnabled(cc *config.ClusterConfig, _, value string) error {
	isCsiDriverEnabled, _ := strconv.ParseBool(value)
	// assets.Addons[].IsEnabled() returns the current status of the addon or default value.
	// config.AddonList contains list of addons to be enabled.
	isVolumesnapshotsEnabled := assets.Addons[volumesnapshotsAddon].IsEnabled(cc) || contains(config.AddonList, volumesnapshotsAddon)
	if isCsiDriverEnabled && !isVolumesnapshotsEnabled {
		// just print out a warning directly, we don't want to return any errors since
		// that would prevent the addon from being enabled (callbacks wouldn't be run)
		out.WarningT(volumesnapshotsDisabledMsg)
	}
	return nil
}

// isAddonValid returns the addon, true if it is valid
// otherwise returns nil, false
func isAddonValid(name string) (*Addon, bool) {
	for _, a := range Addons {
		if a.name == name {
			return a, true
		}
	}
	return nil, false
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
