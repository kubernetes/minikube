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

package cluster

import "k8s.io/minikube/pkg/minikube/config"

// AddNode adds a new node config to a cluster.
func AddNode(cfg config.MachineConfig, name string) error {

}

// StartNode starts the given node
func StartNode(node config.Node) error {

}

// StopNode stops the given node
func StopNode(node config.Node) error {

}

// DeleteNode deletes the given node from the given cluster
func DeleteNode(cfg config.MachineConfig, node config.Node) error {

}
