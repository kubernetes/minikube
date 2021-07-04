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
	"context"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
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
func Add(cc *config.ClusterConfig, n config.Node, delOnFail bool) error {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		return err
	}

	machineName := config.MachineName(*cc, n)
	for _, p := range profiles {
		if p.Config.Name == cc.Name {
			continue
		}

		for _, existNode := range p.Config.Nodes {
			if machineName == config.MachineName(*p.Config, existNode) {
				return errors.Errorf("Node %s already exists in %s profile", machineName, p.Name)
			}
		}
	}

	if err := config.SaveNode(cc, &n); err != nil {
		return errors.Wrap(err, "save node")
	}

	r, p, m, h, err := Provision(cc, &n, false, delOnFail)
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

// drainNode drains then deletes (removes) node from cluster.
func drainNode(cc config.ClusterConfig, name string) (*config.Node, error) {
	n, _, err := Retrieve(cc, name)
	if err != nil {
		return n, errors.Wrap(err, "retrieve")
	}

	m := config.MachineName(cc, *n)
	api, err := machine.NewAPIClient()
	if err != nil {
		return n, err
	}

	// grab control plane to use kubeconfig
	host, err := machine.LoadHost(api, cc.Name)
	if err != nil {
		return n, err
	}

	runner, err := machine.CommandRunner(host)
	if err != nil {
		return n, err
	}

	// kubectl drain with extra options to prevent ending up stuck in the process
	// ref: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#drain
	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	cmd := exec.Command("sudo", "KUBECONFIG=/var/lib/minikube/kubeconfig", kubectl, "drain", m,
		"--force", "--grace-period=1", "--skip-wait-for-delete-timeout=1", "--disable-eviction", "--ignore-daemonsets", "--delete-emptydir-data", "--delete-local-data")
	if _, err := runner.RunCmd(cmd); err != nil {
		klog.Warningf("unable to drain node %q: %v", name, err)
	} else {
		klog.Infof("successfully drained node %q", name)
	}

	// kubectl delete
	client, err := kapi.Client(cc.Name)
	if err != nil {
		return n, err
	}

	// set 'GracePeriodSeconds: 0' option to delete node immediately (ie, w/o waiting)
	var grace *int64
	err = client.CoreV1().Nodes().Delete(context.Background(), m, v1.DeleteOptions{GracePeriodSeconds: grace})
	if err != nil {
		klog.Errorf("unable to delete node %q: %v", name, err)
		return n, err
	}
	klog.Infof("successfully deleted node %q", name)

	return n, nil
}

// Delete calls drainNode to remove node from cluster and deletes the host.
func Delete(cc config.ClusterConfig, name string) (*config.Node, error) {
	n, err := drainNode(cc, name)
	if err != nil {
		return n, err
	}

	m := config.MachineName(cc, *n)
	api, err := machine.NewAPIClient()
	if err != nil {
		return n, err
	}

	err = machine.DeleteHost(api, m)
	if err != nil {
		return n, err
	}

	_, index, err := Retrieve(cc, name)
	if err != nil {
		return n, errors.Wrap(err, "retrieve")
	}

	cc.Nodes = append(cc.Nodes[:index], cc.Nodes[index+1:]...)
	return n, config.SaveProfile(viper.GetString(config.ProfileName), &cc)
}

// Retrieve finds the node by name in the given cluster
func Retrieve(cc config.ClusterConfig, name string) (*config.Node, int, error) {
	if driver.BareMetal(cc.Driver) {
		name = "m01"
	}

	for i, n := range cc.Nodes {
		if n.Name == name {
			return &n, i, nil
		}

		// Accept full machine name as well as just node name
		if config.MachineName(cc, n) == name {
			klog.Infof("Couldn't find node name %s, but found it as a machine name, returning it anyway.", name)
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
