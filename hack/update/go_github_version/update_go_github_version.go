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

	update.Apply(generateSchema(), data)

	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		klog.Fatalf("failed to run go mod tidy: %v", err)
	}
}

func generateSchema() map[string]update.Item {
	files := []string{
		"cmd/minikube/cmd/config/kubernetes_version.go",
		"hack/preload-images/kubernetes.go",
		"hack/update/github.go",
		"hack/update/ingress_version/update_ingress_version.go",
		"hack/update/kubeadm_constants/update_kubeadm_constants.go",
		"hack/update/kubernetes_versions_list/update_kubernetes_versions_list.go",
		"hack/update/site_node_version/update_site_node_version.go",
		"pkg/perf/monitor/github.go",
	}

	schema := make(map[string]update.Item)

	for _, f := range files {
		schema[f] = replace
	}

	return schema
}
