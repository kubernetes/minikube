/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/util/retry"
)

type windowsProvisioner struct {
	// starter holds the node and cluster configuration for this provisioning workflow.
	starter Starter
	// controlplane is the bootstrapper for the control-plane node, used to generate join tokens.
	controlplane bootstrapper.Bootstrapper
	// worker is the bootstrapper for the target node being joined, used to execute join commands.
	worker bootstrapper.Bootstrapper
}

// Compile-time assertion that windowsProvisioner implements the Provisioner interface.
var _ Provisioner = (*windowsProvisioner)(nil)

func (p *windowsProvisioner) Join() error {
	joinConfig, err := p.controlplane.GenerateJoinConfigWindows(*p.starter.Cfg, *p.starter.Node)
	if err != nil {
		return fmt.Errorf("generating Windows join configuration: %w", err)
	}

	// Kubeadm owns the discovery, kubelet-health and TLS-bootstrap deadlines.
	// Never replay join while a previous attempt could still be modifying kubelet.
	klog.Infof("Joining Windows node %q with a single kubeadm join (native phase timeouts)", p.starter.Node.Name)
	output, err := p.worker.JoinClusterWindows(p.starter.Host, joinConfig)
	if err != nil {
		return fmt.Errorf("joining Windows node %q: %w\n%s", p.starter.Node.Name, err, output)
	}
	klog.Infof("Windows kubeadm join completed: %s", output)
	if err := prepareWindowsNodeFlannel(); err != nil {
		return fmt.Errorf("preparing Windows flannel: %w", err)
	}
	if err := prepareWindowsNodeKubeProxy(); err != nil {
		return fmt.Errorf("preparing Windows kube-proxy: %w", err)
	}
	return nil
}

func (p *windowsProvisioner) LabelAndUntaint() error {
	labelFn := func() error {
		return p.controlplane.LabelAndUntaintNode(*p.starter.Cfg, *p.starter.Node)
	}
	// Windows nodes take extra time to register with the API server after kubeadm
	// join returns. Retry labeling until the node object appears (up to 3 minutes).
	if err := retry.Expo(labelFn, 5*time.Second, 3*time.Minute); err != nil {
		return fmt.Errorf("error applying %s node %q label: %w", p.starter.Node.Role(), p.starter.Node.Name, err)
	}
	return nil
}
