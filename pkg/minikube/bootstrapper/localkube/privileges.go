package localkube

import (
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	clientv1 "k8s.io/client-go/pkg/api/v1"
	rbacv1beta1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"
	"k8s.io/minikube/pkg/minikube/service"
)

func elevateKubeSystemPrivileges() error {
	k8s := service.K8sClientGetter{}
	client, err := k8s.GetRBACV1Beta1Client()
	if err != nil {
		return err
	}
	clusterRoleBinding := &rbacv1beta1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: "minikube-rbac",
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

	if _, err := client.ClusterRoleBindings().Create(clusterRoleBinding); err != nil {
		return errors.Wrap(err, "creating clusterrolebinding")
	}
	return nil
}
