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
	"context"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute
)

var (
	schema = map[string]update.Item{
		"hack/jenkins/common.ps1": {
			Replace: map[string]string{
				`gotest.tools/gotestsum@.*`: `gotest.tools/gotestsum@{{.StableVersion}}`,
			},
		},
		"hack/jenkins/installers/check_install_gotestsum.sh": {
			Replace: map[string]string{
				`gotest.tools/gotestsum@.*`: `gotest.tools/gotestsum@{{.StableVersion}}`,
			},
		},
	}
)

// Data holds stable gotestsum version in semver format.
type Data struct {
	StableVersion string
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get gotestsum stable version from https://github.com/gotestyourself/gotestsum
	stable, err := update.StableVersion(ctx, "gotestyourself", "gotestsum")
	if err != nil || stable == "" {
		klog.Fatalf("Unable to get gotestsum stable version: %v", err)
	}
	data := Data{StableVersion: stable}
	klog.Infof("gotestsum stable version: %s", data.StableVersion)

	update.Apply(schema, data)
}
