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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

// streamRunner records what transferAndLoadImage did rather than emulating a
// host: the question under test is which path it took, not what the runtime
// replied.
type streamRunner struct {
	copied []string
	cmds   []string
	stdin  string
}

func (r *streamRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	r.cmds = append(r.cmds, strings.Join(cmd.Args, " "))
	if cmd.Stdin != nil {
		b, err := io.ReadAll(cmd.Stdin)
		if err != nil {
			return &command.RunResult{}, err
		}
		r.stdin = string(b)
	}
	return &command.RunResult{}, nil
}

func (r *streamRunner) StartCmd(*exec.Cmd) (*command.StartedCmd, error) {
	return &command.StartedCmd{}, nil
}

func (r *streamRunner) WaitCmd(*command.StartedCmd) (*command.RunResult, error) {
	return &command.RunResult{}, nil
}

func (r *streamRunner) Copy(f assets.CopyableFile) error {
	r.copied = append(r.copied, f.GetTargetName())
	return nil
}

func (r *streamRunner) CopyFrom(assets.CopyableFile) error { return nil }

func (r *streamRunner) Remove(assets.CopyableFile) error { return nil }

func (r *streamRunner) ReadableFile(string) (assets.ReadableFile, error) { return nil, nil }

// TestTransferAndLoadImageStreams is the behavioural claim of this change: for a
// runtime that can read the image from stdin, nothing is written to the guest's
// disk on the way in. The image reaching the runtime intact is asserted too,
// because "no copy happened" is also true of a path that loads nothing at all.
func TestTransferAndLoadImageStreams(t *testing.T) {
	const payload = "fake image tarball"

	tests := []struct {
		runtime  string
		wantCopy bool
	}{
		{"docker", false},
		{"crio", false},
		// containerd keeps the copy until the ctr/nerdctl swap lands.
		{"containerd", true},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			src := filepath.Join(t.TempDir(), "image.tar")
			if err := os.WriteFile(src, []byte(payload), 0o644); err != nil {
				t.Fatalf("writing image: %v", err)
			}
			runner := &streamRunner{}

			k8s := config.KubernetesConfig{ContainerRuntime: tc.runtime}
			if err := transferAndLoadImage(runner, k8s, src, "image.tar"); err != nil {
				t.Fatalf("transferAndLoadImage: %v", err)
			}

			if gotCopy := len(runner.copied) > 0; gotCopy != tc.wantCopy {
				t.Errorf("copied to guest = %v (%v), want %v",
					gotCopy, runner.copied, tc.wantCopy)
			}
			if tc.wantCopy {
				return
			}
			if runner.stdin != payload {
				t.Errorf("stdin = %q, want %q", runner.stdin, payload)
			}
			// A streamed load must not name a guest path it never created.
			if got := strings.Join(runner.cmds, " "); strings.Contains(got, loadRoot) {
				t.Errorf("command %q still refers to %s", got, loadRoot)
			}
		})
	}
}
