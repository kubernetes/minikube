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

package cmd

import (
	"testing"
)

func TestServiceForwardOpen(t *testing.T) {
	var tests = []struct {
		name           string
		serviceURLMode bool
		all            bool
		args           []string
		want           bool
	}{
		{
			name:           "multiple_urls",
			serviceURLMode: false,
			all:            false,
			args:           []string{"test-service-1", "test-service-2"},
			want:           false,
		},
		{
			name:           "service_url_mode",
			serviceURLMode: true,
			all:            false,
			args:           []string{"test-service-1"},
			want:           false,
		},
		{
			name:           "all",
			serviceURLMode: false,
			all:            true,
			args:           []string{"test-service-1", "test-service-2"},
			want:           false,
		},
		{
			name:           "single_url",
			serviceURLMode: false,
			all:            false,
			args:           []string{"test-service-1"},
			want:           true,
		},
	}

	for _, tc := range tests {
		serviceURLMode = tc.serviceURLMode
		all = tc.all
		t.Run(tc.name, func(t *testing.T) {
			got := shouldOpen(tc.args)
			if got != tc.want {
				t.Errorf("bool(%+v) = %t, want: %t", "shouldOpen", got, tc.want)
			}
		})
	}
}
