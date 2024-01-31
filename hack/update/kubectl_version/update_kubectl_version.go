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
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`"bitnami/kubectl:.*`: `"bitnami/kubectl:{{.Version}}@{{.SHA}}",`,
		},
	},
}

type Data struct {
	Version string
	SHA     string
}

func main() {
	tags, err := update.ImageTagsFromDockerHub("bitnami/kubectl")
	if err != nil {
		klog.Fatal(err)
	}
	tag := latestStableSemverTag(tags)
	sha, err := update.GetImageSHA(fmt.Sprintf("docker.io/bitnami/kubectl:%s", tag))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: tag, SHA: sha}

	update.Apply(schema, data)
}

func latestStableSemverTag(tags []string) string {
	latest := "v0.0.0"
	for _, tag := range tags {
		vTag := fmt.Sprintf("v%s", tag)
		if semver.IsValid(vTag) && semver.Prerelease(vTag) == "" && semver.Compare(vTag, latest) == 1 {
			latest = vTag
		}
	}
	return strings.TrimPrefix(latest, "v")
}
