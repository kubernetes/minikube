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

package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const dockerHubTagsURL = "https://hub.docker.com/v2/repositories/%s/tags?page_size=100"

// Response is used to unmarshal the response from Docker Hub
type Response struct {
	Results []struct {
		Name string `json:"name"`
	}
}

// ImageTagsFromDockerHub returns the 100 latest image tags from Docker Hub for the provided repository
func ImageTagsFromDockerHub(repo string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf(dockerHubTagsURL, repo))
	if err != nil {
		return nil, fmt.Errorf("unable to get tags list: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read HTTP response from Docker Hub: %v", err)
	}

	var content Response
	err = json.Unmarshal(body, &content)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal response from Docker Hub: %v", err)
	}

	versions := make([]string, len(content.Results))
	for i, tag := range content.Results {
		versions[i] = tag.Name
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("tag list is empty")
	}

	return versions, nil
}
