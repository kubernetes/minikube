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
	"context"
	"os/exec"
	"time"

	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var (
	replace = update.Item{
		Replace: map[string]string{
			`github\.com\/google\/go-github\/v.*`: `github.com/google/go-github/{{.Version}}/github"`,
		},
	}
)

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, err := update.StableVersion(ctx, "google", "go-github")
	if err != nil {
		klog.Fatalf("Unable to get releases: %v", err)
	}

	major := semver.Major(stable)

	data := Data{Version: major}

	if err := update.Apply(generateSchema(), data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}

	// run go mod tidy in the root folder
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		klog.Fatalf("failed to run go mod tidy in root: %v", err)
	}

	// run go mod tidy in hack folder
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		klog.Fatalf("failed to run go mod tidy in hack: %v", err)
	}

}

func generateSchema() map[string]update.Item {
	files := []string{
		"cmd/minikube/cmd/config/kubernetes_version.go",
		"pkg/perf/monitor/github.go",
		"pkg/minikube/download/gh/gh.go",
		"pkg/minikube/download/gh/gh_test.go",
		"hack/preload-images/kubernetes.go",
		"hack/update/kubeadm_constants/kubeadm_constants.go",
		"hack/update/ingress_version/ingress_version.go",
		"hack/update/site_node_version/site_node_version.go",
		"hack/update/go_github_version/go_github_version.go",
		"hack/update/kubernetes_versions_list/kubernetes_versions_list.go",
		"hack/update/github.go",
		"hack/changelog/changelog.go",
		"hack/changelog/changelog_test.go",
	}

	schema := make(map[string]update.Item)

	for _, f := range files {
		schema[f] = replace
	}

	return schema
}
