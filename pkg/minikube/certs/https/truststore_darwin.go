//go:build darwin

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

type darwinTrustStore struct{}

func newTrustStore() TrustStore {
	return &darwinTrustStore{}
}

func (s *darwinTrustStore) getKeychainPath() string {
	loginDb := filepath.Join(homedir.HomeDir(), "Library", "Keychains", "login.keychain-db")
	if _, err := os.Stat(loginDb); err == nil {
		return loginDb
	}
	return filepath.Join(homedir.HomeDir(), "Library", "Keychains", "login.keychain")
}

func (s *darwinTrustStore) Install(caCertPath string, profile string) error {
	securityPath, err := exec.LookPath("security")
	if err != nil {
		return fmt.Errorf("security CLI tool not found: %w", err)
	}

	keychain := s.getKeychainPath()
	caName := CAName(profile)

	// Remove existing cert if any
	_ = exec.Command(securityPath, "delete-certificate", "-c", caName, keychain).Run()

	// Add to login keychain as trusted root
	cmd := exec.Command(securityPath, "add-trusted-cert", "-d", "-r", "trustRoot", "-k", keychain, caCertPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add CA to macOS keychain: %s: %w", string(out), err)
	}

	klog.Infof("Successfully installed CA %q into macOS Login Keychain", caName)
	return nil
}

func (s *darwinTrustStore) Remove(profile string) error {
	securityPath, err := exec.LookPath("security")
	if err != nil {
		return nil
	}

	keychain := s.getKeychainPath()
	caName := CAName(profile)

	cmd := exec.Command(securityPath, "delete-certificate", "-c", caName, keychain)
	if out, err := cmd.CombinedOutput(); err != nil {
		klog.Warningf("Failed to remove CA %q from macOS keychain: %s", caName, string(out))
	} else {
		klog.Infof("Successfully removed CA %q from macOS Login Keychain", caName)
	}
	return nil
}
