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

package qemu2

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte("firmware"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", p, err)
	}
	return p
}

// TestFirstExistingFirmware covers the Ubuntu 24.04 case from #23628 where
// only OVMF_CODE_4M.fd ships and the legacy OVMF_CODE.fd name is missing.
func TestFirstExistingFirmware(t *testing.T) {
	preferred := writeTempFile(t, "OVMF_CODE.fd")
	fallback := writeTempFile(t, "OVMF_CODE_4M.fd")

	tests := []struct {
		name  string
		paths []string
		want  string
	}{
		{"preferred exists", []string{preferred, fallback}, preferred},
		{"only fallback exists", []string{filepath.Join(t.TempDir(), "missing.fd"), fallback}, fallback},
		{"none exists returns preferred", []string{preferred + ".missing", fallback + ".missing"}, preferred + ".missing"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstExistingFirmware(tc.paths...); got != tc.want {
				t.Errorf("firstExistingFirmware(%v) = %q, want %q", tc.paths, got, tc.want)
			}
		})
	}
}
