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
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	// DefaultPodCIDR is the default CIDR to use in minikube CNI's.
	DefaultPodCIDR = "10.244.0.0/16"
)

// Runner is the subset of command.Runner this package consumes
type Runner interface {
	RunCmd(cmd *exec.Cmd) (*command.RunResult, error)
	Copy(assets.CopyableFile) error
}

// Manager is a common interface for CNI
type Manager interface {
	// Apply a CNI. The provided runner is for the control plane
	Apply(Runner) error

	// CIDR returns the default CIDR used by this CNI
	CIDR() string

	// String representation
	String() string
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
		klog.Infof("network plugin configured as %q, returning disabled", cc.KubernetesConfig.NetworkPlugin)
		return Disabled{}, nil
	}

	klog.Infof("Creating CNI manager for %q", cc.KubernetesConfig.CNI)

	switch cc.KubernetesConfig.CNI {
	case "", "auto":
		return chooseDefault(cc), nil
	case "false":
		return Disabled{cc: cc}, nil
	case "kindnet", "true":
		return KindNet{cc: cc}, nil
	case "bridge":
		return Bridge{cc: cc}, nil
	case "calico":
		return Calico{cc: cc}, nil
	case "cilium":
		return Cilium{cc: cc}, nil
	case "flannel":
		return Flannel{cc: cc}, nil
	default:
		return NewCustom(cc, cc.KubernetesConfig.CNI)
	}
}

func IsDisabled(cc config.ClusterConfig) bool {
	if cc.KubernetesConfig.NetworkPlugin != "" && cc.KubernetesConfig.NetworkPlugin != "cni" {
		return true
	}

	if cc.KubernetesConfig.CNI == "false" {
		return true
	}

	if chooseDefault(cc).String() == "Disabled" {
		return true
	}
	return false
}

func chooseDefault(cc config.ClusterConfig) Manager {
	// For backwards compatibility with older profiles using --enable-default-cni
	if cc.KubernetesConfig.EnableDefaultCNI {
		klog.Infof("EnableDefaultCNI is true, recommending bridge")
		return Bridge{}
	}

	if cc.KubernetesConfig.ContainerRuntime != "docker" {
		if driver.IsKIC(cc.Driver) {
			klog.Infof("%q driver + %s runtime found, recommending kindnet", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
			return KindNet{cc: cc}
		}
		klog.Infof("%q driver + %s runtime found, recommending bridge", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
		return Bridge{cc: cc}
	}

	if driver.BareMetal(cc.Driver) {
		klog.Infof("Driver %s used, CNI unnecessary in this configuration, recommending no CNI", cc.Driver)
		return Disabled{cc: cc}
	}

	if len(cc.Nodes) > 1 || cc.MultiNodeRequested {
		// Enables KindNet CNI in master in multi node cluster, This solves the network problem
		// inside pod for multi node clusters. See https://github.com/kubernetes/minikube/issues/9838.
		klog.Infof("%d nodes found, recommending kindnet", len(cc.Nodes))
		return KindNet{cc: cc}
	}

	klog.Infof("CNI unnecessary in this configuration, recommending no CNI")
	return Disabled{cc: cc}
}

// manifestPath returns the path to the CNI manifest
func manifestPath() string {
	return path.Join(vmpath.GuestEphemeralDir, "cni.yaml")
}

// manifestAsset returns a copyable asset for the CNI manifest
func manifestAsset(b []byte) assets.CopyableFile {
	return assets.NewMemoryAssetTarget(b, manifestPath(), "0644")
}

// applyManifest applies a CNI manifest
func applyManifest(cc config.ClusterConfig, r Runner, f assets.CopyableFile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	klog.Infof("applying CNI manifest using %s ...", kubectl)

	if err := r.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}

	cmd := exec.CommandContext(ctx, "sudo", kubectl, "apply", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-f", manifestPath())
	if rr, err := r.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}

	return nil
}
