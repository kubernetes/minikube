/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

func main() {
	if _, err := exec.LookPath("helm"); err != nil {
		klog.Fatal("helm not found on system, either install or add to PATH")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "cilium", "cilium")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	version := strings.TrimPrefix(stable.Tag, "v")

	// Add the cilium repo to helm
	if err := exec.Command("helm", "repo", "add", "cilium", "https://helm.cilium.io/").Run(); err != nil {
		klog.Fatal(err)
	}

	// Generate the cilium YAML
	yamlBytes, err := exec.Command("helm", "template", "cilium", "cilium/cilium", "--version", version, "--namespace", "kube-system").Output()
	if err != nil {
		klog.Fatal(err)
	}
	yaml := string(yamlBytes)

	// Remove the cilium/templates/cilium-ca-secret.yaml section
	re := regexp.MustCompile(`# Source: cilium\/templates\/cilium-ca-secret\.yaml(\n.*?)+---\n`)
	yaml = re.ReplaceAllString(yaml, "")

	// Remove the cilium/templates/hubble/tls-helm/server-secret.yaml section
	re = regexp.MustCompile(`# Source: cilium\/templates\/hubble\/tls-helm\/server-secret\.yaml(\n.*?)+---\n`)
	yaml = re.ReplaceAllString(yaml, "")

	// Replace `cluster-pool-ipv4-cidr` with PodSubnet template
	re = regexp.MustCompile(`10\.0\.0\.0\/8`)
	yaml = re.ReplaceAllString(yaml, "{{ .PodSubnet }}")

	// Change replicas to 1
	re = regexp.MustCompile(`replicas:.+`)
	yaml = re.ReplaceAllString(yaml, "replicas: 1")

	filename := filepath.Join(update.FSRoot, "pkg/minikube/cni/cilium.yaml")
	if err := os.WriteFile(filename, []byte(yaml), 0644); err != nil {
		klog.Fatal(err)
	}
}
