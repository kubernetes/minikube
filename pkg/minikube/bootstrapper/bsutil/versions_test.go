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

package bsutil

import (
	"testing"

	"github.com/blang/semver/v4"
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
			description: "greater than max version",
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
			between := versionIsBetween(test.ver, test.gte, test.lte)
			if between != test.expected {
				t.Errorf("Expected: %t, Actual: %t", test.expected, between)
			}
		})
	}
}
