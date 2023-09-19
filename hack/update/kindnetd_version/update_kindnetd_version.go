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

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	dockerHubKindnetdTags = "https://hub.docker.com/v2/repositories/kindest/kindnetd/tags"
)

var (
	schema = map[string]update.Item{
		"pkg/minikube/bootstrapper/images/images.go": {
			Replace: map[string]string{
				`kindnetd:.*"`: `kindnetd:{{.LatestVersion}}"`,
			},
		},
	}
)

type Data struct {
	LatestVersion string
}

type Response struct {
	Results []struct {
		Name string `json:"name"`
	}
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(dockerHubKindnetdTags)
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read HTTP response: %v", err)
	}

	var content Response
	err = json.Unmarshal(body, &content)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(content.Results) == 0 {
		return "", fmt.Errorf("tag list is empty")
	}

	return content.Results[0].Name, nil
}

func main() {
	latest, err := getLatestVersion()
	if err != nil {
		klog.Fatalf("failed to get latest version: %v", err)
	}
	data := Data{LatestVersion: latest}

	update.Apply(schema, data)
}
