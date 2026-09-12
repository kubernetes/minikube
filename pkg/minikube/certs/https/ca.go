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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// Certificate file names in profile directory
const (
	CACertFile     = "ingress-ca.crt"
	ServerCertFile = "ingress-tls.crt"
	ServerKeyFile  = "ingress-tls.key"
)

// CACertPath returns the path to the profile's CA certificate
func CACertPath(profile string) string {
	return filepath.Join(localpath.Profile(profile), CACertFile)
}

// ServerCertPath returns the path to the profile's server TLS certificate
func ServerCertPath(profile string) string {
	return filepath.Join(localpath.Profile(profile), ServerCertFile)
}

// ServerKeyPath returns the path to the profile's server TLS key
func ServerKeyPath(profile string) string {
	return filepath.Join(localpath.Profile(profile), ServerKeyFile)
}

// GenerateCerts creates an in-memory CA and signs a server certificate for the minikube cluster.
// The CA private key is generated in memory, used only to sign the server cert, and discarded.
func GenerateCerts(profile string, clusterIP net.IP, extraSANs ...string) (caCertPEM []byte, serverCertPEM []byte, serverKeyPEM []byte, err error) {
	// 1. Generate CA private key in memory
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	caCommonName := fmt.Sprintf("minikube-%s", profile)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	caSerial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA serial number: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: caSerial,
		Subject: pkix.Name{
			CommonName:   caCommonName,
			Organization: []string{"minikube"},
		},
		NotBefore:             time.Now().Add(-10 * time.Minute),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	// 2. Generate Server private key
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	serverSerial, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate server serial number: %w", err)
	}

	sansMap := map[string]bool{
		fmt.Sprintf("*.%s.minikube.internal", profile): true,
		"*.nip.io":   true,
		"*.sslip.io": true,
		"*.local":    true,
		"localhost":  true,
	}

	ips := []net.IP{net.ParseIP("127.0.0.1")}
	if clusterIP != nil {
		ips = append(ips, clusterIP)
	}

	for _, extra := range extraSANs {
		if extra == "" {
			continue
		}
		if ip := net.ParseIP(extra); ip != nil {
			ips = append(ips, ip)
		} else {
			sansMap[extra] = true
		}
	}

	sans := make([]string, 0, len(sansMap))
	for s := range sansMap {
		sans = append(sans, s)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: serverSerial,
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("%s-server", caCommonName),
			Organization: []string{"minikube"},
		},
		NotBefore:             time.Now().Add(-10 * time.Minute),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              sans,
		IPAddresses:           ips,
		BasicConstraintsValid: true,
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	serverCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})

	serverKeyBytes, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal server private key: %w", err)
	}
	serverKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyBytes})

	// CA key is discarded here as it goes out of scope and is not saved anywhere.
	return caCertPEM, serverCertPEM, serverKeyPEM, nil
}

// SaveCertificates saves the CA cert, server cert, and server key to the profile directory.
func SaveCertificates(profile string, caCertPEM, serverCertPEM, serverKeyPEM []byte) error {
	profileDir := localpath.Profile(profile)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	if err := os.WriteFile(CACertPath(profile), caCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write CA cert: %w", err)
	}

	if err := os.WriteFile(ServerCertPath(profile), serverCertPEM, 0644); err != nil {
		return fmt.Errorf("failed to write server cert: %w", err)
	}

	if err := os.WriteFile(ServerKeyPath(profile), serverKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write server key: %w", err)
	}

	return nil
}

// CertificatesExist checks if all HTTPS certificates exist in the profile directory.
func CertificatesExist(profile string) bool {
	if _, err := os.Stat(CACertPath(profile)); err != nil {
		return false
	}
	if _, err := os.Stat(ServerCertPath(profile)); err != nil {
		return false
	}
	if _, err := os.Stat(ServerKeyPath(profile)); err != nil {
		return false
	}
	return true
}

// IsCertValidForIP checks whether the existing server certificate in the profile directory includes the cluster IP.
func IsCertValidForIP(profile string, ip net.IP) bool {
	if ip == nil {
		return false
	}
	certPEM, err := os.ReadFile(ServerCertPath(profile))
	if err != nil {
		return false
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	// Check if IP is included in IPAddresses
	for _, certIP := range cert.IPAddresses {
		if certIP.Equal(ip) {
			return true
		}
	}
	return false
}
