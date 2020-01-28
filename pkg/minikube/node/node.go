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

package node

import (
	"errors"

	"k8s.io/minikube/pkg/minikube/config"
)

// Add adds a new node config to an existing cluster.
func Add(cc *config.MachineConfig, name string, controlPlane bool, k8sVersion string, profileName string) error {
	n := config.Node{
		Name:   name,
		Worker: true,
	}

	if controlPlane {
		n.ControlPlane = true
	}

	if k8sVersion != "" {
		n.KubernetesVersion = k8sVersion
	}

	cc.Nodes = append(cc.Nodes, n)
	return config.SaveProfile(profileName, cc)
}

// Delete stops and deletes the given node from the given cluster
func Delete(cc *config.MachineConfig, name string) error {
	nd := nil
	for _, n := range cc.Nodes {
		if n.Name == name {
			nd = n
			break
		}
	}

	if nd == nil {
		return errors.New("Could not find node " + name)
	}
}

func Stop(cc *config.MachineConfig, name string) error {
	return nil
}
