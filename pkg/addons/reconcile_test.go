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

package addons

import (
	"fmt"
	"strings"
	"testing"
)

func TestKubectlCommand(t *testing.T) {
	expectedCommand := "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/%s/kubectl apply -f /etc/kubernetes/addons -l kubernetes.io/cluster-service!=true,addonmanager.kubernetes.io/mode=Reconcile --prune=true --prune-whitelist core/v1/ConfigMap --prune-whitelist core/v1/Endpoints --prune-whitelist core/v1/Namespace --prune-whitelist core/v1/PersistentVolumeClaim --prune-whitelist core/v1/PersistentVolume --prune-whitelist core/v1/Pod --prune-whitelist core/v1/ReplicationController --prune-whitelist core/v1/Secret --prune-whitelist core/v1/Service --prune-whitelist batch/v1/Job --prune-whitelist batch/v1beta1/CronJob --prune-whitelist apps/v1/DaemonSet --prune-whitelist apps/v1/Deployment --prune-whitelist apps/v1/ReplicaSet --prune-whitelist apps/v1/StatefulSet --prune-whitelist extensions/v1beta1/Ingress --recursive"

	tests := []struct {
		description string
		k8sVersion  string
		expected    string
	}{
		{
			description: "k8s version < 1.17.0",
			k8sVersion:  "v1.16.0",
			expected:    expectedCommand,
		}, {
			description: "k8s version == 1.17.0",
			k8sVersion:  "v1.17.0",
			expected:    expectedCommand + " --namespace=kube-system",
		}, {
			description: "k8s version > 1.17.0",
			k8sVersion:  "v1.18.0",
			expected:    expectedCommand + " --namespace=kube-system",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			originalK8sVersion := k8sVersion
			defer func() { k8sVersion = originalK8sVersion }()
			k8sVersion = func(_ string) (string, error) {
				return test.k8sVersion, nil
			}

			command, err := kubectlCommand("")
			if err != nil {
				t.Fatalf("error getting kubectl command: %v", err)
			}
			actual := strings.Join(command.Args, " ")

			expected := fmt.Sprintf(test.expected, test.k8sVersion)

			if actual != expected {
				t.Fatalf("expected does not match actual\nExpected: %s\nActual: %s", expected, actual)
			}
		})
	}
}
