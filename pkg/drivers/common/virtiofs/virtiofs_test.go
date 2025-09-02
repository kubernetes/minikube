/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package virtiofs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"k8s.io/minikube/pkg/drivers/common/virtiofs"
)

func TestVirtiofsValidateEmptyMountString(t *testing.T) {
	mounts, err := virtiofs.ValidateMountString("")
	if err != nil {
		t.Fatalf("failed to parse empty mount string: %s", err)
	}
	if mounts != nil {
		t.Fatalf("expected nil mounts, got %v", mounts)
	}
}

func TestVirtiofsValidateMountString(t *testing.T) {
	hostPath := t.TempDir()
	guestPath := "/mnt/models"
	mountString := fmt.Sprintf("%s:%s", hostPath, guestPath)

	mounts, err := virtiofs.ValidateMountString(mountString)
	if err != nil {
		t.Fatalf("failed to parse mountString %q: %s", mountString, err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected a single mount, got %v", mounts)
	}

	mount := mounts[0]
	if mount.HostPath != hostPath {
		t.Fatalf("expected host path %q, got %q", hostPath, mount.HostPath)
	}
	if mount.GuestPath != guestPath {
		t.Fatalf("expected guest path %q, got %q", guestPath, mount.GuestPath)
	}

	tag, err := uuid.Parse(mount.Tag)
	if err != nil {
		t.Fatalf("failed to parse UUID from mount tag: %s", err)
	}
	if tag.Version() != 4 {
		t.Fatalf("mount tag is not a random UUID")
	}

	if err := mount.Validate(); err != nil {
		t.Fatalf("mount is not valid: %s", err)
	}
}

func TestVirtiofsParseInvalidMountString(t *testing.T) {
	for _, tt := range []struct {
		name        string
		mountString string
	}{
		{
			name:        "empty",
			mountString: "",
		},
		{
			name:        "guest path is missing",
			mountString: "host-path",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mount, err := virtiofs.ParseMount(tt.mountString)
			if err == nil {
				t.Fatalf("invalid mount string %q did not fail to parse", tt.mountString)
			}
			if mount != nil {
				t.Fatalf("expected nil mount for %q, got %v", tt.mountString, mount)
			}
		})
	}
}

func TestVirtiofsValidateInvalidMount(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")
	file := filepath.Join(dir, "file")

	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	for _, tt := range []struct {
		name        string
		mountString string
	}{
		{
			name:        "host path contains virtiofs config separator",
			mountString: "/host,path:/guest-path",
		},
		{
			name:        "host path is relative",
			mountString: "host-path:/guest-path",
		},
		{
			name:        "guest path is relative",
			mountString: fmt.Sprintf("%s:guest-path", dir),
		},
		{
			name:        "host path is missing",
			mountString: fmt.Sprintf("%s:/guest-path", missing),
		},
		{
			name:        "host path is not a directory",
			mountString: fmt.Sprintf("%s:/guest-path", file),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mount, err := virtiofs.ParseMount(tt.mountString)
			if err != nil {
				t.Fatalf("failed to parse mount string %q: %s", tt.mountString, err)
			}
			if err := mount.Validate(); err == nil {
				t.Fatalf("invalid mount %q did not failed validation", tt.mountString)
			}
		})
	}
}
