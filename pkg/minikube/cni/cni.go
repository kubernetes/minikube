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
	"net"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

const (
	// DefaultPodCIDR is the default CIDR to use in minikube CNI's.
	DefaultPodCIDR = "10.244.0.0/16"

	// DefaultConfDir is the default CNI Config Directory path
	DefaultConfDir = "/etc/cni/net.d"
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

	if len(cc.Nodes) > 1 || cc.MultiNodeRequested {
		// Enables KindNet CNI in master in multi node cluster, This solves the network problem
		// inside pod for multi node clusters. See https://github.com/kubernetes/minikube/issues/9838.
		klog.Infof("multinode detected (%d nodes found), recommending kindnet", len(cc.Nodes))
		return KindNet{cc: cc}
	}

	if cc.KubernetesConfig.ContainerRuntime != constants.Docker {
		// Always use CNI when running with CRI (without dockershim)
		if driver.IsKIC(cc.Driver) {
			klog.Infof("%q driver + %q runtime found, recommending kindnet", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
			return KindNet{cc: cc}
		}
		klog.Infof("%q driver + %q runtime found, recommending bridge", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
		return Bridge{cc: cc}
	}

	// for docker container runtime and k8s v1.24+ where dockershim and kubenet were removed, we fallback to bridge cni for cri-docker(d)
	// ref: https://github.com/Mirantis/cri-dockerd#important
	// ref: https://github.com/Mirantis/cri-dockerd#to-use-with-kubernetes
	// note: currently, default cni that we "distribute" (in /etc/cni/net.d) is based on cri-o bridge, and
	// because it does not currently use portmap plugin, we pick "our" bridge instead (cri-o one will be disabled automatically)
	// ref: https://github.com/cri-o/cri-o/blob/f317b267ddef21aee5ffc92d890a77112b006815/contrib/cni/10-crio-bridge.conflist
	kv, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err == nil && kv.GTE(semver.MustParse("1.24.0-alpha.2")) {
		klog.Infof("%q driver + %q container runtime found on kubernetes v1.24+, recommending bridge", cc.Driver, cc.KubernetesConfig.ContainerRuntime)
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

// ConfigureLoopbackCNI configures loopback cni.
// If disable is true, sets extension of its config file in /etc/cni/net.d to "mk_disabled".
// Otherwise, ensures loopback cni has expected version ("1.0.0") and valid name ("loopback") in its config file in /etc/cni/net.d.
// Note: cri-o is leaving out name atm (https://github.com/cri-o/cri-o/pull/6273).
// Avoid errors like:
// - Failed to create pod sandbox: rpc error: code = Unknown desc = [failed to set up sandbox container "..." network for pod "...": networkPlugin cni failed to set up pod "..." network: missing network name:,
// - failed to clean up sandbox container "..." network for pod "...": networkPlugin cni failed to teardown pod "..." network: missing network name]
// It is caller's responsibility to restart container runtime for these changes to take effect.
func ConfigureLoopbackCNI(r Runner, disable bool) error {
	loopback := "/etc/cni/net.d/*loopback.conf*" // usually: 200-loopback.conf
	// turn { "cniVersion": "0.3.1", "type": "loopback" }
	// into { "cniVersion": "0.3.1", "name": "loopback", "type": "loopback" }
	if _, err := r.RunCmd(exec.Command("sh", "-c", fmt.Sprintf("stat %s", loopback))); err != nil {
		klog.Warningf("loopback cni configuration skipped: %q not found", loopback)
		return nil
	}

	findExec := []string{"find", filepath.Dir(loopback), "-maxdepth", "1", "-type", "f", "-name", filepath.Base(loopback), "-not", "-name", "*.mk_disabled", "-exec", "sh", "-c"}

	if disable {
		if _, err := r.RunCmd(exec.Command(
			"sudo", append(findExec,
				`sudo mv {} {}.mk_disabled`, ";")...)); err != nil {
			return fmt.Errorf("unable to disable loopback cni %q: %v", loopback, err)
		}
		klog.Infof("loopback cni configuration disabled: %q found", loopback)
		return nil
	}

	if _, err := r.RunCmd(exec.Command(
		"sudo", append(findExec,
			`grep -q loopback {} && ( grep -q name {} || sudo sed -i '/"type": "loopback"/i \ \ \ \ "name": "loopback",' {} ) && sudo sed -i 's|"cniVersion": ".*"|"cniVersion": "1.0.0"|g' {}`, ";")...)); err != nil {
		return fmt.Errorf("unable to patch loopback cni config %q: %v", loopback, err)
	}
	klog.Infof("loopback cni configuration patched: %q found", loopback)
	return nil
}

// ConfigureDefaultBridgeCNIs configures all default bridge CNIs on a node (designated by runner).
// If network plugin is set (could be, eg "cni" or "kubenet"), it will disable all default bridges to avoid conflicts.
// Otherwise, it will configure all default bridges to match DefaultPodCIDR subnet range.
// It's usually called before deploying new CNI and on node restarts, to avoid conflicts and flip-flopping of pods' ip addresses.
// It is caller's responsibility to restart container runtime for these changes to take effect.
func ConfigureDefaultBridgeCNIs(r Runner, networkPlugin string) error {
	if networkPlugin != "" {
		return disableAllBridgeCNIs(r)
	}
	return configureAllBridgeCNIs(r, DefaultPodCIDR)
}

// disableAllBridgeCNIs disables all bridge cnis by changing extension to "mk_disabled" of all *bridge* config file(s) found in default location (ie, /etc/cni/net.d).
func disableAllBridgeCNIs(r Runner) error {
	path := "/etc/cni/net.d"

	out, err := r.RunCmd(exec.Command(
		// for cri-o, we also disable 87-podman.conflist (that does not have 'bridge' in its name)
		"sudo", "find", path, "-maxdepth", "1", "-type", "f", "(", "(", "-name", "*bridge*", "-or", "-name", "*podman*", ")", "-and", "-not", "-name", "*.mk_disabled", ")", "-printf", "%p, ", "-exec", "sh", "-c",
		`sudo mv {} {}.mk_disabled`, ";"))
	if err != nil {
		return fmt.Errorf("failed to disable all bridge cni configs in %q: %v", path, err)
	}
	configs := strings.Trim(out.Stdout.String(), ", ")
	if len(configs) == 0 {
		klog.Infof("no active bridge cni configs found in %q - nothing to disable", path)
		return nil
	}
	klog.Infof("disabled [%s] bridge cni config(s)", configs)
	return nil
}

// configureAllBridgeCNIs configures all bridge cnis by changing ip address range to match DefaultPodCIDR in all *bridge* config file(s) found in default location (ie, /etc/cni/net.d).
// ref: https://github.com/containernetworking/cni/blob/main/libcni/conf.go
// ref: https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/troubleshooting-cni-plugin-related-errors/
func configureAllBridgeCNIs(r Runner, cidr string) error {
	// non-podman bridge configs:
	out, err := r.RunCmd(exec.Command(
		"sudo", "find", DefaultConfDir, "-maxdepth", "1", "-type", "f", "-name", "*bridge*", "-not", "-name", "*podman*", "-not", "-name", "*.mk_disabled", "-printf", "%p, ", "-exec", "sh", "-c",
		// remove ipv6 entries to avoid "failed to set bridge addr: could not add IP address to \"cni0\": permission denied"
		// ref: https://github.com/cri-o/cri-o/issues/3555
		// then also remove trailing comma after ipv4 elements, if any
		// ie, this will transform from, eg:
		// from: "ranges": [ [{ "subnet": "10.85.0.0/16" }], [{ "subnet": "1100:200::/24" }] ]
		// to:   "ranges": [ [{ "subnet": "10.244.0.0/16" }] ]
		// getting something similar to https://github.com/cri-o/cri-o/blob/main/contrib/cni/11-crio-ipv4-bridge.conflist
		fmt.Sprintf(`sudo sed -i -r -e '/"dst": ".*:.*"/d' -e 's|^(.*)"dst": (.*)[,*]$|\1"dst": \2|g' -e '/"subnet": ".*:.*"/d' -e 's|^(.*)"subnet": ".*"(.*)[,*]$|\1"subnet": "%s"\2|g' {}`, cidr), ";"))
	if err != nil {
		return fmt.Errorf("failed to configure non-podman bridge cni configs in %q: %v", DefaultConfDir, err)
	}
	configs := out.Stdout.String()

	// podman bridge config(s):
	// could be eg, 87-podman-bridge.conflist or 87-podman.conflist
	// ref: https://github.com/containers/podman/blob/main/cni/87-podman-bridge.conflist
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil || ip.To4() == nil {
		return fmt.Errorf("cidr %q is not valid ipv4 address: %v", cidr, err)
	}
	gateway := ip.Mask(ipnet.Mask)
	gateway[3]++
	out, err = r.RunCmd(exec.Command(
		"sudo", "find", DefaultConfDir, "-maxdepth", "1", "-type", "f", "-name", "*podman*", "-not", "-name", "*.mk_disabled", "-printf", "%p, ", "-exec", "sh", "-c",
		fmt.Sprintf(`sudo sed -i -r -e 's|^(.*)"subnet": ".*"(.*)$|\1"subnet": "%s"\2|g' -e 's|^(.*)"gateway": ".*"(.*)$|\1"gateway": "%s"\2|g' {}`, cidr, gateway), ";"))
	if err != nil {
		return fmt.Errorf("failed to configure podman bridge cni configs in %q: %v", DefaultConfDir, err)
	}
	configs += out.Stdout.String()

	configs = strings.Trim(configs, ", ")
	if len(configs) == 0 {
		klog.Infof("no active bridge cni configs found in %q - nothing to configure", DefaultConfDir)
		return nil
	}
	klog.Infof("configured [%s] bridge cni config(s)", configs)
	return nil
}
