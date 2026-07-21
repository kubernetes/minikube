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
	"os/exec"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/util/retry"
)

type linuxProvisioner struct {
	// starter holds the node and cluster configuration for this provisioning workflow.
	starter Starter
	// controlplane is the bootstrapper for the control-plane node, used to generate join tokens.
	controlplane bootstrapper.Bootstrapper
	// worker is the bootstrapper for the target node being joined, used to execute join commands.
	worker bootstrapper.Bootstrapper
}

func (p *linuxProvisioner) Join() error {
	joinCmd, err := p.controlplane.GenerateToken(*p.starter.Cfg)
	if err != nil {
		return fmt.Errorf("error generating join token: %w", err)
	}

	join := func() error {
		klog.Infof("trying to join %s node %q to cluster: %+v", p.starter.role(), p.starter.Node.Name, p.starter.Node)
		if err := p.worker.JoinCluster(*p.starter.Cfg, *p.starter.Node, joinCmd); err != nil {
			klog.Errorf("%s node failed to join cluster, will retry: %v", p.starter.role(), err)

			klog.Infof("resetting %s node %q before attempting to rejoin cluster...", p.starter.role(), p.starter.Node.Name)
			kubeadmBinary := bsutil.KubeadmCmdWithPath(p.starter.Cfg.KubernetesConfig.KubernetesVersion)
			cmd := exec.Command("sudo", "/bin/bash", "-c", fmt.Sprintf("%s reset --force", kubeadmBinary))
			if _, err := p.starter.Runner.RunCmd(cmd); err != nil {
				klog.Infof("kubeadm reset failed, continuing anyway: %v", err)
			} else {
				klog.Infof("successfully reset %s node %q", p.starter.role(), p.starter.Node.Name)
			}

			return err
		}
		return nil
	}
	if err := retry.Expo(join, 10*time.Second, 3*time.Minute); err != nil {
		return fmt.Errorf("error joining %s node %q to cluster: %w", p.starter.role(), p.starter.Node.Name, err)
	}

	return nil
}

func (p *linuxProvisioner) LabelAndUntaint() error {
	if err := p.controlplane.LabelAndUntaintNode(*p.starter.Cfg, *p.starter.Node); err != nil {
		return fmt.Errorf("error applying %s node %q label: %w", p.starter.role(), p.starter.Node.Name, err)
	}
	return nil
}

// Compile-time assertion that linuxProvisioner implements the Provisioner interface.
var _ Provisioner = (*linuxProvisioner)(nil)
