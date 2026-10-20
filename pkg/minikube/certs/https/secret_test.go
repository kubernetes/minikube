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

package https

import (
	"context"
	"bytes"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateTLSSecret(t *testing.T) {
	client := fake.NewClientset()

	certPEM := []byte("FAKE CERT PEM")
	caCertPEM := []byte("FAKE CA PEM")
	keyPEM := []byte("FAKE KEY PEM")

	// Test initial creation
	err := CreateOrUpdateTLSSecret(client, certPEM, caCertPEM, keyPEM)
	if err != nil {
		t.Fatalf("CreateOrUpdateTLSSecret initial failed: %v", err)
	}

	secret, err := client.CoreV1().Secrets(SecretNamespace).Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get secret: %v", err)
	}

	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("expected secret type %v, got %v", corev1.SecretTypeTLS, secret.Type)
	}
	expectedChain := append([]byte("FAKE CERT PEM"), []byte("FAKE CA PEM")...)
	if !bytes.Equal(secret.Data[corev1.TLSCertKey], expectedChain) {
		t.Errorf("cert data mismatch")
	}
	if !bytes.Equal(secret.Data[corev1.TLSPrivateKeyKey], keyPEM) {
		t.Errorf("key data mismatch")
	}

	// Test update
	newCertPEM := []byte("NEW FAKE CERT PEM")
	newCACertPEM := []byte("NEW FAKE CA PEM")
	newKeyPEM := []byte("NEW FAKE KEY PEM")
	err = CreateOrUpdateTLSSecret(client, newCertPEM, newCACertPEM, newKeyPEM)
	if err != nil {
		t.Fatalf("CreateOrUpdateTLSSecret update failed: %v", err)
	}

	secret, err = client.CoreV1().Secrets(SecretNamespace).Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated secret: %v", err)
	}
	expectedNewChain := append([]byte("NEW FAKE CERT PEM"), []byte("NEW FAKE CA PEM")...)
	if !bytes.Equal(secret.Data[corev1.TLSCertKey], expectedNewChain) {
		t.Errorf("updated cert data mismatch")
	}
}
