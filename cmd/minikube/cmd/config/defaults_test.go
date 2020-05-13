/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"

	"k8s.io/minikube/pkg/minikube/out"
)

func TestListDefaults(t *testing.T) {
	tcs := []struct {
		property         string
		expectedContents string
		shouldErr        bool
	}{
		{
			property:         "driver",
			expectedContents: "* docker\n*",
		}, {
			property:  "invalid",
			shouldErr: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.property, func(t *testing.T) {
			f := tests.NewFakeFile()
			out.SetOutFile(f)
			err := listDefaults(tc.property)
			if err != nil && !tc.shouldErr {
				t.Fatalf("test shouldn't have failed, error listing defaults: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Fatal("test should have failed but did not")
			}
			actual := f.String()
			if !strings.Contains(actual, tc.expectedContents) {
				t.Fatalf("actual contents don't contain expected contents. Actual: %v\nExpected: %v\n", actual, tc.expectedContents)
			}
		})
	}
}
