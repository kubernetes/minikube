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
	joinCmd, err := p.controlplane.GenerateTokenWindows(*p.starter.Cfg, *p.starter.Node)
	if err != nil {
		return fmt.Errorf("error generating join token: %w", err)
	}

	join := func() error {
		klog.Infof("trying to join %s node %q to cluster: %+v", p.starter.Node.Role(), p.starter.Node.Name, p.starter.Node)

		driverIP, err := p.starter.Host.Driver.GetIP()
		if err != nil {
			klog.Errorf("Unable to get driver IP: %v", err)
		}
		klog.Infof("Driver IP: %s", driverIP)

		// Allow enough time for kubeadm join to complete on a Windows node.
		// The initial 20s was too short; Windows kubeadm join takes longer.
		timeout := 1 * time.Minute

		if commandResult, err := p.worker.JoinClusterWindows(p.starter.Host, *p.starter.Cfg, *p.starter.Node, joinCmd, timeout); err != nil {
			klog.Infof("%s node failed to join cluster, will retry: %v", p.starter.Node.Role(), err)
			klog.Infof("command result: %s", commandResult)

			// sort out the certificates issues
			cmd, certErr := p.worker.SetupMinikubeCert(p.starter.Host)
			if certErr != nil {
				return fmt.Errorf("error setting minikube cert: %w", certErr)
			}
			klog.Infof("command result: %s", cmd)
			// retry the join command
			if commandResult, err := p.worker.JoinClusterWindows(p.starter.Host, *p.starter.Cfg, *p.starter.Node, joinCmd, 0); err != nil {
				klog.Errorf("error retrying join command: %v, command result: %s", err, commandResult)
				return err
			}
		}

		// Apply Windows networking config after a successful join regardless
		// of which join path was taken (first attempt or cert-fix retry).
		if err := prepareWindowsNodeFlannel(); err != nil {
			klog.Errorf("error preparing windows node flannel: %v", err)
		}

		if err := prepareWindowsNodeKubeProxy(); err != nil {
			klog.Errorf("error preparing windows node kube-proxy: %v", err)
		}

		return nil
	}
	if err := retry.Expo(join, 10*time.Second, 3*time.Minute); err != nil {
		return fmt.Errorf("error joining %s node %q to cluster: %w", p.starter.Node.Role(), p.starter.Node.Name, err)
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
