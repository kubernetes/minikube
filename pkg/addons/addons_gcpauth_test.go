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

package addons

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	k8stesting "k8s.io/client-go/testing"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/service"
)

type fakeK8sClient struct {
	core typed_core.CoreV1Interface
}

func (f *fakeK8sClient) GetCoreClient(string) (typed_core.CoreV1Interface, error) {
	return f.core, nil
}

// TestPatchServiceAccountsRetriesTransientErrors covers the race where gcp-auth is enabled
// during `minikube start` before the apiserver accepts connections.
func TestPatchServiceAccountsRetriesTransientErrors(t *testing.T) {
	client := fake.NewClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"}},
		&corev1.ServiceAccount{
			ObjectMeta:       metav1.ObjectMeta{Name: "already-patched", Namespace: "default"},
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: secretName}},
		},
	)

	failures := 2
	client.PrependReactor("list", "namespaces", func(k8stesting.Action) (bool, runtime.Object, error) {
		if failures > 0 {
			failures--
			return true, nil, errors.New(`Get "https://192.168.49.2:8443/api/v1/namespaces": dial tcp 192.168.49.2:8443: connect: connection refused`)
		}
		return false, nil, nil
	})

	orig := service.K8s
	service.K8s = &fakeK8sClient{core: client.CoreV1()}
	t.Cleanup(func() { service.K8s = orig })

	if err := patchServiceAccounts(&config.ClusterConfig{Name: "minikube"}); err != nil {
		t.Fatalf("patchServiceAccounts() returned error: %v", err)
	}
	if failures != 0 {
		t.Fatalf("expected the injected failures to be consumed, %d left", failures)
	}

	for _, name := range []string{"default", "already-patched"} {
		sa, err := client.CoreV1().ServiceAccounts("default").Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("getting service account %q: %v", name, err)
		}
		want := []corev1.LocalObjectReference{{Name: secretName}}
		if len(sa.ImagePullSecrets) != 1 || sa.ImagePullSecrets[0] != want[0] {
			t.Errorf("service account %q image pull secrets = %v, want %v", name, sa.ImagePullSecrets, want)
		}
	}
}
