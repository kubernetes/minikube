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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "flannel-io", "flannel")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}

	updateYAML(stable.Tag)
}

func updateYAML(version string) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/flannel-io/flannel/%s/Documentation/kube-flannel.yml", version))
	if err != nil {
		klog.Fatalf("failed to get kube-flannel.yaml: %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		klog.Fatalf("failed to read body: %v", err)
	}
	yaml := regexp.MustCompile(`10\.244\.0\.0\/16`).ReplaceAll(body, []byte("{{ .PodCIDR }}"))
	if err := os.WriteFile("../../../pkg/minikube/cni/flannel.yaml", yaml, 0644); err != nil {
		klog.Fatalf("failed to write to YAML file: %v", err)
	}
}
