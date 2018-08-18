// +build windows,integration

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"net"
	"os/exec"
	"testing"

	"fmt"

	"strings"
)

func TestWindowsRouteFailsOnConflictIntegrationTest(t *testing.T) {
	r := &OSRouter{types.Route{
		Gateway: net.IPv4(1, 2, 3, 4),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}}

	cleanRoute(t, "10.96.0.0")
	addRoute(t, "10.96.0.0", "255.240.0.0", "1.2.3.5")
	err := r.EnsureRouteIsAdded()
	if err == nil {
		t.Errorf("add should have error, but it is nil")
		t.Fail()
	} else if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("expected to fail with error containg `conflict`, but failed with wrong error %s", err)
		t.Fail()
	}
	cleanRoute(t, "10.96.0.0")
}

func TestWindowsRouteIdempotentIntegrationTest(t *testing.T) {
	r := &OSRouter{types.Route{
		Gateway: net.IPv4(1, 2, 3, 4),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}}

	cleanRoute(t, "10.96.0.0")
	err := r.EnsureRouteIsAdded()
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	err = r.EnsureRouteIsAdded()
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	cleanRoute(t, "10.96.0.0")
}

func TestWindowsRouteCleanupIdempontentIntegrationTest(t *testing.T) {

	r := &OSRouter{types.Route{
		Gateway: net.IPv4(1, 2, 3, 4),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}}

	cleanRoute(t, "10.96.0.0")
	addRoute(t, "10.96.0.0", "255.240.0.0", "1.2.3.4")
	err := r.Cleanup()
	if err != nil {
		t.Errorf("cleanup failed: %s", err)
		t.Fail()
	}
	err = r.Cleanup()
	if err != nil {
		t.Errorf("cleanup failed: %s", err)
		t.Fail()
	}
	cleanRoute(t, "10.96.0.0")
}

func TestRouteTable(t *testing.T) {
	testCases := []struct {
		name           string
		cidr           *net.IPNet
		gateway        net.IP
		expectedResult bool
		expectedError  error
	}{
		{
			name: "Route already exists",
			cidr: &net.IPNet{
				IP:   net.IPv4(10, 96, 0, 0),
				Mask: net.IPv4Mask(255, 240, 0, 0),
			},
			gateway:        net.IPv4(127, 0, 0, 1),
			expectedError:  nil,
			expectedResult: true,
		},

		{
			name: "destination exists but conflicting gateway",
			cidr: &net.IPNet{
				IP:   net.IPv4(10, 96, 0, 0),
				Mask: net.IPv4Mask(255, 240, 0, 0),
			},
			gateway: net.IPv4(127, 0, 0, 2),
			expectedError: fmt.Errorf("conflicting rule in routing table: %s", " 	    10.96.0.0      255.240.0.0        127.0.0.1        127.0.0.1    281"),
			expectedResult: false,
		},

		{
			name: "Route doesn't exist yet",
			cidr: &net.IPNet{
				IP:   net.IPv4(10, 112, 0, 0),
				Mask: net.IPv4Mask(255, 240, 0, 0),
			},
			gateway:        net.IPv4(127, 0, 0, 1),
			expectedError:  nil,
			expectedResult: false,
		},

		{
			name: "Route doesn't exist yet, but there is overlap (warning is only logged)",
			cidr: &net.IPNet{
				IP:   net.IPv4(10, 0, 0, 0),
				Mask: net.IPv4Mask(255, 0, 0, 0),
			},
			gateway:        net.IPv4(127, 0, 0, 1),
			expectedError:  nil,
			expectedResult: false,
		},

		{
			name: "Route doesn't exist yet, but there is overlap (warning is only logged)",
			cidr: &net.IPNet{
				IP:   net.IPv4(10, 96, 1, 0),
				Mask: net.IPv4Mask(255, 255, 0, 0),
			},
			gateway:        net.IPv4(127, 0, 0, 1),
			expectedError:  nil,
			expectedResult: false,
		},
	}

	const table = `===========================================================================
Interface List
 14...00 1c 42 8f 70 58 ......Intel(R) 82574L Gigabit Network Connection
  6...0a 00 27 00 00 06 ......VirtualBox Host-Only Ethernet Adapter
  8...0a 00 27 00 00 08 ......VirtualBox Host-Only Ethernet Adapter #2
  1...........................Software Loopback Interface 1
===========================================================================

IPv4 Route Table
===========================================================================
Active Routes:
Network Destination        Netmask          Gateway       Interface  Metric
          0.0.0.0          0.0.0.0      10.211.55.1      10.211.55.3     25
 	    10.96.0.0      255.240.0.0        127.0.0.1        127.0.0.1    281
      10.211.55.0    255.255.255.0         On-link       10.211.55.3    281
      10.211.55.3  255.255.255.255         On-link       10.211.55.3    281
    10.211.55.255  255.255.255.255         On-link       10.211.55.3    281
        127.0.0.0        255.0.0.0         On-link         127.0.0.1    331
        127.0.0.1  255.255.255.255         On-link         127.0.0.1    331
  127.255.255.255  255.255.255.255         On-link         127.0.0.1    331
     192.168.56.0    255.255.255.0         On-link      192.168.56.1    281
     192.168.56.1  255.255.255.255         On-link      192.168.56.1    281
   192.168.56.255  255.255.255.255         On-link      192.168.56.1    281
     192.168.99.0    255.255.255.0         On-link      192.168.99.1    281
     192.168.99.1  255.255.255.255         On-link      192.168.99.1    281
   192.168.99.255  255.255.255.255         On-link      192.168.99.1    281
        224.0.0.0        240.0.0.0         On-link         127.0.0.1    331
        224.0.0.0        240.0.0.0         On-link       10.211.55.3    281
        224.0.0.0        240.0.0.0         On-link      192.168.56.1    281
        224.0.0.0        240.0.0.0         On-link      192.168.99.1    281
  255.255.255.255  255.255.255.255         On-link         127.0.0.1    331
  255.255.255.255  255.255.255.255         On-link       10.211.55.3    281
  255.255.255.255  255.255.255.255         On-link      192.168.56.1    281
  255.255.255.255  255.255.255.255         On-link      192.168.99.1    281
===========================================================================
Persistent Routes:
  None`

	for _, testCase := range testCases {
		result, e := checkRouteTable(testCase.cidr, testCase.gateway, table)
		errorsEqual := strings.Compare(fmt.Sprintf("%s", e), fmt.Sprintf("%s", testCase.expectedError)) == 0
		if !errorsEqual || result != testCase.expectedResult {
			t.Errorf(`[%s] failed. 
expected 	"%v" | error: [%s]
got 			"%v" | error: [%s]`, testCase.name, testCase.expectedResult, testCase.expectedError, result, e)
			t.Fail()
		}
	}

}

func addRoute(t *testing.T, dstIP string, dstMask string, gw string) {
	command := exec.Command("Route", "ADD", dstIP, "mask", dstMask, gw)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("assertion add Route error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("assertion - successfully added %s (%s) -> %s", dstIP, dstMask, gw)
	}
}

func cleanRoute(t *testing.T, dstIP string) {
	command := exec.Command("Route", "DELETE", dstIP)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("assertion cleanup error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("assertion - successfully cleaned %s", dstIP)
	}
}
