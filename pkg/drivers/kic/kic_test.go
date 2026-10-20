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

package kic

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestGetIPNetwork checks that GetIP picks the IP of the network the node was created on:
// --network if set, the profile name otherwise.
func TestGetIPNetwork(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as the container runtime binary")
	}
	// A container runtime that reports two networks, as `docker container inspect` does
	// after `docker network connect`; the extra one sorts first.
	bin := filepath.Join(t.TempDir(), "docker")
	script := "#!/bin/sh\nprintf 'bridge,172.17.0.2,\\nminikube,192.168.49.2,\\n\\n'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		cluster string
		network string
		want    string
		wantErr bool
	}{
		{name: "profile name when no network is set", cluster: "minikube", want: "192.168.49.2"},
		{name: "network flag wins over profile name", cluster: "bridge", network: "minikube", want: "192.168.49.2"},
		{name: "network flag selects the other network", cluster: "minikube", network: "bridge", want: "172.17.0.2"},
		{name: "no network matches", cluster: "other", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDriver(Config{MachineName: "minikube", ClusterName: tc.cluster, Network: tc.network, OCIBinary: bin})
			got, err := d.GetIP()
			if (err != nil) != tc.wantErr {
				t.Fatalf("GetIP() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("GetIP() = %q, want %q", got, tc.want)
			}
		})
	}
}
