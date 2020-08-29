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
		},
	})
	if err == nil {
		t.Errorf("add should have error, but it is nil")
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
	}

	err = r.EnsureRouteIsAdded(route)
	if err != nil {
		t.Errorf("add error: %s", err)
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
	}
	err = r.Cleanup(route)
	if err != nil {
		t.Errorf("cleanup failed: %s", err)
	}
}

func TestParseTable(t *testing.T) {
	const table = `default via 172.31.126.254 dev eno1 proto dhcp metric 100
10.96.0.0/12 via 192.168.39.47 dev virbr1
10.110.0.0/16 via 127.0.0.1 dev lo
172.31.126.0/24 dev eno1 proto kernel scope link src 172.31.126.54 metric 100
192.168.9.0/24 dev docker0 proto kernel scope link src 192.168.9.1`

	rt := (&osRouter{}).parseTable([]byte(table))

	expectedRt := routingTable{
		routingTableLine{
			route: unsafeParseRoute("192.168.39.47", "10.96.0.0/12"),
			line:  "10.96.0.0/12 via 192.168.39.47 dev virbr1",
		},
		routingTableLine{
			route: unsafeParseRoute("127.0.0.1", "10.110.0.0/16"),
			line:  "10.110.0.0/16 via 127.0.0.1 dev lo",
		},
		routingTableLine{
			route: unsafeParseRoute("0.0.0.0", "172.31.126.0/24"),
			line:  "172.31.126.0/24 dev eno1 proto kernel scope link src 172.31.126.54 metric 100",
		},
		routingTableLine{
			route: unsafeParseRoute("0.0.0.0", "192.168.9.0/24"),
			line:  "192.168.9.0/24 dev docker0 proto kernel scope link src 192.168.9.1",
		},
	}
	if !expectedRt.Equal(&rt) {
		t.Errorf("expected:\n %s\ngot\n %s", expectedRt.String(), rt.String())
	}
}

func addRoute(t *testing.T, cidr string, gw string) {
	command := exec.Command("sudo", "ip", "route", "add", cidr, "via", gw)
	sout, err := command.CombinedOutput()
	if err != nil {
		t.Logf("assertion add route error (should be ok): %s, error: %s", sout, err)
	} else {
		t.Logf("assertion - successfully added %s -> %s", cidr, gw)
	}
}

func cleanRoute(t *testing.T, cidr string) {
	command := exec.Command("sudo", "ip", "route", "delete", cidr)
	sout, err := command.CombinedOutput()
	if err != nil {
		t.Logf("integration test cleanup error (should be ok): %s, error: %s", sout, err)
	} else {
		t.Logf("integration test successfully cleaned %s", cidr)
	}
}
