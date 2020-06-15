/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Package cni configures the Container Networking Interface
package cni

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	// defaultPodCIDR is the default CIDR to use in minikube CNI's.
	defaultPodCIDR = "10.244.0.0/16"
)

// Runner is the subset of command.Runner this package consumes
type Runner interface {
	RunCmd(cmd *exec.Cmd) (*command.RunResult, error)
}

// Manager is a common interface for CNI
type Manager interface {
	// Assets returns a list of assets necessary to enable this CNI
	Assets() ([]assets.CopyableFile, error)

	// Enable enables the CNI
	Apply(context.Context, Runner) error

	// CIDR returns the default CIDR used by this CNI
	CIDR() string
}

// New returns a new CNI manager
func New(cc config.ClusterConfig) Manager {
	if cc.KubernetesConfig.NetworkPlugin != "" && cc.KubernetesConfig.NetworkPlugin != "cni" {
		glog.Infof("network plugin configured as %q, returning noop", cc.KubernetesConfig.NetworkPlugin)
		return Noop{}
	}

	glog.Infof("Creating CNI manager for %q", cc.KubernetesConfig.CNI)

	switch cc.KubernetesConfig.CNI {
	case "", "true":
		return chooseDefault(cc)
	case "false":
		return Noop{}
	case "kindnet":
		return KindNet{cc: cc}
	case "custom":
		return Custom{}
	case "calico":
		return Calico{cc: cc}
	case "flannel":
		return Flannel{cc: cc}
	default:
		glog.Errorf("unknown CNI choice: %q", cc.KubernetesConfig.CNI)
	}
	return nil
}

func chooseDefault(cc config.ClusterConfig) Manager {
	if cc.Driver == "Docker" {
		glog.Infof("%q driver found, recommending kindnet", cc.Driver)
		return KindNet{cc: cc}
	}
	if len(cc.Nodes) > 1 {
		glog.Infof("%d nodes found, recommending kindnet", len(cc.Nodes))
		return KindNet{cc: cc}
	}
	return Noop{}
}

// manifestPath returns the path to the CNI manifest
func manifestPath() string {
	return path.Join(vmpath.GuestEphemeralDir, "cni.yaml")
}

// manifestAsset returns a copyable asset for the CNI manifest
func manifestAsset(b []byte) assets.CopyableFile {
	return assets.NewMemoryAssetTarget(b, manifestPath(), "0644")
}

// kubectlPath returns the path to the kubelet
func kubectlPath(cc config.ClusterConfig) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", cc.KubernetesConfig.KubernetesVersion, "kubectl")
}

// apply applies a CNI manifest
func apply(ctx context.Context, r Runner, cc config.ClusterConfig) error {
	if cc.KubernetesConfig.KubernetesVersion == "" {
		return fmt.Errorf("KubernetesVersion is unknown, empty configuration?")
	}

	kubectl := kubectlPath(cc)
	glog.Infof("applying CNI manifest using %s ...", kubectl)

	cmd := exec.CommandContext(ctx, "sudo", kubectl, "apply",
		fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")),
		"-f", manifestPath())

	if rr, err := r.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}

	return nil
}
