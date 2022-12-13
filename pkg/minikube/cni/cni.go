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
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const (
	// DefaultPodCIDR is the default CIDR to use in minikube CNI's.
	DefaultPodCIDR = "10.244.0.0/16"

	// DefaultConfDir is the default CNI Config Directory path
	DefaultConfDir = "/etc/cni/net.d"
)

var (
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

	// Ready returns if CNI is ready (eg, all required pods have Ready PodCondition).
	Ready() bool
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

	if cc.KubernetesConfig.ContainerRuntime != constants.Docker {
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
	if err := NameLoopback(r); err != nil {
		klog.Warningf("unable to name loopback interface in applyManifest: %v", err)
	}

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

// NameLoopback ensures loopback has a name in its config file in /etc/cni/net.d
// cri-o is leaving it out atm (https://github.com/cri-o/cri-o/pull/6273)
// avoid errors like:
// Failed to create pod sandbox: rpc error: code = Unknown desc = [failed to set up sandbox container "..." network for pod "...": networkPlugin cni failed to set up pod "..." network: missing network name:,
// failed to clean up sandbox container "..." network for pod "...": networkPlugin cni failed to teardown pod "..." network: missing network name]
func NameLoopback(r Runner) error {
	loopback := "/etc/cni/net.d/*loopback.conf*" // usually: 200-loopback.conf
	// turn { "cniVersion": "0.3.1", "type": "loopback" }
	// into { "cniVersion": "0.3.1", "name": "loopback", "type": "loopback" }
	if _, err := r.RunCmd(exec.Command("sh", "-c", fmt.Sprintf("stat %s", loopback))); err != nil {
		klog.Warningf("%q not found, skipping patching loopback config step", loopback)
		return nil
	}
	if _, err := r.RunCmd(exec.Command(
		"sudo", "find", filepath.Dir(loopback), "-maxdepth", "1", "-type", "f", "-name", filepath.Base(loopback), "-exec", "sh", "-c",
		`grep -q loopback {} && ( grep -q name {} || sudo sed -i '/"type": "loopback"/i \ \ \ \ "name": "loopback",' {} )`, ";")); err != nil {
		return fmt.Errorf("unable to patch loopback config %q: %v", loopback, err)
	}
	return nil
}

// DisableBridgeCNIs disables all default bridge CNIs on a node (designated by runner) by changing extension to "mk_disabled" of *bridge* config file(s) found in /etc/cni/net.d.
// It's usually called before deploying new CNI or on restarts, to avoid conflicts and flip-flopping of pods' ip addresses.
// ref: https://github.com/containernetworking/cni/blob/main/libcni/conf.go
func DisableAllBridgeCNIs(r Runner, cc config.ClusterConfig) error {
	path := "/etc/cni/net.d"
	out, err := r.RunCmd(exec.Command(
		"sudo", "find", path, "-maxdepth", "1", "-type", "f", "-name", "*bridge*", "-not", "-name", "*.mk_disabled", "-printf", "%p|", "-exec", "sh", "-c",
		`sudo mv {} {}.mk_disabled`, ";"))
	if err != nil {
		return fmt.Errorf("failed to disable all bridge cni configs in %q: %v", path, err)
	}
	configs := strings.Trim(out.Stdout.String(), "|")
	if len(configs) == 0 {
		klog.Infof("no bridge cni config found in %q - nothing to disable", configs, path)
		return nil
	}
	svc := cc.KubernetesConfig.ContainerRuntime
	klog.Infof("disabled [%s] bridge cni config(s) in %q, now restarting selected %q container runtime", configs, path, svc)

	if svc == "cri-o" {
		svc = "crio"
	}
	if err := sysinit.New(r).Restart(svc); err != nil {
		klog.Warningf("failed to restart %q container runtime service in %q: %v", svc, cc.Name, err)
		return err
	}

	return nil
}
