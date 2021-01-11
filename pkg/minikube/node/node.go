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
	"os/exec"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/reason"
)

// TODO: Share these between cluster and node packages
const (
	mountString = "mount-string"
	createMount = "mount"
)

// Add adds a new node config to an existing cluster.
func Add(cc *config.ClusterConfig, n config.Node, delOnFail bool) error {
	if err := config.SaveNode(cc, &n); err != nil {
		return errors.Wrap(err, "save node")
	}

	r, p, m, h, err := Provision(cc, &n, delOnFail)
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

	_, err = Start(s)
	return err
}

// Delete stops and deletes the given node from the given cluster
func Delete(cc config.ClusterConfig, name string) (*config.Node, error) {
	n, index, err := Retrieve(cc, name)
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

	// leave master
	if n.ControlPlane {
		resetCmd := fmt.Sprintf("%s reset -f", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion))
		rc := exec.Command("/bin/bash", "-c", resetCmd)
		_, err = runner.RunCmd(rc)
		if err != nil {
			klog.Errorf("Failed to reset kubeadm", err)
		}
	}

	// kubectl drain
	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	cmd := exec.Command("sudo", "KUBECONFIG=/var/lib/minikube/kubeconfig", kubectl, "drain", m)
	if _, err := runner.RunCmd(cmd); err != nil {
		klog.Warningf("unable to scale coredns replicas to 1: %v", err)
	} else {
		klog.Infof("successfully scaled coredns replicas to 1")
	}

	// kubectl delete
	client, err := kapi.Client(cc.Name)
	if err != nil {
		return n, err
	}

	err = client.CoreV1().Nodes().Delete(m, nil)
	if err != nil {
		return n, err
	}

	err = machine.DeleteHost(api, m)
	if err != nil {
		return n, err
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

// MustReset reset a stacked control plane to avoid blocking the start of primary control plane
// Exit if failed
func MustReset(cc config.ClusterConfig, n config.Node, api libmachine.API, machineName string) {
	if n.ControlPlane && !n.PrimaryControlPlane {
		host, err := machine.LoadHost(api, machineName)
		if err != nil {
			exit.Error(reason.GuestLoadHost, "Error getting host", err)
		}
		runner, err := machine.CommandRunner(host)
		if err != nil {
			exit.Error(reason.InternalCommandRunner, "Failed to get command runner", err)
		}
		resetCmd := fmt.Sprintf("%s reset -f", bsutil.InvokeKubeadm(cc.KubernetesConfig.KubernetesVersion))
		rc := exec.Command("/bin/bash", "-c", resetCmd)
		_, err = runner.RunCmd(rc)
		if err != nil {
			exit.Error(reason.GuestNodeReset, "Failed to reset kubeadm", err)
		}
	}
}
