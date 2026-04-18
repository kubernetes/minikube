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

package cluster

import (
	"net"
	"testing"
)

func TestParseHostOnlyNet(t *testing.T) {
	// Representative output of `VBoxManage list hostonlynets` with two nets.
	listOutput := `Name:            minikube-hostonly-192.168.59.1
GUID:            249ae3f6-cab1-4fb2-b92b-defd97d59cd0

State:           Enabled
NetworkMask:     255.255.255.0
LowerIP:         192.168.59.100
UpperIP:         192.168.59.254
VBoxNetworkName: hostonly-minikube-hostonly-192.168.59.1

Name:            other-net
GUID:            abcdef00-0000-0000-0000-000000000000

State:           Enabled
NetworkMask:     255.255.0.0
LowerIP:         10.0.0.100
UpperIP:         10.0.0.200
VBoxNetworkName: hostonly-other-net
`

	tests := []struct {
		name      string
		netName   string
		wantMask  string // dotted-quad of mask bytes
		wantNetIP string
		wantErr   bool
		wantNil   bool // whether both mask+net should be nil (not found)
	}{
		{"first net", "minikube-hostonly-192.168.59.1", "255.255.255.0", "192.168.59.0", false, false},
		{"second net", "other-net", "255.255.0.0", "10.0.0.0", false, false},
		{"not found", "no-such-net", "", "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mask, netAddr, err := parseHostOnlyNet(listOutput, tt.netName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantNil {
				if mask != nil || netAddr != nil {
					t.Errorf("expected (nil, nil), got (%v, %v)", mask, netAddr)
				}
				return
			}
			if mask == nil {
				t.Fatalf("mask is nil")
			}
			if net.IP(mask).String() != tt.wantMask {
				t.Errorf("mask = %v, want %v", net.IP(mask).String(), tt.wantMask)
			}
			if netAddr.String() != tt.wantNetIP {
				t.Errorf("netAddr = %v, want %v", netAddr.String(), tt.wantNetIP)
			}
		})
	}
}

// Also test field-reordering resilience: records where LowerIP appears
// before NetworkMask must still parse correctly.
func TestParseHostOnlyNetFieldReordering(t *testing.T) {
	reordered := `Name:            example
LowerIP:         192.168.100.100
NetworkMask:     255.255.255.0
UpperIP:         192.168.100.200
`
	mask, netAddr, err := parseHostOnlyNet(reordered, "example")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if net.IP(mask).String() != "255.255.255.0" {
		t.Errorf("mask = %v, want 255.255.255.0", net.IP(mask))
	}
	if netAddr.String() != "192.168.100.0" {
		t.Errorf("netAddr = %v, want 192.168.100.0", netAddr)
	}
}
