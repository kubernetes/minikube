//go:build integration

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

package findmnt_test

import (
	"os"
	"testing"

	"k8s.io/minikube/test/integration/findmnt"
)

func TestParseAllMounts(t *testing.T) {
	// Output from `findmnt --json` command.
	output, err := os.ReadFile("testdata/findmnt.json")
	if err != nil {
		t.Fatal(err)
	}
	result, err := findmnt.ParseOutput(output)
	if err != nil {
		t.Fatal(err)
	}

	assert(t, len(result.Filesystems), 1)

	root := result.Filesystems[0]
	assert(t, root.Target, "/")
	assert(t, root.Source, "tmpfs")
	assert(t, root.FSType, "tmpfs")
	assert(t, root.Options, "rw,relatime,size=5469192k")
	assert(t, len(root.Children), 20)

	dev := root.Children[0]
	assert(t, dev.Target, "/dev")
	assert(t, dev.Source, "devtmpfs")
	assert(t, dev.FSType, "devtmpfs")
	assert(t, dev.Options, "rw,relatime,size=2838500k,nr_inodes=709625,mode=755")
	assert(t, len(dev.Children), 4)

	devShm := dev.Children[0]
	assert(t, devShm.Target, "/dev/shm")
	assert(t, devShm.Source, "tmpfs")
	assert(t, devShm.FSType, "tmpfs")
	assert(t, devShm.Options, "rw,nosuid,nodev")
}

func TestParseSingle(t *testing.T) {
	// Output from `findmnt --json /proc` command.
	output, err := os.ReadFile("testdata/findmnt-proc.json")
	if err != nil {
		t.Fatal(err)
	}
	result, err := findmnt.ParseOutput(output)
	if err != nil {
		t.Fatal(err)
	}
	expected := &findmnt.Result{
		Filesystems: []findmnt.Filesystem{
			{
				Target:  "/proc",
				Source:  "proc",
				FSType:  "proc",
				Options: "rw,relatime",
			},
		},
	}
	if !result.Equal(expected) {
		t.Fatalf("expected %+v, got %+v", expected, result)
	}
}

func assert[T comparable](t *testing.T, a, b T) {
	if a != b {
		t.Fatalf("expected \"%v\", got \"%v\"", b, a)
	}
}
