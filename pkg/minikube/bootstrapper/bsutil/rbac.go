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

package bsutil

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	rbac "k8s.io/api/rbac/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/util/retry"
)

const (
	rbacName = "minikube-rbac"
)

// ElevateKubeSystemPrivileges gives the kube-system service account
// cluster admin privileges to work with RBAC.
func ElevateKubeSystemPrivileges(client kubernetes.Interface) error {
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

// AdjustResourceLimits makes fine adjustments to pod resources that aren't possible via kubeadm config.
func AdjustResourceLimits(c command.Runner) error {
	rr, err := c.RunCmd(exec.Command("/bin/bash", "-c", "cat /proc/$(pgrep kube-apiserver)/oom_adj"))
	if err != nil {
		return errors.Wrapf(err, "oom_adj check cmd %s. ", rr.Command())
	}
	glog.Infof("apiserver oom_adj: %s", rr.Stdout.String())
	// oom_adj is already a negative number
	if strings.HasPrefix(rr.Stdout.String(), "-") {
		return nil
	}
	glog.Infof("adjusting apiserver oom_adj to -10")

	// Prevent the apiserver from OOM'ing before other pods, as it is our gateway into the cluster.
	// It'd be preferable to do this via Kubernetes, but kubeadm doesn't have a way to set pod QoS.
	if _, err = c.RunCmd(exec.Command("/bin/bash", "-c", "echo -10 | sudo tee /proc/$(pgrep kube-apiserver)/oom_adj")); err != nil {
		return errors.Wrap(err, fmt.Sprintf("oom_adj adjust"))
	}
	return nil
}

// ExistingConfig checks if there are config files from possible previous kubernets cluster
func ExistingConfig(c command.Runner) error {
	args := append([]string{"ls"}, expectedRemoteArtifacts...)
	_, err := c.RunCmd(exec.Command("sudo", args...))
	return err
}
