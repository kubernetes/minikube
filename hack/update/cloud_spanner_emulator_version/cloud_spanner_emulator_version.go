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
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute
)

var schema = map[string]update.Item{
	"pkg/minikube/assets/addons.go": {
		Replace: map[string]string{
			`cloud-spanner-emulator/emulator:.*`: `cloud-spanner-emulator/emulator:{{.Version}}@{{.SHA}}",`,
		},
	},
}

// Data holds stable cloud-spanner-emulator version in semver format.
type Data struct {
	Version string
	SHA     string
}

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get cloud-spanner-emulator stable version
	stable, err := update.StableVersion(ctx, "GoogleCloudPlatform", "cloud-spanner-emulator")
	if err != nil {
		klog.Fatalf("Unable to get cloud-spanner-emulator stable version: %v", err)
	}
	stable = strings.TrimPrefix(stable, "v")
	sha, err := update.GetImageSHA(fmt.Sprintf("gcr.io/cloud-spanner-emulator/emulator:%s", stable))
	if err != nil {
		klog.Fatalf("failed to get image SHA: %v", err)
	}

	data := Data{Version: stable, SHA: sha}
	klog.Infof("cloud-spanner-emulator stable version: %s", data.Version)

	update.Apply(schema, data)
}
