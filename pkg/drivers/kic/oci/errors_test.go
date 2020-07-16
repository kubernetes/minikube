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

import "testing"

func TestFailFastError(t *testing.T) {
	tcs := []struct {
		description      string
		err              error
		shouldBeFailFast bool
	}{
		{
			description:      "fail fast error",
			err:              ErrWindowsContainers,
			shouldBeFailFast: true,
		}, {
			description: "not a fail fast error",
			err:         ErrExitedUnexpectedly,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			_, ff := tc.err.(*FailFastError)
			if ff != tc.shouldBeFailFast {
				t.Fatalf("expected fail fast to be %v, was %v", tc.shouldBeFailFast, ff)
			}
		})
	}
}
