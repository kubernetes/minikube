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
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`vc-webhook-manager:.*`:    `vc-webhook-manager:{{.Version}}@{{.SHAWebhookManager}}",`,
			`vc-controller-manager:.*`: `vc-controller-manager:{{.Version}}@{{.SHAControllerManager}}",`,
			`vc-scheduler:.*`:          `vc-scheduler:{{.Version}}@{{.SHAScheduler}}",`,
		},
	},
}

type Data struct {
	Version              string
	SHAWebhookManager    string
	SHAControllerManager string
	SHAScheduler         string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "volcano-sh", "volcano")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	version := stable.Tag
	shaWebhookManager, err := update.GetImageSHA(fmt.Sprintf("docker.io/volcanosh/vc-webhook-manager:%s", version))
	if err != nil {
		klog.Fatalf("failed to get manifest digest for docker.io/volcanosh/vc-webhook-manager: %v", err)
	}

	shaControllerManager, err := update.GetImageSHA(fmt.Sprintf("docker.io/volcanosh/vc-controller-manager:%s", version))
	if err != nil {
		klog.Fatalf("failed to get manifest digest for docker.io/volcanosh/vc-controller-manager: %v", err)
	}

	shaScheduler, err := update.GetImageSHA(fmt.Sprintf("docker.io/volcanosh/vc-scheduler:%s", version))
	if err != nil {
		klog.Fatalf("failed to get manifest digest for docker.io/volcanosh/vc-scheduler: %v", err)
	}

	data := Data{
		Version:              version,
		SHAWebhookManager:    shaWebhookManager,
		SHAControllerManager: shaControllerManager,
		SHAScheduler:         shaScheduler,
	}

	update.Apply(schema, data)

	updateYAML(version)
}

func updateYAML(version string) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/volcano-sh/volcano/%s/installer/volcano-development.yaml", version))
	if err != nil {
		klog.Fatalf("failed to get yaml file: %v", err)
	}
	defer res.Body.Close()
	yaml, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	replacements := map[string]string{
		`docker\.io\/volcanosh\/vc-webhook-manager:.*`:    "{{.CustomRegistries.vc_webhook_manager | default .ImageRepository | default .Registries.vc_webhook_manager}}{{.Images.vc_webhook_manager}}",
		`docker\.io\/volcanosh\/vc-controller-manager:.*`: "{{.CustomRegistries.vc_controller_manager | default .ImageRepository | default .Registries.vc_controller_manager}}{{.Images.vc_controller_manager}}",
		`docker\.io\/volcanosh\/vc-scheduler:.*`:          "{{.CustomRegistries.vc_scheduler | default .ImageRepository | default .Registries.vc_scheduler}}{{.Images.vc_scheduler}}",
	}
	for re, repl := range replacements {
		yaml = regexp.MustCompile(re).ReplaceAll(yaml, []byte(repl))
	}
	if err := os.WriteFile("../../../deploy/addons/volcano/volcano-development.yaml.tmpl", yaml, 0644); err != nil {
		klog.Fatalf("failed to write to YAML file: %v", err)
	}
}
