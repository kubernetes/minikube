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

package kubeadm

import (
	"encoding/json"
	"net"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/util"
)

const (
	masterTaint = "node-role.kubernetes.io/master"
	rbacName    = "minikube-rbac"
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
			return &util.RetriableError{Err: errors.Wrap(err, "creating clusterrolebinding")}
		}
		return errors.Wrap(err, "creating clusterrolebinding")
	}
	return nil
}
