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

package command

import (
	"os"
	"testing"
)

func TestKICRunner(t *testing.T) {
	t.Parallel()

	t.Run("TestTempDirectory", func(t *testing.T) {
		t.Parallel()

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get user home directory: %v", err)
		}

		tests := []struct {
			isMinikubeSnap bool
			isDockerSnap   bool
			want           string
		}{
			{false, false, ""},
			{true, true, home},
			{false, true, home},
			{true, false, home},
		}

		for _, tt := range tests {
			got, err := tempDirectory(tt.isMinikubeSnap, tt.isDockerSnap)
			if err != nil {
				t.Fatalf("failed to get temp directory: %v", err)
			}

			if got != tt.want {
				t.Errorf("tempDirectory(%t, %t) = %s; want %s", tt.isMinikubeSnap, tt.isDockerSnap, got, tt.want)
			}
		}
	})
}
