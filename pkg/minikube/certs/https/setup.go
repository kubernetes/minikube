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
	"net"
	"os"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// SetupHTTPS orchestrates the zero-configuration trusted HTTPS setup:
// 1. Checks if CA and cert exist and IP hasn't changed.
// 2. Generates in-memory CA and signs cluster TLS cert.
// 3. Installs CA cert in user-level trust store.
// 4. Persists CA cert & server cert/key in profile directory.
// 5. Installs TLS secret (minikube-tls) in kube-system namespace.
// 6. Discards CA key from memory.
func SetupHTTPS(profile string, clusterIP net.IP, client kubernetes.Interface) error {
	caPath := CACertPath(profile)
	serverCertPath := ServerCertPath(profile)
	serverKeyPath := ServerKeyPath(profile)

	var caCertPEM, serverCertPEM, serverKeyPEM []byte
	var err error

	if CertificatesExist(profile) && IsCertValidForIP(profile, clusterIP) {
		klog.Infof("Existing HTTPS certificates for profile %q are valid for IP %s; skipping cert generation", profile, clusterIP)
		caCertPEM, err = os.ReadFile(caPath)
		if err != nil {
			return fmt.Errorf("failed to read CA cert: %w", err)
		}
		serverCertPEM, err = os.ReadFile(serverCertPath)
		if err != nil {
			return fmt.Errorf("failed to read server cert: %w", err)
		}
		serverKeyPEM, err = os.ReadFile(serverKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read server key: %w", err)
		}
	} else {
		ts := NewTrustStore()
		// Remove stale CA from trust store if IP changed or regenerating
		_ = ts.Remove(profile)

		klog.Infof("Generating new HTTPS CA and TLS certificates for profile %q with IP %s", profile, clusterIP)
		caCertPEM, serverCertPEM, serverKeyPEM, err = GenerateCerts(profile, clusterIP)
		if err != nil {
			return fmt.Errorf("failed to generate HTTPS certificates: %w", err)
		}

		if err := SaveCertificates(profile, caCertPEM, serverCertPEM, serverKeyPEM); err != nil {
			return fmt.Errorf("failed to save HTTPS certificates: %w", err)
		}

	}

	ts := NewTrustStore()
	if err := ts.Install(caPath, profile); err != nil {
		out.WarningT("Failed to install CA certificate into user trust store: {{.error}}", out.V{"error": err})
	} else {
		out.Step(style.Check, "Installed trusted CA {{.ca}} in the user trust store", out.V{"ca": CAName(profile)})
	}

	if client != nil {
		if err := CreateOrUpdateTLSSecret(client, serverCertPEM, caCertPEM, serverKeyPEM); err != nil {
			out.WarningT("Failed to create/update minikube-tls secret in cluster: {{.error}}", out.V{"error": err})
		} else {
			out.Step(style.Check, "Installed TLS certificate for {{.ip}} in the cluster", out.V{"ip": clusterIP})
		}
		_ = EnsureTraefikTLSStore(client)
		_ = SyncIngressHosts(profile, clusterIP, client)
	}

	return nil
}

// EnsureTraefikTLSStore configures Traefik default TLSStore to use minikube-tls secret
func EnsureTraefikTLSStore(client kubernetes.Interface) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = nil
		}
	}()
	if client == nil || client.CoreV1() == nil || client.CoreV1().RESTClient() == nil {
		return nil
	}
	body := []byte(`{
		"apiVersion": "traefik.io/v1alpha1",
		"kind": "TLSStore",
		"metadata": {
			"name": "default",
			"namespace": "kube-system"
		},
		"spec": {
			"defaultCertificate": {
				"secretName": "minikube-tls"
			}
		}
	}`)
	restClient := client.CoreV1().RESTClient()
	err = restClient.Post().
		AbsPath("/apis/traefik.io/v1alpha1/namespaces/kube-system/tlsstores").
		Body(body).
		Do(context.Background()).
		Error()
	if err != nil {
		klog.V(2).Infof("Traefik default TLSStore setup: %v", err)
	}
	return nil
}

// SyncIngressHosts scans all cluster Ingress resources, extracts hosts, re-issues certs if needed,
// and syncs the minikube-tls secret into namespaces where Ingresses are deployed.
func SyncIngressHosts(profile string, clusterIP net.IP, client kubernetes.Interface) error {
	if client == nil {
		return nil
	}

	ingresses, err := client.NetworkingV1().Ingresses("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list ingresses for HTTPS host sync: %w", err)
	}

	caPath := CACertPath(profile)
	serverCertPath := ServerCertPath(profile)
	serverKeyPath := ServerKeyPath(profile)

	caCertPEM, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA cert: %w", err)
	}
	serverCertPEM, err := os.ReadFile(serverCertPath)
	if err != nil {
		return fmt.Errorf("failed to read server cert: %w", err)
	}
	serverKeyPEM, err := os.ReadFile(serverKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read server key: %w", err)
	}

	hostMap := make(map[string]bool)
	namespaces := make(map[string]bool)
	namespaces[SecretNamespace] = true

	for _, ing := range ingresses.Items {
		namespaces[ing.Namespace] = true
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hostMap[rule.Host] = true
			}
		}
	}

	var hosts []string
	for h := range hostMap {
		hosts = append(hosts, h)
	}

	// Re-issue certificate if new hosts were discovered
	if len(hosts) > 0 {
		klog.Infof("Syncing %d Ingress host(s) into HTTPS TLS certificate for profile %q: %v", len(hosts), profile, hosts)
		caCertPEM, serverCertPEM, serverKeyPEM, err = GenerateCerts(profile, clusterIP, hosts...)
		if err == nil {
			_ = SaveCertificates(profile, caCertPEM, serverCertPEM, serverKeyPEM)
			ts := NewTrustStore()
			_ = ts.Install(caPath, profile)
		}
	}

	// Sync secret into all namespaces where ingresses exist
	for ns := range namespaces {
		if err := CreateOrUpdateTLSSecretInNamespace(client, ns, serverCertPEM, caCertPEM, serverKeyPEM); err != nil {
			klog.Warningf("Failed to sync minikube-tls secret to namespace %s: %v", ns, err)
		}
	}

	_ = SyncEtcHosts(clusterIP, hosts)

	return nil
}

// SyncEtcHosts updates /etc/hosts with the cluster IP and discovered Ingress hosts
func SyncEtcHosts(clusterIP net.IP, hosts []string) error {
	if clusterIP == nil || len(hosts) == 0 {
		return nil
	}
	hostsPath := "/etc/hosts"
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return nil
	}

	startMarker := "# minikube-hosts-start"
	endMarker := "# minikube-hosts-end"

	var cleanHosts []string
	for _, h := range hosts {
		if h != "" && !strings.Contains(h, "*") {
			cleanHosts = append(cleanHosts, h)
		}
	}
	if len(cleanHosts) == 0 {
		return nil
	}

	hostLine := fmt.Sprintf("%s %s", clusterIP.String(), strings.Join(cleanHosts, " "))
	block := fmt.Sprintf("%s\n%s\n%s\n", startMarker, hostLine, endMarker)

	strContent := string(content)
	if strings.Contains(strContent, startMarker) {
		re := regexp.MustCompile(fmt.Sprintf("(?s)%s.*?%s\n?", regexp.QuoteMeta(startMarker), regexp.QuoteMeta(endMarker)))
		strContent = re.ReplaceAllString(strContent, block)
	} else {
		strContent += "\n" + block
	}

	if err := os.WriteFile(hostsPath, []byte(strContent), 0644); err != nil {
		klog.V(2).Infof("Could not auto-update /etc/hosts (requires root privileges): %v", err)
		return err
	}
	klog.Infof("Auto-updated /etc/hosts with %d Ingress host(s)", len(cleanHosts))
	return nil
}

// CleanupHTTPS removes the CA cert from the user-level trust store on cluster deletion.
func CleanupHTTPS(profile string) error {
	ts := NewTrustStore()
	return ts.Remove(profile)
}
