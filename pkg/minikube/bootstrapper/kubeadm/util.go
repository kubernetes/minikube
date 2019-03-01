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
	"bytes"
	"encoding/json"
	"html/template"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	clientv1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/minikube/pkg/minikube/config"
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
	n, err := client.Nodes().Get(master, v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "getting node %s", master)
	}

	oldData, err := json.Marshal(n)
	if err != nil {
		return errors.Wrap(err, "json marshalling data before patch")
	}

	newTaints := []clientv1.Taint{}
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

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, clientv1.Node{})
	if err != nil {
		return errors.Wrap(err, "creating strategic patch")
	}

	if _, err := client.Nodes().Patch(n.Name, types.StrategicMergePatchType, patchBytes); err != nil {
		if apierrs.IsConflict(err) {
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
	clusterRoleBinding := &rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: rbacName,
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: "kube-system",
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	}

	if _, err := client.RbacV1beta1().ClusterRoleBindings().Get(rbacName, metav1.GetOptions{}); err == nil {
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

const (
	kubeconfigConf         = "kubeconfig.conf"
	kubeProxyConfigmapTmpl = `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    server: https://{{.AdvertiseAddress}}:{{.APIServerPort}}
  name: default
contexts:
- context:
    cluster: default
    namespace: default
    user: default
  name: default
current-context: default
users:
- name: default
  user:
    tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
`
)

// updateKubeProxyConfigMap updates the IP & port kube-proxy listens on, and restarts it.
func updateKubeProxyConfigMap(k8s config.KubernetesConfig) error {
	client, err := util.GetClient()
	if err != nil {
		return errors.Wrap(err, "getting k8s client")
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"k8s-app": "kube-proxy"}))
	if err := util.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "kube-proxy not running")
	}

	cfgMap, err := client.CoreV1().ConfigMaps("kube-system").Get("kube-proxy", metav1.GetOptions{})
	if err != nil {
		return &util.RetriableError{Err: errors.Wrap(err, "getting kube-proxy configmap")}
	}
	glog.Infof("kube-proxy config: %v", cfgMap.Data[kubeconfigConf])
	t := template.Must(template.New("kubeProxyTmpl").Parse(kubeProxyConfigmapTmpl))
	opts := struct {
		AdvertiseAddress string
		APIServerPort    int
	}{
		AdvertiseAddress: k8s.NodeIP,
		APIServerPort:    k8s.NodePort,
	}

	kubeconfig := bytes.Buffer{}
	if err := t.Execute(&kubeconfig, opts); err != nil {
		return errors.Wrap(err, "executing kube proxy configmap template")
	}

	if cfgMap.Data == nil {
		cfgMap.Data = map[string]string{}
	}

	updated := strings.TrimSuffix(kubeconfig.String(), "\n")
	glog.Infof("updated kube-proxy config: %s", updated)

	// An optimization, but also one that's unlikely, as kubeadm writes the address as 'localhost'
	if cfgMap.Data[kubeconfigConf] == updated {
		glog.Infof("kube-proxy config appears to require no change, not restarting kube-proxy")
		return nil
	}
	cfgMap.Data[kubeconfigConf] = updated

	// Make this step retriable, as it can fail with:
	// "Operation cannot be fulfilled on configmaps "kube-proxy": the object has been modified; please apply your changes to the latest version and try again"
	if _, err := client.CoreV1().ConfigMaps("kube-system").Update(cfgMap); err != nil {
		return &util.RetriableError{Err: errors.Wrap(err, "updating configmap")}
	}

	pods, err := client.CoreV1().Pods("kube-system").List(metav1.ListOptions{
		LabelSelector: "k8s-app=kube-proxy",
	})
	if err != nil {
		return errors.Wrap(err, "listing kube-proxy pods")
	}
	for _, pod := range pods.Items {
		// Retriable, as known to fail with: pods "<name>" not found
		if err := client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
			return &util.RetriableError{Err: errors.Wrapf(err, "deleting pod %+v", pod)}
		}
	}

	// Wait for the scheduler to restart kube-proxy
	if err := util.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		return errors.Wrap(err, "kube-proxy not running")
	}

	return nil
}
