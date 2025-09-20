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
				`DEBIAN_BOOKWORM_IMAGE=.*`: `DEBIAN_BOOKWORM_IMAGE="{{.LatestVersion}}"`,
			},
		},
	}
)

// Data holds the latest Debian bookworm slim tag.
type Data struct {
	LatestVersion string
}

func latestBookwormSlimTag(tags []string) (string, error) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "bookworm-") && strings.HasSuffix(tag, "-slim") {
			return tag, nil
		}
	}
	for _, tag := range tags {
		if tag == "bookworm-slim" {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no tag found that matches: bookworm-*-slim")
}

func main() {
	tags, err := update.ImageTagsFromDockerHub("library/debian")
	if err != nil {
		klog.Fatal(err)
	}
	bookwormTag, err := latestBookwormSlimTag(tags)
	if err != nil {
		klog.Fatal(err)
	}
	data := Data{LatestVersion: fmt.Sprintf("debian:%s", bookwormTag)}

	update.Apply(schema, data)
}
