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
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	// DefaultPodCIDR is the default CIDR to use in minikube CNI's.
	DefaultPodCIDR = "10.244.0.0/16"

	// DefaultConfDir is the default CNI Config Directory path
	DefaultConfDir = "/etc/cni/net.d"
	// CustomConfDir is the custom CNI Config Directory path used to avoid conflicting CNI configs
	// ref: https://github.com/kubernetes/minikube/issues/10984 and https://github.com/kubernetes/minikube/pull/11106
	CustomConfDir = "/etc/cni/net.mk"
)

var (
	// ConfDir is the CNI Config Directory path that can be customised, defaulting to DefaultConfDir
	ConfDir = DefaultConfDir

	// Network is the network name that CNI should use (eg, "kindnet").
	// Currently, only crio (and podman) can use it, so that setting custom ConfDir is not necessary.
	// ref: https://github.com/cri-o/cri-o/issues/2121 (and https://github.com/containers/podman/issues/2370)
	// ref: https://github.com/cri-o/cri-o/blob/master/docs/crio.conf.5.md#crionetwork-table
	Network = ""
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
	CNIConfDir   string
}

// New returns a new CNI manager
func New(cc *config.ClusterConfig) (Manager, error) {
	if cc.KubernetesConfig.NetworkPlugin != "" && cc.KubernetesConfig.NetworkPlugin != "cni" {
		klog.Infof("network plugin configured as %q, returning disabled", cc.KubernetesConfig.NetworkPlugin)
		return Disabled{}, nil
	}

	klog.Infof("Creating CNI manager for %q", cc.KubernetesConfig.CNI)

	var cnm Manager
	var err error
	switch cc.KubernetesConfig.CNI {
	case "", "auto":
		cnm = chooseDefault(*cc)
	case "false":
		cnm = Disabled{cc: *cc}
	case "kindnet", "true":
		cnm = KindNet{cc: *cc}
	case "bridge":
		cnm = Bridge{cc: *cc}
	case "calico":
		cnm = Calico{cc: *cc}
	case "cilium":
		cnm = Cilium{cc: *cc}
	case "flannel":
		cnm = Flannel{cc: *cc}
	default:
		cnm, err = NewCustom(*cc, cc.KubernetesConfig.CNI)
	}

	if err := configureCNI(cc, cnm); err != nil {
		klog.Errorf("unable to set CNI Config Directory: %v", err)
	}

	return cnm, err
}

// IsDisabled checks if CNI is disabled
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

	if cc.KubernetesConfig.ContainerRuntime != "docker" {
		if driver.IsKIC(cc.Driver) {
			klog.Infof("%q driver + %s runtime found, recommending kindnet", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
			return KindNet{cc: cc}
		}
		klog.Infof("%q driver + %s runtime found, recommending bridge", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
		return Bridge{cc: cc}
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

// configureCNI - to avoid conflicting CNI configs, it sets:
// - for crio: 'cni_default_network' config param via cni.Network
// - for containerd and docker: kubelet's '--cni-conf-dir' flag to custom CNI Config Directory path (same used also by CNI Deployment).
// ref: https://github.com/kubernetes/minikube/issues/10984 and https://github.com/kubernetes/minikube/pull/11106
// Note: currently, this change affects only Kindnet CNI (and all multinodes using it), but it can be easily expanded to other/all CNIs if needed.
// Note2: Cilium does not need workaround as they automatically restart pods after CNI is successfully deployed.
func configureCNI(cc *config.ClusterConfig, cnm Manager) error {
	if _, kindnet := cnm.(KindNet); kindnet {
		// crio only needs CNI network name; hopefully others (containerd, docker and kubeadm/kubelet) will follow eventually
		if cc.KubernetesConfig.ContainerRuntime == constants.CRIO {
			Network = "kindnet"
			return nil
		}
		// for containerd and docker: auto-set custom CNI via kubelet's 'cni-conf-dir' param, if not user-specified
		eo := fmt.Sprintf("kubelet.cni-conf-dir=%s", CustomConfDir)
		if !cc.KubernetesConfig.ExtraOptions.Exists(eo) {
			klog.Infof("auto-setting extra-config to %q", eo)
			if err := cc.KubernetesConfig.ExtraOptions.Set(eo); err != nil {
				return fmt.Errorf("failed auto-setting extra-config %q: %v", eo, err)
			}
			ConfDir = CustomConfDir
			klog.Infof("extra-config set to %q", eo)
		} else {
			// respect user-specified custom CNI Config Directory
			ConfDir = cc.KubernetesConfig.ExtraOptions.Get("cni-conf-dir", "kubelet")
		}
	}
	return nil
}
