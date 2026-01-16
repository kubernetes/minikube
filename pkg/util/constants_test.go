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

package util

import (
	"net"
	"testing"
)

func TestGetServiceClusterIP(t *testing.T) {
	testData := []struct {
		serviceCIRD string
		expectedIP  string
		err         bool
	}{
		{"1111.0.0.1/12", "", true},
		{"10.96.0.0/24", "10.96.0.1", false},
	}

	for _, tt := range testData {
		ip, err := ServiceClusterIP(tt.serviceCIRD)
		if err != nil && !tt.err {
			t.Fatalf("GetServiceClusterIP() err = %v", err)
		}
		if err == nil && tt.err {
			t.Fatalf("GetServiceClusterIP() should have returned error, but didn't")
		}
		if err == nil {
			if ip.String() != tt.expectedIP {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedIP, ip.String())
			}
		}
	}
}

func TestGetDNSIP(t *testing.T) {
	testData := []struct {
		serviceCIRD string
		expectedIP  string
		err         bool
	}{
		{"1111.0.0.1/12", "", true},
		{"10.96.0.0/24", "10.96.0.10", false},
	}

	for _, tt := range testData {
		ip, err := DNSIP(tt.serviceCIRD)
		if err != nil && !tt.err {
			t.Fatalf("GetDNSIP() err = %v", err)
		}
		if err == nil && tt.err {
			t.Fatalf("GetDNSIP() should have returned error, but didn't")
		}
		if err == nil {
			if ip.String() != tt.expectedIP {
				t.Fatalf("Expected '%s' but got '%s'", tt.expectedIP, ip.String())
			}
		}
	}
}

func TestAddToIP(t *testing.T) {
	v4 := net.ParseIP("192.168.0.1").To4()
	got, ok := addToIP(v4, 1)
	if !ok || got.String() != "192.168.0.2" {
		t.Fatalf("addToIP(v4,1) = (%v,%v), want (192.168.0.2,true)", got, ok)
	}

	maxV4 := net.ParseIP("255.255.255.255").To4()
	if got, ok := addToIP(maxV4, 1); ok || got != nil {
		t.Fatalf("addToIP(maxV4,1) = (%v,%v), want (nil,false)", got, ok)
	}

	v6 := net.ParseIP("fd00::").To16()
	got, ok = addToIP(v6, 0x10)
	if !ok || got.String() != "fd00::10" {
		t.Fatalf("addToIP(v6,0x10) = (%v,%v), want (fd00::10,true)", got, ok)
	}

	maxV6 := net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff").To16()
	if got, ok := addToIP(maxV6, 1); ok || got != nil {
		t.Fatalf("addToIP(maxV6,1) = (%v,%v), want (nil,false)", got, ok)
	}
}
