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
