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

package mac

import (
	"net"
	"testing"
)

func TestFromName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"minikube"},
		{"my-cluster"},
		{"test-profile-123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := FromName(tt.name)

			hw, err := net.ParseMAC(addr)
			if err != nil {
				t.Fatalf("invalid MAC address %q: %v", addr, err)
			}

			// Must be locally administered (bit 1 of first byte set)
			if hw[0]&2 == 0 {
				t.Errorf("MAC %s is not locally administered", addr)
			}

			// Must be unicast (bit 0 of first byte clear)
			if hw[0]&1 != 0 {
				t.Errorf("MAC %s is not unicast", addr)
			}

			// Must be deterministic
			if addr2 := FromName(tt.name); addr != addr2 {
				t.Errorf("not deterministic: %s != %s", addr, addr2)
			}
		})
	}

	// Different names must produce different MACs
	if FromName("cluster-a") == FromName("cluster-b") {
		t.Error("different names produced the same MAC")
	}
}
