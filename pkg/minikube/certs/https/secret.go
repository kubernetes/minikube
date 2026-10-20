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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	SecretName      = "minikube-tls"
	SecretNamespace = "kube-system"
)

// CreateOrUpdateTLSSecret creates or updates the minikube-tls secret in kube-system namespace
func CreateOrUpdateTLSSecret(client kubernetes.Interface, serverCertPEM, caCertPEM, serverKeyPEM []byte) error {
	return CreateOrUpdateTLSSecretInNamespace(client, SecretNamespace, serverCertPEM, caCertPEM, serverKeyPEM)
}

// CreateOrUpdateTLSSecretInNamespace creates or updates the minikube-tls secret in the specified namespace
func CreateOrUpdateTLSSecretInNamespace(client kubernetes.Interface, namespace string, serverCertPEM, caCertPEM, serverKeyPEM []byte) error {
	certChain := append([]byte{}, serverCertPEM...)
	if len(caCertPEM) > 0 {
		certChain = append(certChain, caCertPEM...)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "minikube",
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       certChain,
			corev1.TLSPrivateKeyKey: serverKeyPEM,
		},
	}

	secretsClient := client.CoreV1().Secrets(namespace)

	_, err := secretsClient.Get(context.Background(), SecretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err = secretsClient.Create(context.Background(), secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret %s/%s: %w", namespace, SecretName, err)
			}
			klog.Infof("Created TLS secret %s/%s in cluster", namespace, SecretName)
			return nil
		}
		return fmt.Errorf("failed to get secret %s/%s: %w", namespace, SecretName, err)
	}

	_, err = secretsClient.Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret %s/%s: %w", namespace, SecretName, err)
	}
	klog.Infof("Updated TLS secret %s/%s in cluster", namespace, SecretName)
	return nil
}
