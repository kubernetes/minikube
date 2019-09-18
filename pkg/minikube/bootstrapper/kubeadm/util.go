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
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/util/retry"
)

const (
	rbacName = "minikube-rbac"
)

// elevateKubeSystemPrivileges gives the kube-system service account
// cluster admin privileges to work with RBAC.
func elevateKubeSystemPrivileges(client kubernetes.Interface) error {
	start := time.Now()
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
	if _, err := client.RbacV1beta1().ClusterRoleBindings().Create(clusterRoleBinding); err != nil {
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			return &retry.RetriableError{Err: errors.Wrap(err, "creating clusterrolebinding")}
		}
		return errors.Wrap(err, "creating clusterrolebinding")
	}
	glog.Infof("duration metric: took %s to wait for elevateKubeSystemPrivileges.", time.Since(start))
	return nil
}
