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
	"regexp"
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
	"buildkit":                {"deploy/iso/minikube-iso/arch/x86_64/package/buildkit-bin/buildkit-bin.mk", `BUILDKIT_BIN_VERSION = (.*)`},
	"calico":                  {"pkg/minikube/bootstrapper/images/images.go", `calicoVersion = "(.*)"`},
	"cloud-spanner":           {addonsFile, `cloud-spanner-emulator/emulator:(.*)@`},
	"cni-plugins":             {"deploy/iso/minikube-iso/arch/x86_64/package/cni-plugins-latest/cni-plugins-latest.mk", `CNI_PLUGINS_VERSION = (.*)`},
	"containerd":              {"deploy/iso/minikube-iso/arch/x86_64/package/containerd-bin/containerd-bin.mk", `CONTAINERD_BIN_VERSION = (.*)`},
	"cri-dockerd":             {dockerfile, `CRI_DOCKERD_VERSION="(.*)"`},
	"cri-o":                   {"deploy/iso/minikube-iso/package/crio-bin/crio-bin.mk", `CRIO_BIN_VERSION = (.*)`},
	"crictl":                  {"deploy/iso/minikube-iso/arch/x86_64/package/crictl-bin/crictl-bin.mk", `CRICTL_BIN_VERSION = (.*)`},
	"docker":                  {"deploy/iso/minikube-iso/arch/x86_64/package/docker-bin/docker-bin.mk", `DOCKER_BIN_VERSION = (.*)`},
	"flannel":                 {"pkg/minikube/cni/flannel.yaml", `flannel:(.*)`},
	"gcp-auth":                {addonsFile, `k8s-minikube/gcp-auth-webhook:(.*)@`},
	"gh":                      {"hack/jenkins/installers/check_install_gh.sh", `GH_VERSION="(.*)"`},
	"go":                      {"Makefile", `\nGO_VERSION \?= (.*)`},
	"go-github":               {"go.mod", `github\.com\/google\/go-github\/.* (.*)`},
	"golint":                  {"Makefile", `GOLINT_VERSION \?= (.*)`},
	"gopogh":                  {"hack/jenkins/installers/check_install_gopogh.sh", `github.com/medyagh/gopogh/cmd/gopogh@(.*)`},
	"gotestsum":               {"hack/jenkins/installers/check_install_gotestsum.sh", `gotest\.tools/gotestsum@(.*)`},
	"hugo":                    {"netlify.toml", `HUGO_VERSION = "(.*)"`},
	"ingress":                 {addonsFile, `ingress-nginx/controller:(.*)@`},
	"inspektor-gadget":        {addonsFile, `inspektor-gadget/inspektor-gadget:(.*)@`},
	"istio-operator":          {addonsFile, `istio/operator:(.*)@`},
	"kindnetd":                {"pkg/minikube/bootstrapper/images/images.go", `kindnetd:(.*)"`},
	"kong":                    {addonsFile, `kong:(.*)@`},
	"kong-ingress-controller": {addonsFile, `kong/kubernetes-ingress-controller:(.*)@`},
	"kubectl":                 {addonsFile, `bitnami/kubectl:(.*)@`},
	"metrics-server":          {addonsFile, `metrics-server/metrics-server:(.*)@`},
	"nerdctl":                 {"deploy/kicbase/Dockerfile", `NERDCTL_VERSION="(.*)"`},
	"nerdctld":                {"deploy/kicbase/Dockerfile", `NERDCTLD_VERSION="(.*)"`},
	"node":                    {"netlify.toml", `NODE_VERSION = "(.*)"`},
	"nvidia-device-plugin":    {addonsFile, `nvidia/k8s-device-plugin:(.*)@`},
	"registry":                {addonsFile, `registry:(.*)@`},
	"runc":                    {"deploy/iso/minikube-iso/package/runc-master/runc-master.mk", `RUNC_MASTER_VERSION = (.*)`},
	"ubuntu":                  {dockerfile, `ubuntu:jammy-(.*)"`},
}

func main() {
	depName := os.Getenv("DEP")
	if depName == "" {
		log.Fatalf("the environment variable 'DEP' needs to be set")
	}
	dep, ok := dependencies[depName]
	if !ok {
		log.Fatalf("%s is not a valid dependency", depName)
	}
	re, err := regexp.Compile(dep.versionRegexp)
	if err != nil {
		log.Fatalf("regexp failed to compile: %v", err)
	}
	data, err := os.ReadFile("../../../" + dep.filePath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}
	submatches := re.FindSubmatch(data)
	if len(submatches) < 2 {
		log.Fatalf("less than 2 submatches found")
	}
	os.Stdout.Write(submatches[1])
}
