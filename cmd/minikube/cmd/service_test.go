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

package cmd

import (
	"slices"
	"testing"
)

func TestOpenURLs(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want []string
	}{
		{
			name: "single port",
			rows: [][]string{
				{"default", "web", "http/80", "http://192.168.49.2:30080"},
			},
			want: []string{"http://192.168.49.2:30080"},
		},
		{
			name: "multiple ports",
			rows: [][]string{
				{"default", "web", "http/80\nhttps/443", "http://192.168.49.2:30080\nhttp://192.168.49.2:30443"},
			},
			want: []string{"http://192.168.49.2:30080"},
		},
		{
			name: "multiple services",
			rows: [][]string{
				{"default", "web", "http/80", "http://192.168.49.2:30080"},
				{"default", "api", "", "http://127.0.0.1:51234\nhttp://127.0.0.1:51235"},
			},
			want: []string{"http://192.168.49.2:30080", "http://127.0.0.1:51234"},
		},
		{
			name: "skips rows without urls",
			rows: [][]string{
				{"default", "kubernetes", "No node port"},
				{"default", "web", "", ""},
				{"default", "api", "", "http://127.0.0.1:51234\n"},
			},
			want: []string{"http://127.0.0.1:51234"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var opened []string
			orig := openFirstURL
			openFirstURL = func(urls []string) error {
				opened = append(opened, urls[0])
				return nil
			}
			t.Cleanup(func() { openFirstURL = orig })

			openURLs(tc.rows)

			if !slices.Equal(opened, tc.want) {
				t.Errorf("opened %q, want %q", opened, tc.want)
			}
		})
	}
}
