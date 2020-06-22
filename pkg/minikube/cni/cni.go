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
	"k8s.io/minikube/pkg/minikube/driver"
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

	// NeedsApply returns whether or not CNI requires a manifest to be applied
	NeedsApply() bool

	// Enable enables the CNI
	Apply(context.Context, Runner) error

	// CIDR returns the default CIDR used by this CNI
	CIDR() string
}

// tmplInputs are inputs to CNI templates
type tmplInput struct {
	ImageName    string
	PodCIDR      string
	DefaultRoute string
}

// New returns a new CNI manager
func New(cc config.ClusterConfig) (Manager, error) {
	if cc.KubernetesConfig.NetworkPlugin != "" && cc.KubernetesConfig.NetworkPlugin != "cni" {
		glog.Infof("network plugin configured as %q, returning disabled", cc.KubernetesConfig.NetworkPlugin)
		return Disabled{}, nil
	}

	glog.Infof("Creating CNI manager for %q", cc.KubernetesConfig.CNI)

	switch cc.KubernetesConfig.CNI {
	case "", "true", "auto":
		return chooseDefault(cc), nil
	case "false":
		return Disabled{cc: cc}, nil
	case "kindnet":
		return KindNet{cc: cc}, nil
	case "bridge":
		return Bridge{cc: cc}, nil
	case "calico":
		return Calico{cc: cc}, nil
	case "flannel":
		return Flannel{cc: cc}, nil
	default:
		return NewCustom(cc, cc.KubernetesConfig.CNI)
	}
}

func chooseDefault(cc config.ClusterConfig) Manager {
	// For backwards compatibility with older profiles using --enable-default-cni
	if cc.KubernetesConfig.EnableDefaultCNI {
		glog.Infof("EnableDefaultCNI is true, recommending bridge")
		return Bridge{}
	}

	if driver.IsKIC(cc.Driver) {
		glog.Infof("%q driver found, recommending kindnet", cc.Driver)
		return KindNet{cc: cc}
	}

	if len(cc.Nodes) > 1 {
		glog.Infof("%d nodes found, recommending kindnet", len(cc.Nodes))
		return KindNet{cc: cc}
	}

	glog.Infof("CNI unnecessary in this configuration, recommending no CNI")
	return Disabled{}
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
	kubectl := kubectlPath(cc)
	glog.Infof("applying CNI manifest using %s ...", kubectl)

	cmd := exec.CommandContext(ctx, "sudo", kubectl, "apply", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-f", manifestPath())
	if rr, err := r.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}

	return nil
}
