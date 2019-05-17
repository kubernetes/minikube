// +build darwin,integration

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

	"reflect"
)

func TestDarwinRouteFailsOnConflictIntegrationTest(t *testing.T) {
	cfg := &Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
		ClusterDomain: "cluster.local",
		ClusterDNSIP:  net.IPv4(10, 96, 0, 10),
	}
	conflictingCfg := *cfg
	conflictingCfg.Gateway = net.IPv4(127, 0, 0, 2)

	addRoute(t, &conflictingCfg)
	defer cleanRoute(t, &conflictingCfg)
	err := (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err == nil {
		t.Errorf("add should have error, but it is nil")
	}
}

func TestDarwinRouteIdempotentIntegrationTest(t *testing.T) {
	cfg := &Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
		ClusterDomain: "cluster.local",
		ClusterDNSIP:  net.IPv4(10, 96, 0, 10),
	}

	cleanRoute(t, cfg)
	err := (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err != nil {
		t.Errorf("add error: %s", err)
	}

	err = (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err != nil {
		t.Errorf("add error: %s", err)
	}

	cleanRoute(t, cfg)
}

func TestDarwinRouteCleanupIdempontentIntegrationTest(t *testing.T) {

	cfg := &Route{
		Gateway: net.IPv4(192, 168, 1, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
		ClusterDomain: "cluster.local",
		ClusterDNSIP:  net.IPv4(10, 96, 0, 10),
	}

	cleanRoute(t, cfg)
	addRoute(t, cfg)
	err := (&osRouter{}).Cleanup(cfg)
	if err != nil {
		t.Errorf("cleanup failed with %s", err)
	}
	err = (&osRouter{}).Cleanup(cfg)
	if err != nil {
		t.Errorf("cleanup failed with %s", err)
	}
}

func addRoute(t *testing.T, r *Route) {
	cidr := r.DestCIDR.String()
	gw := r.Gateway.String()
	command := exec.Command("sudo", "route", "-n", "add", cidr, gw)
	_, err := command.CombinedOutput()
	if err != nil {
		t.Logf("add route error (should be ok): %s", err)
	}
	err = writeResolverFile(r)
	if err != nil {
		t.Logf("add route DNS resolver error (should be ok): %s", err)
	}
}

func cleanRoute(t *testing.T, r *Route) {
	cidr := r.DestCIDR.String()
	command := exec.Command("sudo", "route", "-n", "delete", cidr)
	_, err := command.CombinedOutput()
	if err != nil {
		t.Logf("cleanup error (should be ok): %s", err)
	}
	command = exec.Command("sudo", "rm", "-f", fmt.Sprintf("/etc/resolver/%s", r.ClusterDomain))
	_, err = command.CombinedOutput()
	if err != nil {
		t.Logf("cleanup DNS resolver error (should be ok): %s", err)
	}
}

func TestCIDRPadding(t *testing.T) {
	testCases := []struct {
		inputCIDR  string
		paddedCIDR string
	}{
		{inputCIDR: "10", paddedCIDR: "10.0.0.0/8"},
		{inputCIDR: "10.96/12", paddedCIDR: "10.96.0.0/12"},
		{inputCIDR: "192.168.43", paddedCIDR: "192.168.43.0/24"},
		{inputCIDR: "192.168.43.1/32", paddedCIDR: "192.168.43.1/32"},
		{inputCIDR: "127.0.0.1", paddedCIDR: "127.0.0.1/32"},
	}

	for _, test := range testCases {
		testName := fmt.Sprintf("pad(%s) should be %s", test.inputCIDR, test.paddedCIDR)
		t.Run(testName, func(t *testing.T) {
			cidr := (&osRouter{}).padCIDR(test.inputCIDR)
			if cidr != test.paddedCIDR {
				t.Errorf("%s got %s", testName, cidr)
			}
		})
	}
}

func TestRoutingTableParser(t *testing.T) {
	table := `Routing tables

Internet:
Destination        Gateway            Flags        Refs      Use   Netif Expire
127                127.0.0.1          UCS             0        0     lo0
127.0.0.1          127.0.0.1          UH             13    30917     lo0
172.16.128/24      link#17            UC              1        0  vmnet1
192.168.246        link#18            UC              1        0  vmnet8
224.0.0            link#1             UmCS            0        0     lo0
`
	rt := (&osRouter{}).parseTable([]byte(table))

	expectedRt := routingTable{
		routingTableLine{
			route: unsafeParseRoute("127.0.0.1", "127.0.0.0/8"),
			line:  "127                127.0.0.1          UCS             0        0     lo0",
		},
		routingTableLine{
			route: unsafeParseRoute("127.0.0.1", "127.0.0.1/32"),
			line:  "127.0.0.1          127.0.0.1          UH             13    30917     lo0",
		},
	}
	if !reflect.DeepEqual(rt, expectedRt) {
		t.Errorf("expected:\n %s\ngot\n %s", expectedRt.String(), rt.String())
	}
}
