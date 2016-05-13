package cert

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCACertificate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	// cleanup
	defer os.RemoveAll(tmpDir)

	caCertPath := filepath.Join(tmpDir, "ca.pem")
	caKeyPath := filepath.Join(tmpDir, "key.pem")
	testOrg := "test-org"
	bits := 2048
	if err := GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(caCertPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(caKeyPath); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateCert(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	// cleanup
	defer os.RemoveAll(tmpDir)

	caCertPath := filepath.Join(tmpDir, "ca.pem")
	caKeyPath := filepath.Join(tmpDir, "key.pem")
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "cert-key.pem")
	testOrg := "test-org"
	bits := 2048
	if err := GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(caCertPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(caKeyPath); err != nil {
		t.Fatal(err)
	}

	if err := GenerateCert([]string{}, certPath, keyPath, caCertPath, caKeyPath, testOrg, bits); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Fatalf("certificate not created at %s", certPath)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key not created at %s", keyPath)
	}
}
