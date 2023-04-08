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

/*
Copyright 2022 The Kubernetes Authors All rights reserved.
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

const (
	dockerHubUbuntuBaseURL = "https://hub.docker.com/v2/repositories/library/ubuntu/tags"
)

var (
	schema = map[string]update.Item{
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`UBUNTU_FOCAL_IMAGE=.*`: `UBUNTU_FOCAL_IMAGE="{{.LatestVersion}}"`,
			},
		},
	}
)

// Data holds latest Ubuntu focal version in semver format.
type Data struct {
	LatestVersion string
}

// Response is used to unmarshal the response from Docker Hub
type Response struct {
	Results []struct {
		Name string `json:"name"`
	}
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(dockerHubUbuntuBaseURL)
	if err != nil {
		return "", fmt.Errorf("unable to get Ubuntu focal's latest version: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read HTTP response from Docker Hub: %v", err)
	}

	var content Response
	err = json.Unmarshal(body, &content)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshall response from Docker Hub: %v", err)
	}

	for _, i := range content.Results {
		if strings.Contains(i.Name, "focal-") {
			return i.Name, nil
		}
	}

	return "", fmt.Errorf("response from Docker Hub does not contain a latest focal image")
}

func main() {
	// get Ubuntu Focal latest version
	latest, err := getLatestVersion()
	if err != nil {
		klog.Fatalf("Unable to find latest ubuntu:focal version: %v\n", err)
	}
	data := Data{LatestVersion: fmt.Sprintf("ubuntu:%s", latest)}
	klog.Infof("Ubuntu focal latest version: %s", latest)

	update.Apply(schema, data)
}
