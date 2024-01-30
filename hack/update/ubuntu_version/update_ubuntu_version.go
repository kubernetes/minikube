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
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var (
	schema = map[string]update.Item{
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`UBUNTU_JAMMY_IMAGE=.*`: `UBUNTU_JAMMY_IMAGE="{{.LatestVersion}}"`,
			},
		},
	}
)

// Data holds latest Ubuntu jammy version in semver format.
type Data struct {
	LatestVersion string
}

func latestJammyTag(tags []string) (string, error) {
	for _, tag := range tags {
		if strings.Contains(tag, "jammy-") {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tag found that matches: jammy-")
}

func main() {
	tags, err := update.ImageTagsFromDockerHub("library/ubuntu")
	if err != nil {
		klog.Fatal(err)
	}
	jammyTag, err := latestJammyTag(tags)
	if err != nil {
		klog.Fatal(err)
	}
	data := Data{LatestVersion: fmt.Sprintf("ubuntu:%s", jammyTag)}

	update.Apply(schema, data)
}
