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
)

func TestLinuxRouteFailsOnConflictIntegrationTest(t *testing.T) {
	r := &osRouter{}

	cleanRoute(t, "10.96.0.0/12")
	addRoute(t, "10.96.0.0/12", "127.0.0.1")
	err := r.EnsureRouteIsAdded(&Route{
		Gateway: net.IPv4(127, 0, 0, 2),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		}})
	if err == nil {
		t.Errorf("add should have error, but it is nil")
		t.Fail()
	}
	cleanRoute(t, "10.96.0.0/12")
}

func TestLinuxRouteIdempotentIntegrationTest(t *testing.T) {
	r := &osRouter{}

	cleanRoute(t, "10.96.0.0/12")
	route := &Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}
	err := r.EnsureRouteIsAdded(route)
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	err = r.EnsureRouteIsAdded(route)
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	cleanRoute(t, "10.96.0.0/12")
}

func TestLinuxRouteCleanupIdempontentIntegrationTest(t *testing.T) {

	r := &osRouter{}
	route := &Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	cleanRoute(t, "10.96.0.0/12")
	addRoute(t, "10.96.0.0/12", "127.0.0.1")
	err := r.Cleanup(route)
	if err != nil {
		t.Errorf("cleanup failed: %s", err)
		t.Fail()
	}
	err = r.Cleanup(route)
	if err != nil {
		t.Errorf("cleanup failed: %s", err)
		t.Fail()
	}
}

func TestParseTable(t *testing.T) {

	const table = `Kernel IP routing table
Destination     Gateway         Genmask         Flags   MSS Window  irtt Iface
0.0.0.0         172.31.126.254  0.0.0.0         UG        0 0          0 eno1
10.96.0.0       127.0.0.1   		255.240.0.0			UG        0 0          0 eno1
172.31.126.0    0.0.0.0         255.255.255.0   U         0 0          0 eno1
`

	rt := (&osRouter{}).parseTable(table)

	expectedRt := routingTable{
		routingTableLine{
			route: unsafeParseRoute("127.0.0.1", "10.96.0.0/12"),
			line: "10.96.0.0       127.0.0.1   		255.240.0.0			UG        0 0          0 eno1",
		},
		routingTableLine{
			route: unsafeParseRoute("0.0.0.0", "172.31.126.0/24"),
			line:  "172.31.126.0    0.0.0.0         255.255.255.0   U         0 0          0 eno1",
		},
	}
	if !expectedRt.Equal(&rt) {
		t.Errorf("expected:\n %s\ngot\n %s", expectedRt.String(), rt.String())
	}
}

func addRoute(t *testing.T, cidr string, gw string) {
	command := exec.Command("sudo", "ip", "route", "add", cidr, "via", gw)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("assertion add Route error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("assertion - successfully added %s -> %s", cidr, gw)
	}
}

func cleanRoute(t *testing.T, cidr string) {
	command := exec.Command("sudo", "ip", "route", "delete", cidr)
	sout, e := command.CombinedOutput()
	if e != nil {
		t.Logf("integration test cleanup error (should be ok): %s, error: %s", sout, e)
	} else {
		t.Logf("integration test successfully cleaned %s", cidr)
	}
}
