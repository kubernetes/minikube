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

type dependency struct {
	filePath      string
	versionRegexp string
}

var dependencies = map[string]dependency{
	"buildkit":       {"deploy/iso/minikube-iso/arch/x86_64/package/buildkit-bin/buildkit-bin.mk", `BUILDKIT_BIN_VERSION = (.*)`},
	"cloud-spanner":  {"pkg/minikube/assets/addons.go", `cloud-spanner-emulator/emulator:(.*)@`},
	"containerd":     {"deploy/iso/minikube-iso/arch/x86_64/package/containerd-bin/containerd-bin.mk", `CONTAINERD_BIN_VERSION = (.*)`},
	"cri-o":          {"deploy/iso/minikube-iso/package/crio-bin/crio-bin.mk", `CRIO_BIN_VERSION = (.*)`},
	"gh":             {"hack/jenkins/installers/check_install_gh.sh", `GH_VERSION="(.*)"`},
	"go":             {"Makefile", `GO_VERSION \?= (.*)`},
	"golint":         {"Makefile", `GOLINT_VERSION \?= (.*)`},
	"gopogh":         {"hack/jenkins/common.sh", `github.com/medyagh/gopogh/cmd/gopogh@(.*)`},
	"gotestsum":      {"hack/jenkins/installers/check_install_gotestsum.sh", `gotest\.tools/gotestsum@(.*)`},
	"hugo":           {"netlify.toml", `HUGO_VERSION = "(.*)"`},
	"metrics-server": {"pkg/minikube/assets/addons.go", `metrics-server/metrics-server:(.*)@`},
	"runc":           {"deploy/iso/minikube-iso/package/runc-master/runc-master.mk", `RUNC_MASTER_VERSION = (.*)`},
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
