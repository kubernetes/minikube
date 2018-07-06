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

package kubeadm

import (
	"reflect"
	"testing"

	"github.com/blang/semver"
)

func TestVersionIsBetween(t *testing.T) {
	tests := []struct {
		description string
		ver         semver.Version
		gte         semver.Version
		lte         semver.Version
		expected    bool
	}{
		{
			description: "between",
			ver:         semver.MustParse("1.8.0"),
			gte:         semver.MustParse("1.7.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    true,
		},
		{
			description: "less than minimum version",
			ver:         semver.MustParse("1.6.0"),
			gte:         semver.MustParse("1.7.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    false,
		},
		{
			description: "greather than max version",
			ver:         semver.MustParse("2.8.0"),
			gte:         semver.MustParse("1.7.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    true,
		},
		{
			description: "equal to max version",
			ver:         semver.MustParse("1.9.0"),
			gte:         semver.MustParse("1.7.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    true,
		},
		{
			description: "equal to min version",
			ver:         semver.MustParse("1.7.0"),
			gte:         semver.MustParse("1.7.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    true,
		},
		{
			description: "alpha between",
			ver:         semver.MustParse("1.8.0-alpha.0"),
			gte:         semver.MustParse("1.8.0"),
			lte:         semver.MustParse("1.9.0"),
			expected:    true,
		},
		{
			description: "beta greater than alpha",
			ver:         semver.MustParse("1.8.0-beta.1"),
			gte:         semver.MustParse("1.8.0"),
			lte:         semver.MustParse("1.8.0-alpha.0"),
			expected:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			between := VersionIsBetween(test.ver, test.gte, test.lte)
			if between != test.expected {
				t.Errorf("Expected: %t, Actual: %t", test.expected, between)
			}
		})
	}
}

func TestParseKubernetesVersion(t *testing.T) {
	version, err := ParseKubernetesVersion("v1.8.0-alpha.5")
	if err != nil {
		t.Fatalf("Error parsing version: %s", err)
	}
	if version.NE(semver.MustParse("1.8.0-alpha.5")) {
		t.Errorf("Expected: %s, Actual:%s", "1.8.0-alpha.5", version)
	}
}

func TestParseFeatureArgs(t *testing.T) {
	tests := []struct {
		description                  string
		featureGates                 string
		expectedKubeadmFeatureArgs   map[string]bool
		expectedComponentFeatureArgs string
	}{
		{
			description:  "only kubeadm feature",
			featureGates: "Auditing=true,SelfHosting=false",
			expectedKubeadmFeatureArgs: map[string]bool{
				"Auditing":    true,
				"SelfHosting": false,
			},
			expectedComponentFeatureArgs: "",
		},
		{
			description:                  "only component feature",
			featureGates:                 "PodPriority=true,Accelerators=false",
			expectedKubeadmFeatureArgs:   map[string]bool{},
			expectedComponentFeatureArgs: "PodPriority=true,Accelerators=false",
		},
		{
			description:  "between component and kubeadm feature",
			featureGates: "Auditing=true,PodPriority=true,SelfHosting=false,Accelerators=false",
			expectedKubeadmFeatureArgs: map[string]bool{
				"Auditing":    true,
				"SelfHosting": false,
			},
			expectedComponentFeatureArgs: "PodPriority=true,Accelerators=false",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			kubeadm, component, err := ParseFeatureArgs(test.featureGates)

			if err != nil {
				t.Fatalf("Error parsing feature args: %s", err)
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
