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

package docker

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergeReferenceStores(t *testing.T) {
	initial := ReferenceStore{
		Repositories: map[string]repository{
			"image1": {
				"r1": "d1",
				"r2": "d2",
			},
			"image2": {
				"r1": "d1",
				"r2": "d2",
			},
		},
	}

	afterPreload := ReferenceStore{
		Repositories: map[string]repository{
			"image1": {
				"r1": "updated",
				"r2": "updated",
			},
			"image3": {
				"r3": "d3",
			},
		},
	}

	expected := ReferenceStore{
		Repositories: map[string]repository{
			"image1": {
				"r1": "updated",
				"r2": "updated",
			},
			"image2": {
				"r1": "d1",
				"r2": "d2",
			},
			"image3": {
				"r3": "d3",
			},
		},
	}

	s := &Storage{
		refStores: []ReferenceStore{initial, afterPreload},
	}

	actual := s.mergeReferenceStores()
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("Actual: %v, Expected: %v, Diff: %s", actual, expected, diff)
	}
}
