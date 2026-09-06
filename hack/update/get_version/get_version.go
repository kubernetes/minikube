/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const versionsFile = "../pkg/minikube/assets/versions.json"

func main() {
	depName := os.Getenv("DEP")
	if depName == "" {
		log.Fatalf("the environment variable 'DEP' needs to be set")
	}
	depName = standrizeComponentName(depName)

	// Handle special cases not stored as simple version strings in versions.json
	switch depName {
	case "docsy":
		version, err := getDocsyVersion()
		if err != nil {
			log.Fatalf("failed to get docsy version: %v", err)
		}
		os.Stdout.WriteString(version)
		return
	case "kubeadm-constants":
		version, err := getKubeadmConstantsVersion()
		if err != nil {
			log.Fatalf("failed to get kubeadm constants version: %v", err)
		}
		os.Stdout.WriteString(version)
		return
	case "kubernetes":
		version, err := getKubernetesVersion()
		if err != nil {
			log.Fatalf("failed to get kubernetes version: %v", err)
		}
		os.Stdout.WriteString(version)
		return
	case "kubernetes-versions-list":
		version, err := getKubernetesVersionsList()
		if err != nil {
			log.Fatalf("failed to get kubernetes versions list: %v", err)
		}
		os.Stdout.WriteString(version)
		return
	case "site-node":
		depName = "node"
	}

	data, err := os.ReadFile(versionsFile)
	if err != nil {
		log.Fatalf("failed to read versions.json: %v", err)
	}

	var versions map[string]string
	if err := json.Unmarshal(data, &versions); err != nil {
		log.Fatalf("failed to parse versions.json: %v", err)
	}

	version, ok := versions[depName]
	if !ok {
		log.Fatalf("%s is not a valid dependency in versions.json", depName)
	}

	os.Stdout.WriteString(version)
}

// some components have _ or - in their names vs their make folders, standardizing for automation such as as update-all
func standrizeComponentName(name string) string {
	// Convert the component name to lowercase and replace underscores with hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")

	// Remove "-version" suffix only at the end to avoid breaking words like "versions"
	name = strings.TrimSuffix(name, "-version")
	return name
}

// getDocsyVersion returns the current commit hash of the docsy submodule
func getDocsyVersion() (string, error) {
	// Change to parent directory since we're running from hack/
	cmd := exec.Command("git", "submodule", "status", "site/themes/docsy")
	cmd.Dir = ".." // Change to the repo root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Output format: " commit-hash path/to/submodule (tag or branch)"
	// We want just the commit hash (first 8 characters for short hash)
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return "", log.New(os.Stderr, "", 0).Output(1, "no commit hash found in git submodule status")
	}

	commitHash := strings.TrimSpace(parts[0])
	// Remove leading space or other characters and take first 8 characters
	if len(commitHash) > 8 {
		commitHash = commitHash[:8]
	}
	return commitHash, nil
}

// getKubeadmConstantsVersion returns a summary of kubeadm constants versions
func getKubeadmConstantsVersion() (string, error) {
	// Read the constants file to get a representative version
	data, err := os.ReadFile("../pkg/minikube/constants/constants_kubeadm_images.go")
	if err != nil {
		return "", err
	}

	// Look for the latest kubernetes version entry in the KubeadmImages map
	re := regexp.MustCompile(`"(v\d+\.\d+\.\d+[^"]*)":\s*{`)
	matches := re.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return "no-versions", nil
	}

	// Return the last (latest) version found
	lastMatch := matches[len(matches)-1]
	return string(lastMatch[1]), nil
}

// getKubernetesVersion returns the default kubernetes version
func getKubernetesVersion() (string, error) {
	data, err := os.ReadFile("../pkg/minikube/constants/constants.go")
	if err != nil {
		return "", err
	}

	// Look for DefaultKubernetesVersion
	re := regexp.MustCompile(`DefaultKubernetesVersion = "(.*?)"`)
	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return "", log.New(os.Stderr, "", 0).Output(1, "DefaultKubernetesVersion not found")
	}

	return string(matches[1]), nil
}

// getKubernetesVersionsList returns a count of supported kubernetes versions
func getKubernetesVersionsList() (string, error) {
	data, err := os.ReadFile("../pkg/minikube/constants/constants_kubernetes_versions.go")
	if err != nil {
		return "", err
	}

	// Count the number of versions in ValidKubernetesVersions
	re := regexp.MustCompile(`"v\d+\.\d+\.\d+[^"]*"`)
	matches := re.FindAll(data, -1)

	if len(matches) == 0 {
		return "0-versions", nil
	}

	// Return count and range if available
	if len(matches) >= 2 {
		first := string(matches[0])
		last := string(matches[len(matches)-1])
		first = strings.Trim(first, "\"")
		last = strings.Trim(last, "\"")
		return first + ".." + last + " (" + strconv.Itoa(len(matches)) + " versions)", nil
	}

	return strings.Trim(string(matches[0]), "\""), nil
}
