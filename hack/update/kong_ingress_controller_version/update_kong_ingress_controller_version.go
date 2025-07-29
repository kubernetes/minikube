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
	"strings"
	"time"

	"k8s.io/minikube/hack/update"

	"k8s.io/klog/v2"
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`kong/kubernetes-ingress-controller:.*`: `kong/kubernetes-ingress-controller:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	stable, _, _, err := update.GHReleases(ctx, "Kong", "kubernetes-ingress-controller")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	version := strings.TrimPrefix(stable.Tag, "v")
	sha, err := update.GetImageSHA(fmt.Sprintf("docker.io/kong/kubernetes-ingress-controller:%s", version))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: version, SHA: sha}

	update.Apply(schema, data)
}
