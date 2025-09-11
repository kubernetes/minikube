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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMntCmd(t *testing.T) {
	var tests = []struct {
		name   string
		source string
		target string
		cfg    *MountConfig
		want   string
	}{
		{
			name:   "simple",
			source: "src",
			target: "target",
			cfg:    &MountConfig{Type: "9p"},
			want:   "sudo mount -t 9p -o dfltgid=0,dfltuid=0,trans=tcp src target",
		},
		{
			name:   "named uid",
			source: "src",
			target: "target",
                       cfg:    &MountConfig{Type: "9p", UID: "root", GID: "root"},
                       want:   "sudo mount -t 9p -o dfltgid=$(grep ^root: /etc/group | cut -d: -f3),dfltuid=$(id -u root),trans=tcp src target",
		},
		{
			name:   "everything",
			source: "10.0.0.1",
			target: "/target",
			cfg: &MountConfig{Type: "9p", UID: "82", GID: "72", Version: "9p2000.u", Options: map[string]string{
				"noextend": "",
				"cache":    "fscache",
			}},
			want: "sudo mount -t 9p -o cache=fscache,dfltgid=72,dfltuid=82,noextend,trans=tcp,version=9p2000.u 10.0.0.1 /target",
		},
		{
			name:   "version-conflict",
			source: "src",
			target: "tgt",
			cfg: &MountConfig{Type: "9p", Version: "9p2000.u", Options: map[string]string{
				"version": "9p2000.L",
			}},
			want: "sudo mount -t 9p -o dfltgid=0,dfltuid=0,trans=tcp,version=9p2000.L src tgt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mntCmd(tc.source, tc.target, tc.cfg)
			want := tc.want
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("command diff (-want +got): %s", diff)
			}
		})
	}
}
