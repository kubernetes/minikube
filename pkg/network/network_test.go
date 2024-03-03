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

	"github.com/juju/mutex/v2"
)

func TestFreeSubnet(t *testing.T) {
	reserveSubnet = func(_ string) (mutex.Releaser, error) { return nil, nil }

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
		originalIsSubnetTaken := isSubnetTaken
		defer func() {
			isSubnetTaken = originalIsSubnetTaken
		}()

		isSubnetTaken = func(_ string) (bool, error) {
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
		originalInspect := Inspect
		defer func() {
			Inspect = originalInspect
		}()

		Inspect = func(addr string) (*Parameters, error) {
			count++
			p := &Parameters{IP: addr, IsPrivate: true}
			if count == 1 {
				p.IP = "0.0.0.0"
				p.IsPrivate = false
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

func TestParseAddr(t *testing.T) {
	t.Run("ValidIP", func(t *testing.T) {
		addr := "192.168.9.0"
		ip, _, err := ParseAddr(addr)
		if err != nil {
			t.Fatal(err)
		}
		if ip.String() != addr {
			t.Errorf("expected IP = %q; got = %q", addr, ip.String())
		}
	})

	t.Run("ValidCIDR", func(t *testing.T) {
		addr := "192.168.9.0/30"
		ip, cidr, err := ParseAddr(addr)
		if err != nil {
			t.Fatal(err)
		}

		expected := "192.168.9.0"
		if ip.String() != expected {
			t.Errorf("expected IP = %q; got = %q", expected, ip.String())
		}

		mask, _ := cidr.Mask.Size()
		expectedMask := 30
		if mask != expectedMask {
			t.Errorf("expected mask = %q; got = %q", mask, expectedMask)
		}
	})

	t.Run("InvalidAddr", func(t *testing.T) {
		tests := []string{
			"192.168.9",
			"192.168.9.0/30000",
		}
		for _, test := range tests {
			_, _, err := ParseAddr(test)
			if err == nil {
				t.Fatalf("expected to fail since address is invalid")
			}
		}
	})
}
