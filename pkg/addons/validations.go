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

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

// containerdOnlyMsg is the message shown when a containerd-only addon is enabled
const containerdOnlyAddonMsg = `
This addon can only be enabled with the containerd runtime backend. To enable this backend, please first stop minikube with:

minikube stop

and then start minikube again with the following flags:

minikube start --container-runtime=containerd --docker-opt containerd=/var/run/containerd/containerd.sock`

// IsContainerdRuntime is a validator which returns an error if the current runtime is not containerd
func IsContainerdRuntime(_, _, profile string) error {
	config, err := config.Load(profile)
	if err != nil {
		return fmt.Errorf("config.Load: %v", err)
	}
	r, err := cruntime.New(cruntime.Config{Type: config.KubernetesConfig.ContainerRuntime})
	if err != nil {
		return err
	}
	_, ok := r.(*cruntime.Containerd)
	if !ok {
		return fmt.Errorf(containerdOnlyAddonMsg)
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
