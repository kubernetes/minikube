//go:build linux

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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

type linuxTrustStore struct{}

func newTrustStore() TrustStore {
	return &linuxTrustStore{}
}

func (s *linuxTrustStore) getNSSDBDir() string {
	return filepath.Join(homedir.HomeDir(), ".pki", "nssdb")
}

func (s *linuxTrustStore) ensureNSSDB() error {
	dbDir := s.getNSSDBDir()
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0700); err != nil {
			return fmt.Errorf("failed to create NSS DB directory %s: %w", dbDir, err)
		}
		// Initialize empty NSS DB without password
		dbPath := fmt.Sprintf("sql:%s", dbDir)
		cmd := exec.Command("certutil", "-d", dbPath, "-N", "--empty-password")
		if err := cmd.Run(); err != nil {
			klog.Warningf("Failed to initialize NSS database at %s: %v", dbDir, err)
		}
	}
	return nil
}

func (s *linuxTrustStore) Install(caCertPath string, profile string) error {
	certutilPath, err := exec.LookPath("certutil")
	if err != nil {
		return fmt.Errorf("certutil not found in PATH (please install libnss3-tools / nss-tools): %w", err)
	}

	if err := s.ensureNSSDB(); err != nil {
		return err
	}

	dbDir := s.getNSSDBDir()
	dbPath := fmt.Sprintf("sql:%s", dbDir)
	caName := CAName(profile)

	// Remove existing cert if any
	_ = exec.Command(certutilPath, "-d", dbPath, "-D", "-n", caName).Run()

	// Add cert to NSS database with trust flags for SSL/TLS ("TCu,cu,cu")
	cmd := exec.Command(certutilPath, "-d", dbPath, "-A", "-t", "TCu,cu,cu", "-n", caName, "-i", caCertPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add CA to NSS trust store: %s: %w", string(out), err)
	}

	klog.Infof("Successfully installed CA %q into Linux NSS trust store (%s)", caName, dbDir)
	return nil
}

func (s *linuxTrustStore) Remove(profile string) error {
	certutilPath, err := exec.LookPath("certutil")
	if err != nil {
		return nil // If certutil is not installed, nothing to remove
	}

	dbDir := s.getNSSDBDir()
	dbPath := fmt.Sprintf("sql:%s", dbDir)
	caName := CAName(profile)

	cmd := exec.Command(certutilPath, "-d", dbPath, "-D", "-n", caName)
	if out, err := cmd.CombinedOutput(); err != nil {
		klog.Warningf("Failed to remove CA %q from NSS trust store: %s", caName, string(out))
	} else {
		klog.Infof("Successfully removed CA %q from Linux NSS trust store", caName)
	}
	return nil
}
