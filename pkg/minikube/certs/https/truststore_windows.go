//go:build windows

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
	"os/exec"

	"k8s.io/klog/v2"
)

type windowsTrustStore struct{}

func newTrustStore() TrustStore {
	return &windowsTrustStore{}
}

func (s *windowsTrustStore) Install(caCertPath string, profile string) error {
	certutilPath, err := exec.LookPath("certutil.exe")
	if err != nil {
		certutilPath = "certutil.exe"
	}

	caName := CAName(profile)

	// Delete existing if any
	_ = exec.Command(certutilPath, "-user", "-delstore", "Root", caName).Run()

	// Add to user store Root
	cmd := exec.Command(certutilPath, "-user", "-addstore", "Root", caCertPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add CA to Windows user certificate store: %s: %w", string(out), err)
	}

	klog.Infof("Successfully installed CA %q into Windows user certificate store", caName)
	return nil
}

func (s *windowsTrustStore) Remove(profile string) error {
	certutilPath, err := exec.LookPath("certutil.exe")
	if err != nil {
		certutilPath = "certutil.exe"
	}

	caName := CAName(profile)

	cmd := exec.Command(certutilPath, "-user", "-delstore", "Root", caName)
	if out, err := cmd.CombinedOutput(); err != nil {
		klog.Warningf("Failed to remove CA %q from Windows user certificate store: %s", caName, string(out))
	} else {
		klog.Infof("Successfully removed CA %q from Windows user certificate store", caName)
	}
	return nil
}
