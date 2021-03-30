/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"

	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var (
	webhookName       = "env-inject-webhook"
	webhookConfigName = "env-inject.zyanshu.io"
	skipLabel         = "auto-pause-skip"
)

// Create a clientset with in-cluster config.
func client() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	return clientset
}

// Retrieve the CA cert that will signed the cert used by the
// "GenericAdmissionWebhook" plugin admission controller.
func apiServerCert(clientset *kubernetes.Clientset) []byte {
	c, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "extension-apiserver-authentication", metav1.GetOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	pem, ok := c.Data["requestheader-client-ca-file"]
	if !ok {
		klog.Fatalf(fmt.Sprintf("cannot find the ca.crt in the configmap, configMap.Data is %#v", c.Data))
	}
	klog.Info("client-ca-file=", pem)
	return []byte(pem)
}

func configTLS(clientset *kubernetes.Clientset, serverCert []byte, serverKey []byte) *tls.Config {
	cert := apiServerCert(clientset)
	apiserverCA := x509.NewCertPool()
	apiserverCA.AppendCertsFromPEM(cert)

	sCert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		ClientCAs:    apiserverCA,
		ClientAuth:   tls.VerifyClientCertIfGiven, // TODO: actually require client cert
	}
}

// register this example webhook admission controller with the kube-apiserver
// by creating externalAdmissionHookConfigurations.
func selfRegistration(clientset *kubernetes.Clientset, caCert []byte) {
	client := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations()
	_, err := client.Get(context.TODO(), webhookName, metav1.GetOptions{})
	if err == nil {
		if err2 := client.Delete(context.TODO(), webhookName, metav1.DeleteOptions{}); err2 != nil {
			klog.Fatal(err2)
		}
	}
	var failurePolicy v1.FailurePolicyType = v1.Fail
	var sideEffects v1.SideEffectClass = v1.SideEffectClassNone

	webhookConfig := &v1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []v1.MutatingWebhook{
			{
				Name: webhookConfigName,
				Rules: []v1.RuleWithOperations{
					{
						Operations: []v1.OperationType{v1.Create, v1.Update},
						Rule: v1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
					{
						Operations: []v1.OperationType{v1.Create, v1.Update},
						Rule: v1.Rule{
							APIGroups:   []string{"extensions"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments"},
						},
					},
				},
				FailurePolicy: &failurePolicy,
				ObjectSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      skipLabel,
							Operator: metav1.LabelSelectorOpDoesNotExist,
						},
					},
				},
				ClientConfig: v1.WebhookClientConfig{
					Service: &v1.ServiceReference{
						Namespace: "auto-pause",
						Name:      "webhook",
					},
					CABundle: caCert,
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             &sideEffects,
			},
		},
	}
	if _, err := client.Create(context.TODO(), webhookConfig, metav1.CreateOptions{}); err != nil {
		klog.Fatalf("Client creation failed with %s", err)
	}
	log.Println("CLIENT CREATED")
}
