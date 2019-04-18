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

package cluster

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockMountRunner struct {
	cmds []string
	T    *testing.T
}

func newMockMountRunner(t *testing.T) *mockMountRunner {
	return &mockMountRunner{
		T:    t,
		cmds: []string{},
	}
}

func (m *mockMountRunner) CombinedOutput(cmd string) (string, error) {
	m.cmds = append(m.cmds, cmd)
	return "", nil
}

func TestMount(t *testing.T) {
	var tests = []struct {
		name   string
		source string
		target string
		cfg    *MountConfig
		want   []string
	}{
		{
			name:   "simple",
			source: "src",
			target: "target",
			cfg:    &MountConfig{Type: "9p", Mode: os.FileMode(0700)},
			want: []string{
				"findmnt -T target | grep target && sudo umount target || true",
				"sudo mkdir -m 700 -p target && sudo mount -t 9p -o dfltgid=0,dfltuid=0 src target",
			},
		},
		{
			name:   "named uid",
			source: "src",
			target: "target",
			cfg:    &MountConfig{Type: "9p", Mode: os.FileMode(0700), UID: "docker", GID: "docker"},
			want: []string{
				"findmnt -T target | grep target && sudo umount target || true",
				"sudo mkdir -m 700 -p target && sudo mount -t 9p -o dfltgid=$(grep ^docker: /etc/group | cut -d: -f3),dfltuid=$(id -u docker) src target",
			},
		},
		{
			name:   "everything",
			source: "10.0.0.1",
			target: "/target",
			cfg: &MountConfig{Type: "9p", Mode: os.FileMode(0777), UID: "82", GID: "72", Version: "9p2000.u", Options: map[string]string{
				"noextend": "",
				"cache":    "fscache",
			}},
			want: []string{
				"findmnt -T /target | grep /target && sudo umount /target || true",
				"sudo mkdir -m 777 -p /target && sudo mount -t 9p -o cache=fscache,dfltgid=72,dfltuid=82,noextend,version=9p2000.u 10.0.0.1 /target",
			},
		},
		{
			name:   "version-conflict",
			source: "src",
			target: "tgt",
			cfg: &MountConfig{Type: "9p", Mode: os.FileMode(0700), Version: "9p2000.u", Options: map[string]string{
				"version": "9p2000.L",
			}},
			want: []string{
				"findmnt -T tgt | grep tgt && sudo umount tgt || true",
				"sudo mkdir -m 700 -p tgt && sudo mount -t 9p -o dfltgid=0,dfltuid=0,version=9p2000.L src tgt",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newMockMountRunner(t)
			err := Mount(r, tc.source, tc.target, tc.cfg)
			if err != nil {
				t.Fatalf("Mount(%s, %s, %+v): %v", tc.source, tc.target, tc.cfg, err)
			}
			if diff := cmp.Diff(r.cmds, tc.want); diff != "" {
				t.Errorf("command diff (-want +got): %s", diff)
			}
		})
	}
}

func TestUnmount(t *testing.T) {
	r := newMockMountRunner(t)
	err := Unmount(r, "/mnt")
	if err != nil {
		t.Fatalf("Unmount(/mnt): %v", err)
	}

	want := []string{"findmnt -T /mnt | grep /mnt && sudo umount /mnt || true"}
	if diff := cmp.Diff(r.cmds, want); diff != "" {
		t.Errorf("command diff (-want +got): %s", diff)
	}
}
