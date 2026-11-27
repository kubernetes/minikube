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
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"errors"

	"github.com/blang/semver/v4"
	"github.com/spf13/viper"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

// Add adds a new node config to an existing cluster.
func Add(cc *config.ClusterConfig, n config.Node, delOnFail bool, options *run.CommandOptions) error {
	profiles, err := config.ListValidProfiles()
	if err != nil {
		return err
	}

	machineName := config.MachineName(*cc, n)
	for _, p := range profiles {
		if p.Config.Name == cc.Name {
			continue
		}

		for _, existNode := range p.Config.Nodes {
			if machineName == config.MachineName(*p.Config, existNode) {
				return fmt.Errorf("Node %s already exists in %s profile", machineName, p.Name)
			}
		}
	}

	if n.ControlPlane && n.Port == 0 {
		n.Port = cc.APIServerPort
	}

	if err := config.SaveNode(cc, &n); err != nil {
		return fmt.Errorf("save node: %w", err)
	}

	r, p, m, h, err := Provision(cc, &n, delOnFail, options)
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

	_, err = Start(s, options)
	return err
}

// teardown drains, then resets and finally deletes node from cluster.
// ref: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/#tear-down
func teardown(cc config.ClusterConfig, name string, options *run.CommandOptions) (*config.Node, error) {
	// get runner for named node - has to be done before node is drained
	n, _, err := Retrieve(cc, name)
	if err != nil {
		return n, fmt.Errorf("retrieve node: %w", err)
	}
	m := config.MachineName(cc, *n)

	api, err := machine.NewAPIClient(options)
	if err != nil {
		return n, fmt.Errorf("get api client: %w", err)
	}

	h, err := machine.LoadHost(api, m)
	if err != nil {
		return n, fmt.Errorf("load host: %w", err)
	}

	r, err := machine.CommandRunner(h)
	if err != nil {
		return n, fmt.Errorf("get command runner: %w", err)
	}

	// get runner for healthy control-plane node
	cpr := mustload.Healthy(cc.Name, options).CP.Runner

	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)

	// kubectl drain node with extra options to prevent ending up stuck in the process
	// ref: https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#drain
	// ref: https://github.com/kubernetes/kubernetes/pull/95076
	cmd := exec.Command("sudo", "KUBECONFIG=/var/lib/minikube/kubeconfig", kubectl, "drain", m,
		"--force", "--grace-period=1", "--skip-wait-for-delete-timeout=1", "--disable-eviction", "--ignore-daemonsets", "--delete-emptydir-data")
	if _, err := cpr.RunCmd(cmd); err != nil {
		klog.Warningf("kubectl drain node %q failed (will continue): %v", m, err)
	} else {
		klog.Infof("successfully drained node %q", m)
	}

	// kubeadm reset node to revert any changes made by previous kubeadm init/join
	// it's to inform cluster of the node that is about to be removed and should be unregistered (eg, from etcd quorum, that would otherwise complain)
	// ref: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-reset/
	// avoid "Found multiple CRI endpoints on the host. Please define which one do you wish to use by setting the 'criSocket' field in the kubeadm configuration file: unix:///var/run/containerd/containerd.sock, unix:///var/run/cri-dockerd.sock" error
	// intentionally non-fatal on any error, propagate and check at the end of segment
	var kerr error
	var kv semver.Version
	kv, kerr = util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if kerr == nil {
		var crt cruntime.Manager
		crt, kerr = cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime, Runner: r, Socket: cc.KubernetesConfig.CRISocket, KubernetesVersion: kv})
		if kerr == nil {
			sp := crt.SocketPath()
			// avoid warning/error:
			// 'Usage of CRI endpoints without URL scheme is deprecated and can cause kubelet errors in the future.
			//  Automatically prepending scheme "unix" to the "criSocket" with value "/var/run/cri-dockerd.sock".
			//  Please update your configuration!'
			if !strings.HasPrefix(sp, "unix://") {
				sp = "unix://" + sp
			}

			cmd := exec.Command("sudo", "/bin/bash", "-c", fmt.Sprintf("KUBECONFIG=/var/lib/minikube/kubeconfig %s reset --force --ignore-preflight-errors=all --cri-socket=%s",
				bsutil.KubeadmCmdWithPath(cc.KubernetesConfig.KubernetesVersion), sp))
			if _, kerr = r.RunCmd(cmd); kerr == nil {
				klog.Infof("successfully reset node %q", m)
			}
		}
	}
	if kerr != nil {
		klog.Warningf("kubeadm reset node %q failed (will continue, but cluster might become unstable): %v", m, kerr)
	}

	// kubectl delete node
	client, err := kapi.Client(cc.Name)
	if err != nil {
		return n, err
	}

	// set 'GracePeriodSeconds: 0' option to delete node immediately (ie, w/o waiting)
	var grace *int64
	// for ha clusters, in case we're deleting control-plane node that's current leader, we retry to allow leader re-election process to complete
	deleteNode := func() error {
		return client.CoreV1().Nodes().Delete(context.Background(), m, v1.DeleteOptions{GracePeriodSeconds: grace})
	}
	err = retry.Expo(deleteNode, kconst.APICallRetryInterval, 2*time.Minute)
	if err != nil {
		klog.Errorf("kubectl delete node %q failed: %v", m, err)
		return n, err
	}
	klog.Infof("successfully deleted node %q", m)

	return n, nil
}

// removeFromProfile drops the node matching name from cc.Nodes and persists
// the profile. It returns the removed Node on success. Factored out of Delete
// so the orphaned-node fast path and the normal teardown path share the same
// profile-update logic, and so it is directly unit-testable without a driver.
func removeFromProfile(cc *config.ClusterConfig, name string, miniHome ...string) (*config.Node, error) {
	n, index, err := Retrieve(*cc, name)
	if err != nil {
		return nil, fmt.Errorf("retrieve: %w", err)
	}
	cc.Nodes = append(cc.Nodes[:index], cc.Nodes[index+1:]...)
	profileName := cc.Name
	if viper.IsSet(config.ProfileName) {
		profileName = viper.GetString(config.ProfileName)
	}
	return n, config.SaveProfile(profileName, cc, miniHome...)
}

// Delete calls teardownNode to remove node from cluster and deletes the host.
// For orphaned nodes (empty IP, usually left behind by a `minikube node add`
// that was interrupted between SaveNode and Provision), Delete short-circuits
// the teardown/apiserver path (which would otherwise try to drain and
// kubeadm-reset a host that was never provisioned) and just evicts the stale
// entry from the profile. See kubernetes/minikube#9834.
func Delete(cc config.ClusterConfig, name string, options *run.CommandOptions) (*config.Node, error) {
	if n, _, err := Retrieve(cc, name); err == nil && n.IP == "" {
		klog.Infof("node %q has empty IP; skipping teardown (never provisioned)", name)
		if api, aerr := machine.NewAPIClient(options); aerr == nil {
			// DeleteHost reports mcnerror.ErrHostDoesNotExist for a missing
			// host, which we swallow; it still runs KIC container cleanup
			// (docker/podman rm) as a side effect for any stale leftover
			// with the same name.
			if derr := machine.DeleteHost(api, config.MachineName(cc, *n)); derr != nil {
				klog.Warningf("delete host for orphaned node %q: %v", name, derr)
			}
		} else {
			klog.Warningf("get api client to clean orphan %q: %v", name, aerr)
		}
		return removeFromProfile(&cc, name)
	}

	n, err := teardown(cc, name, options)
	if err != nil {
		return n, err
	}

	m := config.MachineName(cc, *n)
	api, err := machine.NewAPIClient(options)
	if err != nil {
		return n, err
	}

	if err = machine.DeleteHost(api, m); err != nil {
		return n, err
	}

	return removeFromProfile(&cc, name)
}

// Retrieve finds the node by name in the given cluster
func Retrieve(cc config.ClusterConfig, name string) (*config.Node, int, error) {
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

// PruneOrphanedNodes drops non-primary nodes with an empty IP from existing
// (on-disk profile) and keeps inflight (the in-flight ClusterConfig held by
// start) in sync, so a later node.Add → config.SaveNode does not re-persist
// a pruned entry via SaveProfile(inflight). An interrupted
// `minikube node add` leaves the profile with a Node whose IP is "" because
// SaveNode runs before Provision; the next `minikube start` then fails to
// bring the orphan up. This is the recovery half of kubernetes/minikube#9834.
//
// The primary control-plane node is never pruned here: startMachine owns its
// lifecycle and may still be re-provisioning it. Returns the names of nodes
// that were dropped (iteration order) so the caller can surface them to the
// user. The profile is persisted via config.SaveProfile only when at least
// one node was dropped.
func PruneOrphanedNodes(existing, inflight *config.ClusterConfig, profileName string, miniHome ...string) ([]string, error) {
	if existing == nil {
		return nil, nil
	}
	var dropped []string
	keptExisting := existing.Nodes[:0]
	for _, n := range existing.Nodes {
		if n.IP == "" && !config.IsPrimaryControlPlane(*existing, n) {
			dropped = append(dropped, n.Name)
			continue
		}
		keptExisting = append(keptExisting, n)
	}
	if len(dropped) == 0 {
		return nil, nil
	}
	existing.Nodes = keptExisting
	if inflight != nil {
		keptInflight := inflight.Nodes[:0]
		for _, n := range inflight.Nodes {
			if n.IP == "" && !config.IsPrimaryControlPlane(*inflight, n) {
				continue
			}
			keptInflight = append(keptInflight, n)
		}
		inflight.Nodes = keptInflight
	}
	return dropped, config.SaveProfile(profileName, existing, miniHome...)
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

// Name returns the appropriate name for the node given the node index.
func Name(index int) string {
	if index == 0 {
		return ""
	}
	return fmt.Sprintf("m%02d", index)
}

// ID returns the appropriate node id from the node name.
// ID of first (primary control-plane) node (with empty name) is 1, so next one would be "m02", etc.
// Eg, "m05" should return "5", regardless if any preceded nodes were deleted.
func ID(name string) (int, error) {
	if name == "" {
		return 1, nil
	}

	name = strings.TrimPrefix(name, "m")
	i, err := strconv.Atoi(name)
	if err != nil {
		return -1, err
	}
	return i, nil
}
