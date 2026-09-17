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
	"io"
	"log"
	"os"
	"path/filepath"
)

// Disable reverts containerd config files and restarts containerd
func Disable() error {
	log.Print("Disabling gvisor...")
	if err := disableAt(nodeDir); err != nil {
		return err
	}
	// restart containerd
	if err := restartContainerd(); err != nil {
		return fmt.Errorf("restarting containerd: %w", err)
	}
	log.Print("Successfully disabled gvisor")
	return nil
}

func disableAt(root string) error {
	if err := restoreConfig(root); err != nil {
		return err
	}
	if err := removeGvisorFiles(root); err != nil {
		return err
	}
	return nil
}

// restoreConfig puts the pre-gvisor containerd config back atomically. It
// never deletes the live config first, so a missing backup fails loudly
// instead of leaving the node configless.
func restoreConfig(root string) error {
	backupPath := filepath.Join(root, containerdConfigBackupPath)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		legacy := filepath.Join(root, legacyContainerdConfigBackupPath)
		if _, err := os.Stat(legacy); err != nil {
			return fmt.Errorf("no backup at %s (or legacy %s): refusing to touch %s",
				containerdConfigBackupPath, legacyContainerdConfigBackupPath, containerdConfigPath)
		}
		backupPath = legacy
	}
	configPath := filepath.Join(root, containerdConfigPath)
	tmp, err := os.CreateTemp(filepath.Dir(configPath), ".config.toml-*")
	if err != nil {
		return fmt.Errorf("creating temp config: %w", err)
	}
	tmpName := tmp.Name()
	src, err := os.Open(backupPath)
	if err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("opening backup: %w", err)
	}
	if _, err := io.Copy(tmp, src); err != nil {
		src.Close()
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("copying backup: %w", err)
	}
	src.Close()
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp config: %w", err)
	}
	log.Printf("Restoring default config.toml at %s", containerdConfigPath)
	if err := os.Rename(tmpName, configPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("reverting back to default config.toml: %w", err)
	}
	return nil
}

// removeGvisorFiles cleans up the binaries and log dirs Enable created.
// Missing paths are fine, so disable stays idempotent.
func removeGvisorFiles(root string) error {
	binDir := filepath.Join(root, "usr/bin")
	for _, p := range []string{
		filepath.Join(binDir, "runsc"),
		filepath.Join(binDir, "containerd-shim-runsc-v1"),
		filepath.Join(binDir, "gvisor-bin"),
		filepath.Join(root, "run/containerd/runsc"),
		filepath.Join(root, "tmp/runsc"),
	} {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("removing %s: %w", p, err)
		}
	}
	return nil
}
