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
	"testing"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestGetDefaults(t *testing.T) {
	tcs := []struct {
		property         string
		expectedContents string
		shouldErr        bool
	}{
		{
			property:         "driver",
			expectedContents: "docker",
		}, {
			property:  "invalid",
			shouldErr: true,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.property, func(t *testing.T) {
			defaults, err := getDefaults(tc.property)
			if err != nil && !tc.shouldErr {
				t.Fatalf("test shouldn't have failed, error listing defaults: %v", err)
			}
			if err == nil && tc.shouldErr {
				t.Fatal("test should have failed but did not")
			}
			if tc.shouldErr {
				return
			}
			for _, d := range defaults {
				if d == tc.expectedContents {
					return
				}
			}
			t.Fatalf("defaults didn't contain expected default. Actual: %v\nExpected: %v\n", defaults, tc.expectedContents)
		})
	}
}

func TestPrintDefaults(t *testing.T) {
	defaults := []string{"a", "b", "c"}
	tcs := []struct {
		description string
		format      string
		expected    string
	}{
		{
			description: "print to stdout",
			expected:    "* a\n* b\n* c\n",
		}, {
			description: "print in json",
			format:      "json",
			expected:    "[\"a\",\"b\",\"c\"]\n",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			defaultsOutput = tc.format
			f := tests.NewFakeFile()
			out.SetOutFile(f)
			printDefaults(defaults)
			if f.String() != tc.expected {
				t.Fatalf("Expected: %v\n Actual: %v\n", tc.expected, f.String())
			}
		})
	}
}
