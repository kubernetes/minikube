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

package virtualbox

import (
	"net"
	"testing"
)

func TestFindHostOnlyNetByCIDR(t *testing.T) {
	nets := map[string]*hostOnlyNet{
		"minikube-hostonly-192.168.59.1": {
			Name:        "minikube-hostonly-192.168.59.1",
			NetworkMask: parseIPv4Mask("255.255.255.0"),
			LowerIP:     net.ParseIP("192.168.59.100"),
			UpperIP:     net.ParseIP("192.168.59.254"),
		},
		"other": {
			Name:        "other",
			NetworkMask: parseIPv4Mask("255.255.0.0"),
			LowerIP:     net.ParseIP("10.0.0.100"),
			UpperIP:     net.ParseIP("10.0.0.200"),
		},
	}

	tests := []struct {
		name     string
		hostIP   string
		mask     string
		wantName string // expected matched net name, empty = no match
	}{
		{"exact subnet match (host .1)", "192.168.59.1", "255.255.255.0", "minikube-hostonly-192.168.59.1"},
		{"exact subnet match (host .200)", "192.168.59.200", "255.255.255.0", "minikube-hostonly-192.168.59.1"},
		{"different subnet, same mask", "192.168.60.1", "255.255.255.0", ""},
		{"wrong mask width", "192.168.59.1", "255.255.0.0", ""},
		{"match second record", "10.0.5.5", "255.255.0.0", "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostIP := net.ParseIP(tt.hostIP)
			mask := parseIPv4Mask(tt.mask)
			got := findHostOnlyNetByCIDR(nets, hostIP, mask)
			if tt.wantName == "" {
				if got != nil {
					t.Errorf("expected nil, got %q", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected match %q, got nil", tt.wantName)
			}
			if got.Name != tt.wantName {
				t.Errorf("got %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

// listHostOnlyNets coverage: happy path + duplicate Name detection.
func TestListHostOnlyNets(t *testing.T) {
	out := `Name:            alpha
GUID:            00000000-0000-0000-0000-000000000001

State:           Enabled
NetworkMask:     255.255.255.0
LowerIP:         192.168.10.100
UpperIP:         192.168.10.200
VBoxNetworkName: hostonly-alpha

Name:            beta
GUID:            00000000-0000-0000-0000-000000000002

State:           Disabled
NetworkMask:     255.255.0.0
LowerIP:         10.0.0.100
UpperIP:         10.0.0.200
VBoxNetworkName: hostonly-beta
`
	vbox := &VBoxManagerMock{args: "list hostonlynets", stdOut: out}
	nets, err := listHostOnlyNets(vbox)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(nets) != 2 {
		t.Fatalf("expected 2 nets, got %d", len(nets))
	}
	if nets["alpha"].VBoxNetworkName != "hostonly-alpha" {
		t.Errorf("alpha VBoxNetworkName = %q", nets["alpha"].VBoxNetworkName)
	}
	if !nets["alpha"].Enabled {
		t.Errorf("alpha Enabled should be true")
	}
	if nets["beta"].Enabled {
		t.Errorf("beta Enabled should be false")
	}
	if nets["alpha"].LowerIP.String() != "192.168.10.100" {
		t.Errorf("alpha LowerIP = %v", nets["alpha"].LowerIP)
	}
}

func TestListHostOnlyNetsDuplicate(t *testing.T) {
	// Two records with the same Name should error.
	out := `Name:            same
GUID:            11111111-1111-1111-1111-111111111111

State:           Enabled
NetworkMask:     255.255.255.0
LowerIP:         192.168.10.100
UpperIP:         192.168.10.200
VBoxNetworkName: hostonly-same-1

Name:            same
GUID:            22222222-2222-2222-2222-222222222222

State:           Enabled
NetworkMask:     255.255.0.0
LowerIP:         10.0.0.100
UpperIP:         10.0.0.200
VBoxNetworkName: hostonly-same-2
`
	vbox := &VBoxManagerMock{args: "list hostonlynets", stdOut: out}
	_, err := listHostOnlyNets(vbox)
	if err == nil {
		t.Fatalf("expected error for duplicate Name, got nil")
	}
}

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
