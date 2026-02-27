/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	cxTimeout = 5 * time.Minute
	// KubeVirt publishes stable version at this URL
	stableVersionURL = "https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt"
)

var schema = map[string]update.Item{
	"deploy/addons/kubevirt/pod.yaml.tmpl": {
		Replace: map[string]string{
			`KUBEVIRT_VERSION="v.*"`: `KUBEVIRT_VERSION="{{.Version}}"`,
		},
	},
}

// Data holds the version information for templating
type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	stable, err := getStableKubeVirtVersion(ctx)
	if err != nil {
		klog.Fatalf("Unable to get KubeVirt stable version: %v", err)
	}

	data := Data{Version: stable}
	klog.Infof("KubeVirt stable version: %s", data.Version)

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
}

// getStableKubeVirtVersion fetches the stable KubeVirt version from the official release endpoint
func getStableKubeVirtVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stableVersionURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
