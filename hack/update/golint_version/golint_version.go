/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
		"Makefile": {
			Replace: map[string]string{
				`GOLINT_VERSION \?= v1.*`: `GOLINT_VERSION ?= {{.StableVersion}}`,
			},
		},
	}
)

// Data holds stable golint version in semver format.
type Data struct {
	StableVersion string
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Golint stable version
	stable, err := update.StableVersion(ctx, "golangci", "golangci-lint")
	if err != nil {
		klog.Fatalf("Unable to get Golint stable version: %v", err)
	}
	data := Data{StableVersion: stable}
	klog.Infof("Golint stable version: %s", data.StableVersion)

	update.Apply(schema, data)
}
