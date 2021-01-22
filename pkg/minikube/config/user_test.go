/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"strings"
	"testing"
)

func TestUser(t *testing.T) {
	t.Run("Length", func(t *testing.T) {
		tests := []struct {
			in   string
			want bool
		}{
			{strings.Repeat("a", 60), true},
			{strings.Repeat("a", 61), false},
		}

		for _, tt := range tests {
			got := UserNameValid(tt.in)

			if got != tt.want {
				t.Errorf("UserNameValid(%q, length: %d) = %t; want %t", tt.in, len(tt.in), got, tt.want)
			}
		}
	})
}
