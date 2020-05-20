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
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/machine"
)

// TODO: Share these between cluster and node packages
const (
	mountString = "mount-string"
	createMount = "mount"
)

// Add adds a new node config to an existing cluster.
func Add(cc *config.ClusterConfig, n config.Node) error {
	if err := config.SaveNode(cc, &n); err != nil {
		return errors.Wrap(err, "save node")
	}

	r, p, m, h, err := Provision(cc, &n, false)
	if err != nil {
		return err
	}
	s := Starter{
		Runner:         r,
		PreExists:      p,
		MachineAPI:     m,
		Host:           h,
		Cfg:            cc,
		Node:           &n,
		ExistingAddons: nil,
	}

	_, err = Start(s, false)
	return err
}

// Delete stops and deletes the given node from the given cluster
func Delete(cc config.ClusterConfig, name string) error {
	n, index, err := Retrieve(&cc, name)
	if err != nil {
		return errors.Wrap(err, "retrieve")
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		return err
	}

	err = machine.DeleteHost(api, driver.MachineName(cc, *n))
	if err != nil {
		return err
	}

	cc.Nodes = append(cc.Nodes[:index], cc.Nodes[index+1:]...)
	return config.SaveProfile(viper.GetString(config.ProfileName), &cc)
}

// Retrieve finds the node by name in the given cluster
func Retrieve(cc *config.ClusterConfig, name string) (*config.Node, int, error) {
	for i, n := range cc.Nodes {
		if n.Name == name {
			return &n, i, nil
		}
	}

	return nil, -1, errors.New("Could not find node " + name)
}

// Save saves a node to a cluster
func Save(cfg *config.ClusterConfig, node *config.Node) error {
	update := false
	for i, n := range cfg.Nodes {
		if n.Name == node.Name {
			cfg.Nodes[i] = *node
			update = true
			break
		}
	}

	if !update {
		cfg.Nodes = append(cfg.Nodes, *node)
	}
	return config.SaveProfile(viper.GetString(config.ProfileName), cfg)
}

// Name returns the appropriate name for the node given the current number of nodes
func Name(index int) string {
	return fmt.Sprintf("m%02d", index)
}
