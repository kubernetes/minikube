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

const cxTimeout = 1 * time.Minute

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`inspektor-gadget/inspektor-gadget:.*`: `inspektor-gadget/inspektor-gadget:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	stable, err := update.StableVersion(ctx, "inspektor-gadget", "inspektor-gadget")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	sha, err := update.GetImageSHA(fmt.Sprintf("ghcr.io/inspektor-gadget/inspektor-gadget:%s", stable))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: stable, SHA: sha}
	klog.Infof("inspektor-gadget stable version: %s", data.Version)

	update.Apply(schema, data)
	updateDeploymentYAML(stable)
	updateCRDYAML(stable)
}

func updateDeploymentYAML(version string) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/inspektor-gadget/inspektor-gadget/refs/tags/%s/pkg/resources/manifests/deploy.yaml", version))
	if err != nil {
		klog.Fatalf("failed to get yaml file: %v", err)
	}
	defer res.Body.Close()
	yaml, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	replacements := map[string]string{
		`ghcr\.io\/inspektor-gadget\/inspektor-gadget:.*`: "{{.CustomRegistries.InspektorGadget  | default .ImageRepository | default .Registries.InspektorGadget }}{{.Images.InspektorGadget}}",
	}
	for re, repl := range replacements {
		yaml = regexp.MustCompile(re).ReplaceAll(yaml, []byte(repl))
	}
	if err := os.WriteFile("../../../deploy/addons/inspektor-gadget/ig-deployment.yaml.tmpl", yaml, 0644); err != nil {
		klog.Fatalf("failed to write to YAML file: %v", err)
	}
}

func updateCRDYAML(version string) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/inspektor-gadget/inspektor-gadget/refs/tags/%s/pkg/resources/crd/bases/gadget.kinvolk.io_traces.yaml", version))
	if err != nil {
		klog.Fatalf("failed to get yaml file: %v", err)
	}
	defer res.Body.Close()
	yaml, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	if err := os.WriteFile("../../../deploy/addons/inspektor-gadget/ig-crd.yaml", yaml, 0644); err != nil {
		klog.Fatalf("failed to write to YAML file: %v", err)
	}
}
