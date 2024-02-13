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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

const dockerHubRegistryURL = "https://hub.docker.com/v2/repositories/library/registry/tags"

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`"registry:.*`: `"registry:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

// Response is used to unmarshal the response from Docker Hub
type Response struct {
	Results []struct {
		Name string `json:"name"`
	}
}

func main() {
	version, err := getLatestVersion()
	if err != nil {
		klog.Fatalf("failed to get latest version: %v", err)
	}
	version = strings.TrimPrefix(version, "v")
	sha, err := update.GetImageSHA(fmt.Sprintf("docker.io/registry:%s", version))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: version, SHA: sha}

	update.Apply(schema, data)
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(dockerHubRegistryURL)
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var content Response
	err = json.Unmarshal(body, &content)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	for _, i := range content.Results {
		if !strings.Contains(i.Name, "latest") {
			return i.Name, nil
		}
	}

	return "", fmt.Errorf("didn't find a non-latest image")
}
