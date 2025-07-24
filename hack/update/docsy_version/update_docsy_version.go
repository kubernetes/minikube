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
	"fmt"
	"os/exec"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

const (
	// default context timeout
	cxTimeout = 5 * time.Minute
)

func main() {
	// set a context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), cxTimeout)
	defer cancel()

	// get Docsy stable version
	stable, err := update.StableVersion(ctx, "google", "docsy")
	if err != nil {
		klog.Fatalf("Unable to get Docsy stable version: %v", err)
	}

	if err := exec.CommandContext(ctx, "./update_docsy_version.sh", stable).Run(); err != nil {
		klog.Fatalf("failed to update docsy commit: %v", err)
	}

	fmt.Print(stable)
}
