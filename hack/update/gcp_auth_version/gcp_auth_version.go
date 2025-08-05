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
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`k8s-minikube/gcp-auth-webhook:.*`: `k8s-minikube/gcp-auth-webhook:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stable, err := update.StableVersion(ctx, "GoogleContainerTools", "gcp-auth-webhook")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	sha, err := update.GetImageSHA(fmt.Sprintf("gcr.io/k8s-minikube/gcp-auth-webhook:%s", stable))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: stable, SHA: sha}
	klog.Infof("gcp-auth stable version: %s", data.Version)

	update.Apply(schema, data)
}
