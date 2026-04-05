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

package gvisor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/minikube/pkg/libmachine/mcnutils"
)

// Disable reverts containerd config files and restarts containerd
func Disable() error {
	log.Print("Disabling gvisor...")
	if err := os.Remove(filepath.Join(nodeDir, containerdConfigPath)); err != nil {
		return fmt.Errorf("removing %s: %w", containerdConfigPath, err)
	}
	log.Printf("Restoring default config.toml at %s", containerdConfigPath)
	if err := mcnutils.CopyFile(filepath.Join(nodeDir, containerdConfigBackupPath), filepath.Join(nodeDir, containerdConfigPath)); err != nil {
		return fmt.Errorf("reverting back to default config.toml: %w", err)
	}
	// restart containerd
	if err := restartContainerd(); err != nil {
		return fmt.Errorf("restarting containerd: %w", err)
	}
	log.Print("Successfully disabled gvisor")
	return nil
}
