/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package images

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_kubernetesRepo(t *testing.T) {
	tests := []struct {
		mirror string
		want   string
	}{
		{
			"",
			DefaultKubernetesRepo,
		},
		{
			"mirror.k8s.io",
			"mirror.k8s.io",
		},
	}
	for _, tc := range tests {
		got := kubernetesRepo(tc.mirror)
		if !cmp.Equal(got, tc.want) {
			t.Errorf("mirror mismatch, want: %s, got: %s", tc.want, got)
		}
	}

}
