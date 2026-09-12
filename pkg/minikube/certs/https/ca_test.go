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
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
)

func TestGenerateCerts(t *testing.T) {
	profile := "test-profile"
	clusterIP := net.ParseIP("192.168.64.3")

	caCertPEM, serverCertPEM, serverKeyPEM, err := GenerateCerts(profile, clusterIP)
	if err != nil {
		t.Fatalf("GenerateCerts failed: %v", err)
	}

	if len(caCertPEM) == 0 {
		t.Errorf("expected non-empty caCertPEM")
	}
	if len(serverCertPEM) == 0 {
		t.Errorf("expected non-empty serverCertPEM")
	}
	if len(serverKeyPEM) == 0 {
		t.Errorf("expected non-empty serverKeyPEM")
	}

	// Parse CA Cert
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		t.Fatalf("failed to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse CA cert: %v", err)
	}
	if caCert.Subject.CommonName != "minikube-test-profile" {
		t.Errorf("unexpected CA CN: got %q, want %q", caCert.Subject.CommonName, "minikube-test-profile")
	}
	if !caCert.IsCA {
		t.Errorf("expected CA cert to have IsCA=true")
	}

	// Parse Server Cert
	block, _ = pem.Decode(serverCertPEM)
	if block == nil {
		t.Fatalf("failed to decode Server cert PEM")
	}
	serverCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse Server cert: %v", err)
	}

	// Verify SANs
	expectedDNS := "*.test-profile.minikube.internal"
	foundDNS := false
	for _, dns := range serverCert.DNSNames {
		if dns == expectedDNS {
			foundDNS = true
			break
		}
	}
	if !foundDNS {
		t.Errorf("expected DNS SAN %q in server cert SANs: %v", expectedDNS, serverCert.DNSNames)
	}

	foundIP := false
	for _, ip := range serverCert.IPAddresses {
		if ip.Equal(clusterIP) {
			foundIP = true
			break
		}
	}
	if !foundIP {
		t.Errorf("expected IP SAN %v in server cert IP addresses: %v", clusterIP, serverCert.IPAddresses)
	}

	// Verify server cert is signed by CA cert
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots:       roots,
		CurrentTime: serverCert.NotBefore.Add(1 * timeMinute),
		DNSName:     "app.test-profile.minikube.internal",
	}

	if _, err := serverCert.Verify(opts); err != nil {
		t.Errorf("server cert verification against CA failed: %v", err)
	}
}

const timeMinute = 60 * 1000 * 1000 * 1000
