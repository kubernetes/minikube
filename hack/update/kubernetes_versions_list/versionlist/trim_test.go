/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package versionlist

import (
	"reflect"
	"testing"
)

func TestTrimKubernetesVersions(t *testing.T) {
	tests := []struct {
		name           string
		defaultVersion string
		releases       []string
		want           []string
	}{
		{
			name:           "current default",
			defaultVersion: "v1.37.0",
			releases: []string{
				"v1.38.0-alpha.1", // newer than the default
				"v1.37.0",
				"v1.36.4",
				"v1.35.8",
				"v1.34.11", // most recent EOL minor
				"v1.34.0-rc.1",
			},
			want: []string{
				"v1.37.0",
				"v1.36.4",
				"v1.35.8",
				"v1.34.11",
				"v1.34.0-rc.1",
			},
		},
		{
			name:           "after default advances",
			defaultVersion: "v1.38.0",
			releases: []string{
				"v1.39.0-alpha.1", // newer than the default
				"v1.38.0",
				"v1.38.0-alpha.1",
				"v1.37.0",
				"v1.36.4",
				"v1.35.8",  // most recent EOL minor after the default advances
				"v1.34.11", // now older than the EOL minor
			},
			want: []string{
				"v1.38.0",
				"v1.38.0-alpha.1",
				"v1.37.0",
				"v1.36.4",
				"v1.35.8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TrimKubernetesVersions(tt.releases, tt.defaultVersion)
			if err != nil {
				t.Fatalf("TrimKubernetesVersions() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("TrimKubernetesVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimKubernetesVersionsRequiresDefaultRelease(t *testing.T) {
	_, err := TrimKubernetesVersions([]string{"v1.37.0", "v1.36.4"}, "v1.38.0")
	if err == nil {
		t.Fatal("TrimKubernetesVersions() error = nil, want missing default release error")
	}
}
