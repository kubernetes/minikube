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

package oci

import (
	"testing"
)

func TestParseMountString(t *testing.T) {
	testCases := []struct {
		Name          string
		MountString   string
		ExpectErr     bool
		ExpectedMount Mount
	}{
		{
			Name:        "basic linux",
			MountString: "/foo:/bar",
			ExpectErr:   false,
			ExpectedMount: Mount{
				HostPath:      "/foo",
				ContainerPath: "/bar",
			},
		},
		{
			Name:        "linux read only",
			MountString: "/foo:/bar:ro",
			ExpectErr:   false,
			ExpectedMount: Mount{
				HostPath:      "/foo",
				ContainerPath: "/bar",
				Readonly:      true,
			},
		},
		{
			Name:        "windows style",
			MountString: "C:\\Windows\\Path:/foo",
			ExpectErr:   false,
			ExpectedMount: Mount{
				HostPath:      "C:\\Windows\\Path",
				ContainerPath: "/foo",
			},
		},
		{
			Name:        "windows style read/write",
			MountString: "C:\\Windows\\Path:/foo:rw",
			ExpectErr:   false,
			ExpectedMount: Mount{
				HostPath:      "C:\\Windows\\Path",
				ContainerPath: "/foo",
				Readonly:      false,
			},
		},
		{
			Name:        "container only",
			MountString: "/foo",
			ExpectErr:   false,
			ExpectedMount: Mount{
				ContainerPath: "/foo",
			},
		},
		{
			Name:        "selinux relabel & bidirectional propagation",
			MountString: "/foo:/bar/baz:Z,rshared",
			ExpectErr:   false,
			ExpectedMount: Mount{
				HostPath:       "/foo",
				ContainerPath:  "/bar/baz",
				SelinuxRelabel: true,
				Propagation:    MountPropagationBidirectional,
			},
		},
		{
			Name:        "invalid mount option",
			MountString: "/foo:/bar:Z,bat",
			ExpectErr:   true,
			ExpectedMount: Mount{
				HostPath:       "/foo",
				ContainerPath:  "/bar",
				SelinuxRelabel: true,
			},
		},
		{
			Name:          "empty spec",
			MountString:   "",
			ExpectErr:     false,
			ExpectedMount: Mount{},
		},
		{
			Name:        "relative container path",
			MountString: "/foo/bar:baz/bat:private",
			ExpectErr:   true,
			ExpectedMount: Mount{
				HostPath:      "/foo/bar",
				ContainerPath: "baz/bat",
				Propagation:   MountPropagationNone,
			},
		},
	}

	for _, tc := range testCases {
		mount, err := ParseMountString(tc.MountString)
		if err != nil && !tc.ExpectErr {
			t.Errorf("Unexpected error for \"%s\": %v", tc.Name, err)
		}
		if err == nil && tc.ExpectErr {
			t.Errorf("Expected error for \"%s\" but didn't get any: %v %v", tc.Name, mount, err)
		}
		if mount != tc.ExpectedMount {
			t.Errorf("Unexpected mount for \"%s\":\n expected %+v\ngot %+v", tc.Name, tc.ExpectedMount, mount)
		}
	}
}
