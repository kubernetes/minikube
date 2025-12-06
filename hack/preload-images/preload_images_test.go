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
