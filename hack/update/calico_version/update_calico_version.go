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
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

var schema = map[string]update.Item{
	"pkg/minikube/bootstrapper/images/images.go": {
		Replace: map[string]string{
			`calicoVersion = .*`: `calicoVersion = "{{.Version}}"`,
		},
	},
}

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "projectcalico", "calico")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	data := Data{Version: stable.Tag}

	update.Apply(schema, data)

	updateYAML(stable.Tag)
}

func updateYAML(version string) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/calico.yaml", version))
	if err != nil {
		klog.Fatalf("failed to get calico.yaml: %v", err)
	}
	defer res.Body.Close()
	yaml, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	replacements := map[string]string{
		`policy\/v1`:                              "policy/v1{{if .LegacyPodDisruptionBudget}}beta1{{end}}",
		`docker\.io\/calico\/cni:.*`:              "{{ .BinaryImageName }}",
		`docker\.io\/calico\/node:.*`:             "{{ .DaemonSetImageName }}",
		`docker\.io\/calico\/kube-controllers:.*`: "{{ .DeploymentImageName }}",
	}
	for re, repl := range replacements {
		yaml = regexp.MustCompile(re).ReplaceAll(yaml, []byte(repl))
	}
	if err := os.WriteFile("../../../pkg/minikube/cni/calico.yaml", yaml, 0644); err != nil {
		klog.Fatalf("failed to write to YAML file: %v", err)
	}
}
