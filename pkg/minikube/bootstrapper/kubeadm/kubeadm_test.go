/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"path"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

func TestControlPlaneArtifactsExist(t *testing.T) {
	expectedCmd := "sudo ls " + strings.Join([]string{
		"/etc/kubernetes/admin.conf",
		"/etc/kubernetes/kubelet.conf",
		"/etc/kubernetes/controller-manager.conf",
		"/etc/kubernetes/scheduler.conf",
		path.Join(vmpath.GuestManifestsDir, "kube-apiserver.yaml"),
		path.Join(vmpath.GuestManifestsDir, "kube-controller-manager.yaml"),
		path.Join(vmpath.GuestManifestsDir, "kube-scheduler.yaml"),
		path.Join(vmpath.GuestManifestsDir, "etcd.yaml"),
	}, " ")

	t.Run("all artifacts present", func(t *testing.T) {
		r := command.NewFakeCommandRunner()
		r.SetCommandToOutput(map[string]string{
			expectedCmd: "",
		})
		b := &Bootstrapper{c: r}
		if !b.controlPlaneArtifactsExist() {
			t.Errorf("controlPlaneArtifactsExist() = false, want true")
		}
	})

	t.Run("artifacts missing", func(t *testing.T) {
		r := command.NewFakeCommandRunner()
		b := &Bootstrapper{c: r}
		if b.controlPlaneArtifactsExist() {
			t.Errorf("controlPlaneArtifactsExist() = true, want false")
		}
	})
}
