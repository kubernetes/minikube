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

package config

// AddNode adds a new node config to an existing cluster.
func AddNode(cc *MachineConfig, name string, controlPlane bool, k8sVersion string, profileName string) error {
	node := Node{
		Name:   name,
		Worker: true,
	}

	if controlPlane {
		node.ControlPlane = true
	}

	if k8sVersion != "" {
		node.KubernetesVersion = k8sVersion
	}

	cc.Nodes = append(cc.Nodes, node)
	return SaveProfile(profileName, cc)
}
