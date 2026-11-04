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

package main

import "testing"

func TestFindMatchingImageTag(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		ghTag   string
		want    string
		wantErr bool
	}{
		{
			name:  "exact match with v prefix",
			tags:  []string{"v1.25.2", "v1.25.1"},
			ghTag: "v1.25.2",
			want:  "v1.25.2",
		},
		{
			name:  "numeric patch suffix with v prefix",
			tags:  []string{"v1.25.2.8", "v1.25.2"},
			ghTag: "v1.25.2",
			want:  "v1.25.2.8",
		},
		{
			name:  "docker tag without v prefix matches v-prefixed release",
			tags:  []string{"1.25.2.8", "1.25.2"},
			ghTag: "v1.25.2.8",
			want:  "1.25.2.8",
		},
		{
			name:  "suffix match without v prefix",
			tags:  []string{"1.25.2.8", "1.25.1"},
			ghTag: "v1.25.2",
			want:  "1.25.2.8",
		},
		{
			name:  "live shape skips labeller and rhubi variants",
			tags:  []string{"latest", "labeller-rhubi-1.31.0.11", "labeller-1.31.0.11", "rhubi-1.31.0.11", "1.31.0.11", "1.31.0.10"},
			ghTag: "v1.31.0.11",
			want:  "1.31.0.11",
		},
		{
			name:    "variant-only tags do not match",
			tags:    []string{"labeller-1.31.0.11", "rhubi-1.31.0.11"},
			ghTag:   "v1.31.0.11",
			wantErr: true,
		},
		{
			name:    "no match returns error",
			tags:    []string{"1.30.0.1", "latest"},
			ghTag:   "v1.31.0.11",
			wantErr: true,
		},
		{
			name:    "empty tag list returns error",
			tags:    nil,
			ghTag:   "v1.31.0.11",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := findMatchingImageTag(tc.tags, tc.ghTag)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("findMatchingImageTag(%v, %q) = %q, want error", tc.tags, tc.ghTag, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("findMatchingImageTag(%v, %q) returned error: %v", tc.tags, tc.ghTag, err)
			}
			if got != tc.want {
				t.Errorf("findMatchingImageTag(%v, %q) = %q, want %q", tc.tags, tc.ghTag, got, tc.want)
			}
		})
	}
}
