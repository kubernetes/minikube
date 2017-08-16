package kubeadm

import (
	"encoding/json"

	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	clientv1 "k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
	"k8s.io/minikube/pkg/minikube/service"
)

const masterTaint = "node-role.kubernetes.io/master"

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
	client, err := k8s.GetClientset()
	clusterRoleBinding := &rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: "minikube-rbac",
		},
		Subjects: []rbacv1beta1.Subject{
			rbacv1beta1.Subject{
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

	_, err = client.RbacV1beta1().ClusterRoleBindings().Create(clusterRoleBinding)
	if err != nil {
		return errors.Wrap(err, "creating clusterrolebinding")
	}
	return nil
}
