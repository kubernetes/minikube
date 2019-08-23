/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package kic

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"encoding/json"
	"net"
	"time"

	"github.com/blang/semver"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/util/retry"
)

const (
	masterTaint = "node-role.kubernetes.io/master"
	rbacName    = "minikube-rbac"
	// These are the components that can be configured
	// through the "extra-config"
	Kubelet           = "kubelet"
	Kubeadm           = "kubeadm"
	Apiserver         = "apiserver"
	Scheduler         = "scheduler"
	ControllerManager = "controller-manager"
)

var master = ""

func unmarkMaster() error {
	k8s := service.K8s
	client, err := k8s.GetCoreClient()
	if err != nil {
		return errors.Wrap(err, "getting core client")
	}
	n, err := client.Nodes().Get(master, meta.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "getting node %s", master)
	}

	oldData, err := json.Marshal(n)
	if err != nil {
		return errors.Wrap(err, "json marshalling data before patch")
	}

	newTaints := []core.Taint{}
	for _, taint := range n.Spec.Taints {
		if taint.Key == masterTaint {
			continue
		}

		newTaints = append(newTaints, taint)
	}
	n.Spec.Taints = newTaints

	newData, err := json.Marshal(n)
	if err != nil {
		return errors.Wrapf(err, "json marshalling data after patch")
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, core.Node{})
	if err != nil {
		return errors.Wrap(err, "creating strategic patch")
	}

	if _, err := client.Nodes().Patch(n.Name, types.StrategicMergePatchType, patchBytes); err != nil {
		if apierr.IsConflict(err) {
			return errors.Wrap(err, "strategic patch conflict")
		}
		return errors.Wrap(err, "applying strategic patch")
	}

	return nil
}

// elevateKubeSystemPrivileges gives the kube-system service account
// cluster admin privileges to work with RBAC.
func elevateKubeSystemPrivileges() error {
	start := time.Now()
	k8s := service.K8s
	client, err := k8s.GetClientset(constants.DefaultK8sClientTimeout)
	if err != nil {
		return errors.Wrap(err, "getting clientset")
	}
	clusterRoleBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: meta.ObjectMeta{
			Name: rbacName,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: "kube-system",
			},
		},
		RoleRef: rbac.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}

	if _, err := client.RbacV1beta1().ClusterRoleBindings().Get(rbacName, meta.GetOptions{}); err == nil {
		glog.Infof("Role binding %s already exists. Skipping creation.", rbacName)
		return nil
	}
	_, err = client.RbacV1beta1().ClusterRoleBindings().Create(clusterRoleBinding)
	if err != nil {
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			return &retry.RetriableError{Err: errors.Wrap(err, "creating clusterrolebinding")}
		}
		return errors.Wrap(err, "creating clusterrolebinding")
	}
	glog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges.", time.Since(start))
	return nil
}

// for now will duplicate and copy the helpers from kubeadm till we merge these two into one bootstrapper

// enum to differentiate kubeadm command line parameters from kubeadm config file parameters (see the
// KubeadmExtraArgsWhitelist variable below for more info)
const (
	KubeadmCmdParam    = iota
	KubeadmConfigParam = iota
)

// KubeadmExtraArgsWhitelist is a whitelist of supported kubeadm params that can be supplied to kubeadm through
// minikube's ExtraArgs parameter. The list is split into two parts - params that can be supplied as flags on the
// command line and params that have to be inserted into the kubeadm config file. This is because of a kubeadm
// constraint which allows only certain params to be provided from the command line when the --config parameter
// is specified
var KubeadmExtraArgsWhitelist = map[int][]string{
	KubeadmCmdParam: {
		"ignore-preflight-errors",
		"dry-run",
		"kubeconfig",
		"kubeconfig-dir",
		"node-name",
		"cri-socket",
		"experimental-upload-certs",
		"certificate-key",
		"rootfs",
	},
	KubeadmConfigParam: {
		"pod-network-cidr",
	},
}

// type pod struct {
// 	// Human friendly name
// 	name  string
// 	key   string
// 	value string
// }

// // PodsByLayer are queries we run when health checking, sorted roughly by dependency layer
// var PodsByLayer = []pod{
// 	{"proxy", "k8s-app", "kube-proxy"},
// 	{"etcd", "component", "etcd"},
// 	{"scheduler", "component", "kube-scheduler"},
// 	{"controller", "component", "kube-controller-manager"},
// 	{"dns", "k8s-app", "kube-dns"},
// }

// yamlConfigPath is the path to the kubeadm configuration
var yamlConfigPath = filepath.Join(constants.GuestEphemeralDir, "kubeadm.yaml")

// SkipAdditionalPreflights are additional preflights we skip depending on the runtime in use.
var SkipAdditionalPreflights = map[string][]string{}

// parseKubernetesVersion parses the kubernetes version
func parseKubernetesVersion(version string) (semver.Version, error) {
	// Strip leading 'v' prefix from version for semver parsing
	v, err := semver.Make(version[1:])
	if err != nil {
		return semver.Version{}, errors.Wrap(err, "invalid version, must begin with 'v'")
	}

	return v, nil
}

// createFlagsFromExtraArgs converts kubeadm extra args into flags to be supplied from the commad linne
func createFlagsFromExtraArgs(extraOptions config.ExtraOptionSlice) string {
	kubeadmExtraOpts := extraOptions.AsMap().Get(Kubeadm)

	// kubeadm allows only a small set of parameters to be supplied from the command line when the --config param
	// is specified, here we remove those that are not allowed
	for opt := range kubeadmExtraOpts {
		if !config.ContainsParam(KubeadmExtraArgsWhitelist[KubeadmCmdParam], opt) {
			// kubeadmExtraOpts is a copy so safe to delete
			delete(kubeadmExtraOpts, opt)
		}
	}
	return convertToFlags(kubeadmExtraOpts)
}

// invokeKubeadm returns the invocation command for Kubeadm
func invokeKubeadm() string {
	return fmt.Sprintf("env PATH=%s:$PATH kubeadm", "/usr/bin/")
}

func convertToFlags(opts map[string]string) string {
	var flags []string
	var keys []string
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flags = append(flags, fmt.Sprintf("--%s=%s", k, opts[k]))
	}
	return strings.Join(flags, " ")
}
