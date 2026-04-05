/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// These components do not support before/after version comparison
	// TODO: add support before/after https://github.com/kubernetes/minikube/issues/21246
	noVersionCheck = map[string]bool{
		// All components now support version checking!
	}

	// Skip these components from auto-updating
	skipList = map[string]bool{
		"get_version":                   true, // self (internal)
		"update_all":                    true, // self
		"k8s-lib":                       true, // not needed anymore (TODO: remove in future)
		"amd_device_gpu_plugin_version": true, // sem vers issue https://github.com/ROCm/k8s-device-plugin/issues/144eadm_auto_build
		"istio_operator_version":        true, // till this is fixed https://github.com/istio/istio/issues/57185
		"kicbase_version":               true, // This one is not related to auto updating, this is a tool used by kicbae_auto_build
		"preload_version":               true, // This is an internal tool to bump the preload version, not a component update
	}
)

func shouldSkip(component string) bool {
	if runtime.GOOS != "linux" && component == "kubeadm_constants" { // kubeadm constants update job only works on linux
		log.Printf("Skipping %s on non-linux OS: %s", component, runtime.GOOS)
		return true
	}
	return skipList[component]
}

func getVersion(component string) (string, error) {
	cmd := exec.Command("go", "run", "update/get_version/get_version.go")
	cmd.Env = append(os.Environ(), fmt.Sprintf("DEP=%s", component))

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		log.Printf("failed to get version for %s: %v", component, err)
		log.Printf("command output: %s", out.String())
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func main() {
	const updateDir = "update"
	const maxUpdates = 100

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting update process in:", cwd)

	entries, err := os.ReadDir(updateDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", updateDir, err)
	}

	var changes []string
	updatesRun := 0

	for _, entry := range entries {
		if !entry.IsDir() || updatesRun >= maxUpdates {
			continue
		}

		component := entry.Name()
		if shouldSkip(component) {
			continue
		}

		log.Printf("Processing %s...\n", component)

		var oldVersion string
		if !noVersionCheck[component] {
			oldVersion, err = getVersion(component)
			if err != nil {
				log.Fatalf("Could not get old version for %s: %v", component, err)
			}
		}

		script := filepath.Join(updateDir, component, fmt.Sprintf("%s.go", component))
		cmd := exec.Command("go", "run", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to update %s: %v", component, err)
		}

		if !noVersionCheck[component] {
			newVersion, err := getVersion(component)
			if err != nil {
				log.Printf("Could not get new version for %s: %v", component, err)
				continue
			}
			if oldVersion != newVersion {
				change := fmt.Sprintf("- **%s:** `%s` -> `%s`", component, oldVersion, newVersion)
				changes = append(changes, change)
				fmt.Println(change)
			} else {
				fmt.Printf("No change for %s.\n", component)
			}
			fmt.Println()
		}

		updatesRun++
	}

	fmt.Println("---")
	fmt.Println("Updated components summary:")
	fmt.Println(strings.Join(changes, "\n"))

	// Output machine-readable summary for GitHub Actions
	fmt.Printf("updates<<EOF\n%s\nEOF\n", strings.Join(changes, "\n"))
}
