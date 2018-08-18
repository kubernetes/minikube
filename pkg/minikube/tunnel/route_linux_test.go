// +build linux,integration

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

func TestLinuxRouteFailsOnConflictIntegrationTest(t *testing.T) {
	r := &OSRouter{types.Route{
		Gateway: net.IPv4(127, 0, 0, 2),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}}

	cleanRoute(t, "10.96.0.0/12")
	addRoute(t, "10.96.0.0/12", "127.0.0.1")
	err := r.EnsureRouteIsAdded()
	if err == nil {
		t.Errorf("add should have error, but it is nil")
		t.Fail()
	}
	cleanRoute(t, "10.96.0.0/12")
}

func TestLinuxRouteIdempotentIntegrationTest(t *testing.T) {
	r := newRouter(types.Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	})

	cleanRoute(t, "10.96.0.0/12")
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

	cleanRoute(t, "10.96.0.0/12")
}

func TestLinuxRouteCleanupIdempontentIntegrationTest(t *testing.T) {

	r := newRouter(types.Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	})

	cleanRoute(t, "10.96.0.0/12")
	addRoute(t, "10.96.0.0/12", "127.0.0.1")
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
			expectedError: fmt.Errorf("conflicting rule in routing table: %s", "10.96.0.0       127.0.0.1   		255.240.0.0			UG        0 0          0 eno1"),
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

	const table = `Kernel IP routing table
Destination     Gateway         Genmask         Flags   MSS Window  irtt Iface
0.0.0.0         172.31.126.254  0.0.0.0         UG        0 0          0 eno1
10.96.0.0       127.0.0.1   		255.240.0.0			UG        0 0          0 eno1
172.31.126.0    0.0.0.0         255.255.255.0   U         0 0          0 eno1
192.168.9.0     0.0.0.0         255.255.255.0   U         0 0          0 docker0
192.168.39.0    0.0.0.0         255.255.255.0   U         0 0          0 virbr1
192.168.122.0   0.0.0.0         255.255.255.0   U         0 0          0 virbr0
`

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

func addRoute(t *testing.T, cidr string, gw string) {
	command := exec.Command("sudo", "ip", "Route", "add", cidr, "via", gw)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("assertion add Route error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("assertion - successfully added %s -> %s", cidr, gw)
	}
}

func cleanRoute(t *testing.T, cidr string) {
	command := exec.Command("sudo", "ip", "Route", "delete", cidr)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("assertion cleanup error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("assertion - successfully cleaned %s", cidr)
	}
}
