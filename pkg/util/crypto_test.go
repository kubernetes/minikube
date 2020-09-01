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

package util

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
)

func TestGenerateCACert(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	defer func() { // clean up tempdir
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", tmpDir)
		}
	}()
	if err != nil {
		t.Fatalf("Error generating tmpdir: %v", err)
	}

	certPath := filepath.Join(tmpDir, "cert")
	keyPath := filepath.Join(tmpDir, "key")
	if err := GenerateCACert(certPath, keyPath, constants.APIServerName); err != nil {
		t.Fatalf("GenerateCACert() error = %v", err)
	}

	// Check the cert has the right shape.
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Error reading cert data: %v", err)
	}
	data, _ := pem.Decode(certBytes)
	c, err := x509.ParseCertificate(data.Bytes)
	if err != nil {
		t.Fatalf("Error parsing certificate: %v", err)
	}
	if !c.IsCA {
		t.Fatalf("Cert is not a CA cert.")
	}
}

func TestGenerateSignedCert(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	defer func() { // clean up tempdir
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", tmpDir)
		}
	}()
	if err != nil {
		t.Fatalf("Error generating tmpdir: %v", err)
	}

	signerTmpDir, err := ioutil.TempDir("", "")
	defer func() { // clean up tempdir
		err := os.RemoveAll(signerTmpDir)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", signerTmpDir)
		}
	}()
	if err != nil {
		t.Fatalf("Error generating signer tmpdir: %v", err)
	}

	validSignerCertPath := filepath.Join(signerTmpDir, "cert")
	validSignerKeyPath := filepath.Join(signerTmpDir, "key")

	err = GenerateCACert(validSignerCertPath, validSignerKeyPath, constants.APIServerName)
	if err != nil {
		t.Fatalf("Error generating signer cert")
	}

	certPath := filepath.Join(tmpDir, "cert")
	keyPath := filepath.Join(tmpDir, "key")

	ips := []net.IP{net.ParseIP("192.168.99.100"), net.ParseIP("10.0.0.10")}
	alternateDNS := []string{"kubernetes.default.svc.cluster.local", "kubernetes.default"}

	tests := []struct {
		description    string
		signerCertPath string
		signerKeyPath  string
		err            bool
	}{
		{
			description:    "wrong cert path",
			signerCertPath: "",
			signerKeyPath:  validSignerKeyPath,
			err:            true,
		},
		{
			description:    "wrong key path",
			signerCertPath: validSignerCertPath,
			signerKeyPath:  "",
			err:            true,
		},
		{
			description:    "valid cert",
			signerCertPath: validSignerCertPath,
			signerKeyPath:  validSignerKeyPath,
		},
		{
			description:    "wrong key file",
			signerCertPath: validSignerCertPath,
			signerKeyPath:  validSignerCertPath,
			err:            true,
		},
		{
			description:    "wrong cert file",
			signerCertPath: validSignerKeyPath,
			signerKeyPath:  validSignerKeyPath,
			err:            true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			err := GenerateSignedCert(
				certPath, keyPath, "minikube", ips, alternateDNS, test.signerCertPath,
				test.signerKeyPath,
			)
			if err != nil && !test.err {
				t.Errorf("GenerateSignedCert() error = %v", err)
			}
			if err == nil && test.err {
				t.Errorf("GenerateSignedCert() should have returned error, but didn't")
			}
			if err == nil {
				certBytes, err := ioutil.ReadFile(certPath)
				if err != nil {
					t.Errorf("Error reading cert data: %v", err)
				}
				data, _ := pem.Decode(certBytes)
				_, err = x509.ParseCertificate(data.Bytes)
				if err != nil {
					t.Errorf("Error parsing certificate: %v", err)
				}
			}
		})
	}
}
