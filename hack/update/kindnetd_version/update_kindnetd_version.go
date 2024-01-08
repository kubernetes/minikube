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
	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
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

func main() {
	tags, err := update.ImageTagsFromDockerHub("kindest/kindnetd")
	if err != nil {
		klog.Fatal(err)
	}
	data := Data{LatestVersion: tags[0]}

	update.Apply(schema, data)
}
