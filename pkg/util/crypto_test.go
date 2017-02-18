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
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
)

func TestGenerateCACert(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Error generating tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
