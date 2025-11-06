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

package bsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
)

type copyFailRunner struct {
	command.Runner
}

func (copyFailRunner) Copy(_ assets.CopyableFile) error {
	return fmt.Errorf("test error during copy file")
}

func newFakeCommandRunnerCopyFail() command.Runner {
	return copyFailRunner{command.NewFakeCommandRunner()}
}

func TestCopyBinary(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source")
	if err := os.WriteFile(srcFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	tests := []struct {
		name    string
		runner  command.Runner
		src     string
		dst     string
		wantErr bool
	}{
		{
			name:    "missing source",
			runner:  command.NewFakeCommandRunner(),
			src:     filepath.Join(tmpDir, "missing"),
			dst:     filepath.Join(tmpDir, "dest"),
			wantErr: true,
		},
		{
			name:    "success",
			runner:  command.NewFakeCommandRunner(),
			src:     srcFile,
			dst:     filepath.Join(tmpDir, "dest"),
			wantErr: false,
		},
		{
			name:    "copy failure",
			runner:  newFakeCommandRunnerCopyFail(),
			src:     srcFile,
			dst:     filepath.Join(tmpDir, "dest"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := copyBinary(tc.runner, tc.src, tc.dst)
			if (err != nil) != tc.wantErr {
				t.Fatalf("copyBinary() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
