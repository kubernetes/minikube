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
	"net"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSetupHTTPSAndCleanup(t *testing.T) {
	profile := "test-https-profile"
	ip1 := net.ParseIP("192.168.49.2")
	ip2 := net.ParseIP("192.168.49.3")

	client := fake.NewClientset()

	// Clean up any left-over test files at the end
	defer func() {
		_ = os.RemoveAll(CACertPath(profile))
		_ = os.RemoveAll(ServerCertPath(profile))
		_ = os.RemoveAll(ServerKeyPath(profile))
		_ = CleanupHTTPS(profile)
	}()

	// Step 1: Initial SetupHTTPS
	if err := SetupHTTPS(profile, ip1, client); err != nil {
		t.Fatalf("SetupHTTPS failed: %v", err)
	}

	// Verify certificate files exist on disk
	if !CertificatesExist(profile) {
		t.Errorf("expected certificate files to exist on disk for profile %s", profile)
	}

	// Verify IP validity check
	if !IsCertValidForIP(profile, ip1) {
		t.Errorf("expected cert to be valid for IP %s", ip1)
	}
	if IsCertValidForIP(profile, ip2) {
		t.Errorf("expected cert NOT to be valid for IP %s yet", ip2)
	}

	// Verify Kubernetes secret in kube-system namespace
	secret, err := client.CoreV1().Secrets(SecretNamespace).Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to retrieve TLS secret from cluster: %v", err)
	}
	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("expected secret type %v, got %v", corev1.SecretTypeTLS, secret.Type)
	}
	if len(secret.Data[corev1.TLSCertKey]) == 0 || len(secret.Data[corev1.TLSPrivateKeyKey]) == 0 {
		t.Errorf("secret data tls.crt or tls.key is empty")
	}

	// Verify OS trust store installation
	verifyTrustStoreInstalled(t, profile)

	// Step 2: Idempotent SetupHTTPS call with same IP
	caModTime1, _ := os.Stat(CACertPath(profile))
	if err := SetupHTTPS(profile, ip1, client); err != nil {
		t.Fatalf("SetupHTTPS second call failed: %v", err)
	}
	caModTime2, _ := os.Stat(CACertPath(profile))
	if caModTime1.ModTime() != caModTime2.ModTime() {
		t.Errorf("expected certificate files to be preserved on identical IP call")
	}

	// Step 3: Call SetupHTTPS with new IP (ip2) -> triggers regeneration
	if err := SetupHTTPS(profile, ip2, client); err != nil {
		t.Fatalf("SetupHTTPS with new IP failed: %v", err)
	}
	if !IsCertValidForIP(profile, ip2) {
		t.Errorf("expected regenerated cert to be valid for new IP %s", ip2)
	}

	// Step 4: CleanupHTTPS
	if err := CleanupHTTPS(profile); err != nil {
		t.Fatalf("CleanupHTTPS failed: %v", err)
	}

	// Verify removal from OS trust store
	verifyTrustStoreRemoved(t, profile)
}

func TestSyncIngressHosts(t *testing.T) {
	profile := "test-sync-ingress-profile"
	ip := net.ParseIP("192.168.49.2")

	client := fake.NewClientset()
	defer func() {
		_ = os.RemoveAll(CACertPath(profile))
		_ = os.RemoveAll(ServerCertPath(profile))
		_ = os.RemoveAll(ServerKeyPath(profile))
		_ = CleanupHTTPS(profile)
	}()

	if err := SetupHTTPS(profile, ip, client); err != nil {
		t.Fatalf("SetupHTTPS failed: %v", err)
	}

	// Create an Ingress in default namespace with host myapp.example
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "myapp.example",
				},
			},
		},
	}
	_, err := client.NetworkingV1().Ingresses("default").Create(context.Background(), ing, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create fake ingress: %v", err)
	}

	if err := SyncIngressHosts(profile, ip, client); err != nil {
		t.Fatalf("SyncIngressHosts failed: %v", err)
	}

	// Verify minikube-tls secret was created in default namespace
	secret, err := client.CoreV1().Secrets("default").Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected minikube-tls secret to be synced to default namespace: %v", err)
	}
	if secret.Type != corev1.SecretTypeTLS {
		t.Errorf("expected secret type %v, got %v", corev1.SecretTypeTLS, secret.Type)
	}
}
