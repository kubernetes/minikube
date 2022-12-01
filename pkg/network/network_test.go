/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package network

import (
	"strings"
	"testing"
	"time"
)

func TestIsSubnetPrivate(t *testing.T) {
	tests := []struct {
		subnet   string
		expected bool
	}{
		{"9.255.255.255", false},
		{"10.0.0.0", true},
		{"10.255.255.255", true},
		{"11.0.0.0", false},
		{"172.15.255.255", false},
		{"172.16.0.0", true},
		{"172.31.255.255", true},
		{"172.32.0.0", false},
		{"192.167.255.255", false},
		{"192.168.0.0", true},
		{"192.168.255.255", true},
		{"192.169.0.0", false},
	}
	for _, test := range tests {
		got := isSubnetPrivate(test.subnet)
		if got != test.expected {
			t.Errorf("isSubnetPrivate(%q) = %t; expected = %t", test.subnet, got, test.expected)
		}
	}
}

func TestFreeSubnet(t *testing.T) {
	reserveSubnet = func(subnet string, period time.Duration) bool { return true }

	t.Run("NoRetriesSuccess", func(t *testing.T) {
		startingSubnet := "192.168.0.0"
		subnet, err := FreeSubnet(startingSubnet, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		expectedIP := startingSubnet
		if subnet.IP != expectedIP {
			t.Errorf("expected IP = %q; got = %q", expectedIP, subnet.IP)
		}
	})

	t.Run("FirstSubnetTaken", func(t *testing.T) {
		count := 0
		isSubnetTaken = func(subnet string) (bool, error) {
			count++
			return count == 1, nil
		}

		startingSubnet := "192.168.0.0"
		subnet, err := FreeSubnet(startingSubnet, 9, 2)
		if err != nil {
			t.Fatal(err)
		}
		expectedIP := "192.168.9.0"
		if subnet.IP != expectedIP {
			t.Errorf("expected IP = %q; got = %q", expectedIP, subnet.IP)
		}
	})

	t.Run("FirstSubnetIPV6NetworkFound", func(t *testing.T) {
		count := 0
		inspect = func(addr string) (*Parameters, error) {
			count++
			p := &Parameters{IP: addr}
			if count == 1 {
				p.IP = "0.0.0.0"
			}
			return p, nil
		}

		startingSubnet := "10.0.0.0"
		subnet, err := FreeSubnet(startingSubnet, 9, 2)
		if err != nil {
			t.Fatal(err)
		}
		expectedIP := "10.9.0.0"
		if subnet.IP != expectedIP {
			t.Errorf("expepcted IP = %q; got = %q", expectedIP, subnet.IP)
		}
	})

	t.Run("NonPrivateSubnet", func(t *testing.T) {
		startingSubnet := "192.167.0.0"
		_, err := FreeSubnet(startingSubnet, 9, 1)
		if err == nil {
			t.Fatalf("expected to fail since IP non-private but no error thrown")
		}
		if !strings.Contains(err.Error(), startingSubnet) {
			t.Errorf("expected starting subnet of %q to be included in error, but intead got: %v", startingSubnet, err)
		}
	})
}
