/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package main

import "testing"

const containerd2Config = `version = 3
root = '/var/lib/containerd'

[plugins]
  [plugins.'io.containerd.cri.v1.images']
    snapshotter = 'overlayfs'

  [plugins.'io.containerd.cri.v1.runtime'.containerd]
      [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes]
        [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc]
          snapshotter = ''
`

const containerd17Config = `disabled_plugins = []
imports = ["/etc/containerd/config.toml"]
oom_score = 0
plugin_dir = ""
required_plugins = []
root = "/var/lib/containerd"
state = "/run/containerd"
temp = ""
version = 2

[plugins]

  [plugins."io.containerd.grpc.v1.cri"]
    [plugins."io.containerd.grpc.v1.cri".containerd]
      snapshotter = "overlayfs"

      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]

        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          snapshotter = ""
`

func TestParseContainerdSnapshotterCases(t *testing.T) {
	tcs := []struct {
		name string
		cfg  string
		want string
	}{
		{name: "v2 config with single quotes", cfg: containerd2Config, want: "overlayfs"},
		{name: "v1.7 config with double quotes", cfg: containerd17Config, want: "overlayfs"},
		{name: "single line double quotes", cfg: "snapshotter = \"overlayfs\"\n", want: "overlayfs"},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := parseContainerdSnapshotter(tc.cfg); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
