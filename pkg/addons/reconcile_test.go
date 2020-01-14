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
	"strings"
	"testing"
)

func TestKubectlCommand(t *testing.T) {
	expected := "sudo KUBECONFIG=/var/lib/minikube/kubeconfig /var/lib/minikube/binaries/v1.13.2/kubectl apply -f /etc/kubernetes/addons -l kubernetes.io/cluster-service!=true,addonmanager.kubernetes.io/mode=Reconcile --prune=true --prune-whitelist core/v1/ConfigMap --prune-whitelist core/v1/Endpoints --prune-whitelist core/v1/Namespace --prune-whitelist core/v1/PersistentVolumeClaim --prune-whitelist core/v1/PersistentVolume --prune-whitelist core/v1/Pod --prune-whitelist core/v1/ReplicationController --prune-whitelist core/v1/Secret --prune-whitelist core/v1/Service --prune-whitelist batch/v1/Job --prune-whitelist batch/v1beta1/CronJob --prune-whitelist apps/v1/DaemonSet --prune-whitelist apps/v1/Deployment --prune-whitelist apps/v1/ReplicaSet --prune-whitelist apps/v1/StatefulSet --prune-whitelist extensions/v1beta1/Ingress --recursive"
	kc := kubectlCommand()
	actual := strings.Join(kc.Args, " ")

	if actual != expected {
		t.Fatalf("expected does not match actual\nExpected: %s\nActual: %s", expected, actual)
	}
}
