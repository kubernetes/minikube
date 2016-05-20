/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"encoding/json"
	"fmt"
	"k8s.io/minikube/pkg/minikube/constants"
	"net/http"
)

type version struct {
	ISOURL string
}

type isoMap map[string]version

func GetIsoUrl(mapURL, v string) (string, error) {
	// Download the version map file.
	resp, err := http.Get(mapURL)
	if err != nil {
		return "", err
	}
	dec := json.NewDecoder(resp.Body)
	var m isoMap
	if err := dec.Decode(&m); err != nil {
		return "", err
	}
	version, ok := m[v]
	if !ok {
		return "", fmt.Errorf("Version %s not found in: %s", constants.ISOVersion, m)
	}
	return version.ISOURL, nil
}
