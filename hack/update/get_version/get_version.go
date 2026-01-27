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
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	addonsFile = "pkg/minikube/assets/addons.go"
	dockerfile = "deploy/kicbase/Dockerfile"
)

type dependency struct {
	filePath      string
	versionRegexp string
}

var dependencies = map[string]dependency{
	"amd-device-gpu-plugin":   {addonsFile, `rocm/k8s-device-plugin:(.*)@`},
	"buildkit":                {"deploy/iso/minikube-iso/arch/x86_64/package/buildkit-bin/buildkit-bin.mk", `BUILDKIT_BIN_VERSION = (.*)`},
	"calico":                  {"pkg/minikube/bootstrapper/images/images.go", `calicoVersion = "(.*)"`},
	"cilium":                  {"pkg/minikube/cni/cilium.yaml", `quay.io/cilium/cilium:(.*)@`},
	"cloud-spanner-emulator":  {addonsFile, `cloud-spanner-emulator/emulator:(.*)@`},
	"cni-plugins":             {"deploy/iso/minikube-iso/arch/x86_64/package/cni-plugins-latest/cni-plugins-latest.mk", `CNI_PLUGINS_LATEST_VERSION = (.*)`},
	"containerd":              {"deploy/iso/minikube-iso/arch/x86_64/package/containerd-bin/containerd-bin.mk", `CONTAINERD_BIN_VERSION = (.*)`},
	"cri-dockerd":             {dockerfile, `CRI_DOCKERD_VERSION="(.*)"`},
	"cri-o":                   {"deploy/iso/minikube-iso/package/crio-bin/crio-bin.mk", `CRIO_BIN_VERSION = (.*)`},
	"crictl":                  {"deploy/iso/minikube-iso/arch/x86_64/package/crictl-bin/crictl-bin.mk", `CRICTL_BIN_VERSION = (.*)`},
	"crun":                    {"deploy/iso/minikube-iso/package/crun-latest/crun-latest.mk", `CRUN_LATEST_VERSION = (.*)`},
	"docker":                  {"deploy/iso/minikube-iso/arch/x86_64/package/docker-bin/docker-bin.mk", `DOCKER_BIN_VERSION = (.*)`},
	"docker-buildx":           {"deploy/iso/minikube-iso/arch/x86_64/package/docker-buildx/docker-buildx.mk", `DOCKER_BUILDX_VERSION = (.*)`},
	"flannel":                 {"pkg/minikube/cni/flannel.yaml", `flannel:(.*)`},
	"gcp-auth":                {addonsFile, `k8s-minikube/gcp-auth-webhook:(.*)@`},
	"gh":                      {"hack/jenkins/installers/check_install_gh.sh", `GH_VERSION="(.*)"`},
	"golang":                  {"Makefile", `\nGO_VERSION \?= (.*)`},
	"go-github":               {"go.mod", `github\.com\/google\/go-github\/.* (.*)`},
	"golint":                  {"Makefile", `GOLINT_VERSION \?= (.*)`},
	"gopogh":                  {"hack/jenkins/installers/check_install_gopogh.sh", `github.com/medyagh/gopogh/cmd/gopogh@(.*)`},
	"gotestsum":               {"hack/jenkins/installers/check_install_gotestsum.sh", `gotest\.tools/gotestsum@(.*)`},
	"headlamp":                {addonsFile, `headlamp-k8s/headlamp:(.*)@`},
	"hugo":                    {"netlify.toml", `HUGO_VERSION = "(.*)"`},
	"ingress":                 {addonsFile, `ingress-nginx/controller:(.*)@`},
	"inspektor-gadget":        {addonsFile, `inspektor-gadget/inspektor-gadget:(.*)@`},
	"istio-operator":          {addonsFile, `istio/operator:(.*)@`},
	"kindnetd":                {"pkg/minikube/bootstrapper/images/images.go", `kindnetd:(.*)"`},
	"kong":                    {addonsFile, `kong:(.*)@`},
	"kong-ingress-controller": {addonsFile, `kong/kubernetes-ingress-controller:(.*)@`},
	"kube-registry-proxy":     {addonsFile, `"k8s-minikube/kube-registry-proxy:(.*)@`},
	"kube-vip":                {"pkg/minikube/cluster/ha/kube-vip/kube-vip.go", `image: ghcr.io/kube-vip/kube-vip:(.*)`},
	"kubectl":                 {addonsFile, `bitnami/kubectl:(.*)@`},
	"kubevirt":                {"deploy/addons/kubevirt/pod.yaml.tmpl", `KUBEVIRT_VERSION="(.*)"`},
	"metrics-server":          {addonsFile, `metrics-server/metrics-server:(.*)@`},
	"nerdctl":                 {"deploy/kicbase/Dockerfile", `NERDCTL_VERSION="(.*)"`},
	"nerdctld":                {"deploy/kicbase/Dockerfile", `NERDCTLD_VERSION="(.*)"`},
	"node":                    {"netlify.toml", `NODE_VERSION = "(.*)"`},
	"nvidia-device-plugin":    {addonsFile, `nvidia/k8s-device-plugin:(.*)@`},
	"portainer":               {addonsFile, `portainer/portainer-ce:(.*)@`},
	"registry":                {addonsFile, `registry:(.*)@`},
	"runc":                    {"deploy/iso/minikube-iso/package/runc-master/runc-master.mk", `RUNC_MASTER_VERSION = (.*)`},
	"debian":                  {dockerfile, `debian:bookworm-(.*)-slim`},
	"volcano":                 {addonsFile, `volcanosh/vc-webhook-manager:(.*)@`},
	"yakd":                    {addonsFile, `manusa/yakd:(.*)@`},
}

func main() {
	depName := os.Getenv("DEP")
	if depName == "" {
		log.Fatalf("the environment variable 'DEP' needs to be set")
	}
	depName = standrizeComponentName(depName)

	// Handle special cases
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
		// Use regular handling for site-node-version (same as "node" from netlify.toml)
		depName = "node"
	}

	dep, ok := dependencies[depName]
	if !ok {
		log.Fatalf("%s is not a valid dependency", depName)
	}
	re, err := regexp.Compile(dep.versionRegexp)
	if err != nil {
		log.Fatalf("regexp failed to compile: %v", err)
	}
	// because in the Makefile we run it as @(cd hack && go run update/get_version/get_version.go) we need ../
	data, err := os.ReadFile("../" + dep.filePath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// this handles cases where multiple versions exist (e.g., old and new versions in go.mod)
	allMatches := re.FindAllSubmatch(data, -1)
	if len(allMatches) == 0 {
		log.Fatalf("no matches found")
	}

	// Take the last match (most recent version)
	submatches := allMatches[len(allMatches)-1]
	if len(submatches) < 2 {
		log.Fatalf("less than 2 submatches found")
	}
	os.Stdout.Write(submatches[1])
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
