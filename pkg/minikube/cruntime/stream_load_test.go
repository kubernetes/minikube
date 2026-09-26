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

package cruntime

import (
	"strings"
	"testing"
)

// payload stands in for an image tarball. Its contents do not matter, only that
// every byte reaches the runtime's load command on stdin.
const payload = "fake image tarball"

func TestLoadImageStream(t *testing.T) {
	tests := []struct {
		runtime string
		want    string
	}{
		{"docker", "docker load"},
		{"crio", "sudo podman load"},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			r, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%q): %v", tc.runtime, err)
			}
			sl, ok := r.(StreamLoader)
			if !ok {
				t.Fatalf("%s does not implement StreamLoader", r.Name())
			}
			// New may run commands of its own; only the load matters here.
			runner.cmds = []string{}

			if err := sl.LoadImageStream(strings.NewReader(payload)); err != nil {
				t.Fatalf("LoadImageStream: %v", err)
			}

			if got := strings.Join(runner.cmds, " "); got != tc.want {
				t.Errorf("command = %q, want %q", got, tc.want)
			}
			// The whole point of the change: the image arrives as bytes on stdin
			// rather than as a path to a file already copied into the guest. A
			// command that looks right but receives nothing still loads nothing.
			if runner.stdin != payload {
				t.Errorf("stdin = %q, want %q", runner.stdin, payload)
			}
		})
	}
}

// containerd is deliberately left out of this PR: swapping ctr for nerdctl also
// means removing the #22309 retry, which needs its own thread. Until then it
// must keep working through the file-based path, so it must NOT satisfy
// StreamLoader, or transferAndLoadImage would route it to a method it lacks.
func TestContainerdDoesNotStreamYet(t *testing.T) {
	r, err := New(Config{Type: "containerd", Runner: NewFakeRunner(t)})
	if err != nil {
		t.Fatalf("New(containerd): %v", err)
	}
	if _, ok := r.(StreamLoader); ok {
		t.Error("containerd implements StreamLoader; if that is intentional, " +
			"the ctr retry in LoadImage needs to be resolved in the same change")
	}
}
