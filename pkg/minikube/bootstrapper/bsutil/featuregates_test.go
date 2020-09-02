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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"reflect"
	"testing"
)

func TestParseFeatureArgs(t *testing.T) {
	tests := []struct {
		description                  string
		featureGates                 string
		expectedKubeadmFeatureArgs   map[string]bool
		expectedComponentFeatureArgs string
	}{
		{
			description:  "IPv6DualStack enabled",
			featureGates: "IPv6DualStack=true",
			expectedKubeadmFeatureArgs: map[string]bool{
				"IPv6DualStack": true,
			},
			expectedComponentFeatureArgs: "",
		},
		{
			description:  "IPv6DualStack disabled",
			featureGates: "IPv6DualStack=false",
			expectedKubeadmFeatureArgs: map[string]bool{
				"IPv6DualStack": false,
			},
			expectedComponentFeatureArgs: "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			kubeadm, component, err := parseFeatureArgs(test.featureGates)

			if err != nil {
				t.Fatalf("Error parsing feature args: %v", err)
			}

			if !reflect.DeepEqual(kubeadm, test.expectedKubeadmFeatureArgs) {
				t.Errorf("Kubeadm Actual: %v, Expected: %v", kubeadm, test.expectedKubeadmFeatureArgs)
			}

			if !reflect.DeepEqual(component, test.expectedComponentFeatureArgs) {
				t.Errorf("Component Actual: %v, Expected: %v", component, test.expectedComponentFeatureArgs)
			}
		})
	}
}

func TestSupport(t *testing.T) {

	tests := []struct {
		name     string
		expected bool
	}{
		{
			name:     "IPv6DualStack",
			expected: true,
		},
		{
			name:     "Life is Beautiful !",
			expected: false,
		},
		{
			name:     "",
			expected: false,
		},
	}
	for _, tc := range tests {
		if supportedFG(tc.name) != tc.expected {
			t.Errorf("expected supportedFG(%s) to be %t ! ", tc.name, tc.expected)
		}
	}

}
