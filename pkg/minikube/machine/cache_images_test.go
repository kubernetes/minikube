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

package machine

import (
	"fmt"
	"testing"

	"k8s.io/minikube/pkg/minikube/cruntime"
)

// mockRuntime implements the subset of cruntime.Manager used by removeExistingImage.
type mockRuntime struct {
	cruntime.Manager
	removeErr error    // error to return from RemoveImage
	removed   []string // records which images RemoveImage was called with
}

func (m *mockRuntime) RemoveImage(name string) error {
	m.removed = append(m.removed, name)
	return m.removeErr
}

func TestRemoveExistingImage(t *testing.T) {
	tests := []struct {
		name      string
		src       string // source path of the cached image file (e.g., "/cache/myapp_latest")
		imgName   string // image name to remove (e.g., "myapp:latest")
		removeErr error  // error that mockRuntime.RemoveImage returns
		wantCalls int    // expected number of RemoveImage invocations
	}{
		{
			name:      "src equals imgName should skip removal",
			src:       "/tmp/image.tar",
			imgName:   "/tmp/image.tar",
			wantCalls: 0,
		},
		{
			name:      "successful removal",
			src:       "/cache/myapp_latest",
			imgName:   "myapp:latest",
			removeErr: nil,
			wantCalls: 1,
		},
		{
			name:      "removal error should not prevent loading",
			src:       "/cache/myapp_latest",
			imgName:   "myapp:latest",
			removeErr: fmt.Errorf("removal failed for any reason"),
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockRuntime{removeErr: tt.removeErr}
			removeExistingImage(r, tt.src, tt.imgName)
			if len(r.removed) != tt.wantCalls {
				t.Errorf("RemoveImage called %d times, want %d", len(r.removed), tt.wantCalls)
			}
		})
	}
}
