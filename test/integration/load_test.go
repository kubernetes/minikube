// +build integration

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

package integration

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestImageLoad(t *testing.T) {
	if NoneDriver() {
		t.Skip("skipping on none driver")
	}
	profile := UniqueProfileName("load-image")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2000"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	// pull busybox
	busybox := "busybox:latest"
	rr, err = Run(t, exec.CommandContext(ctx, "docker", "pull", busybox))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	// tag busybox
	newImage := fmt.Sprintf("busybox:%s", profile)
	rr, err = Run(t, exec.CommandContext(ctx, "docker", "tag", busybox, newImage))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	// try to load the new image into minikube
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "load", newImage))
	if err != nil {
		t.Fatalf("loading image into minikube: %v\n%s", err, rr.Output())
	}

	// make sure the image was correctly loaded
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "--", "sudo", "ctr", "-n=k8s.io", "image", "ls"))
	if err != nil {
		t.Fatalf("listing images: %v\n%s", err, rr.Output())
	}
	if !strings.Contains(rr.Output(), newImage) {
		t.Fatalf("expected %s to be loaded into minikube but the image is not there", newImage)
	}
}
