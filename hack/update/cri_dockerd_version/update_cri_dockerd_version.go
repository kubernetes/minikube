/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"log"
	"os"
	"os/exec"

	"k8s.io/minikube/hack/update"
)

var (
	schema = map[string]update.Item{
		".github/workflows/master.yml": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="{{.FullCommit}}"`,
			},
		},
		".github/workflows/pr.yml": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="{{.FullCommit}}"`,
			},
		},
		"hack/jenkins/linux_integration_tests_none.sh": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION=".*"`: `CRI_DOCKERD_VERSION="{{.FullCommit}}"`,
			},
		},
		"deploy/iso/minikube-iso/arch/aarch64/package/cri-dockerd-aarch64/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_AARCH64_VERSION = .*`: `CRI_DOCKERD_AARCH64_VERSION = {{.FullCommit}}`,
				`CRI_DOCKERD_AARCH64_REV = .*`:     `CRI_DOCKERD_AARCH64_REV = {{.ShortCommit}}`,
			},
		},
		"deploy/iso/minikube-iso/arch/x86_64/package/cri-dockerd/cri-dockerd.mk": {
			Replace: map[string]string{
				`CRI_DOCKERD_VERSION = .*`: `CRI_DOCKERD_VERSION = {{.FullCommit}}`,
				`CRI_DOCKERD_REV = .*`:     `CRI_DOCKERD_REV = {{.ShortCommit}}`,
			},
		},
	}
)

// Data holds stable cri-dockerd version in semver format.
type Data struct {
	FullCommit  string
	ShortCommit string
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: update_cri_dockerd_version.go <version> <archlist>")
	}

	commit := os.Args[1]
	archs := os.Args[2]

	data := Data{FullCommit: commit, ShortCommit: commit[:7]}

	update.Apply(schema, data)

	if out, err := exec.Command("./update_cri_dockerd_version.sh", commit, archs).CombinedOutput(); err != nil {
		log.Fatalf("failed to build and upload cri-dockerd binaries: %s", string(out))
	}

	fmt.Println("Don't forget you still need to update the hash files!")
}
