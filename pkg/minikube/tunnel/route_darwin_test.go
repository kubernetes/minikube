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
	"k8s.io/minikube/pkg/minikube/tunnel/types"
)

func TestDarwinRouteFailsOnConflictIntegrationTest(t *testing.T) {
	cfg := &types.Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	addRoute(t, "10.96.0.0/12", "127.0.0.2")
	err := (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err == nil {
		t.Errorf("add should have error, but it is nil")
		t.Fail()
	}
}

func TestDarwinRouteIdempotentIntegrationTest(t *testing.T) {
	cfg := &types.Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	cleanRoute(t, "10.96.0.0/12")
	err := (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	err = (&osRouter{}).EnsureRouteIsAdded(cfg)
	if err != nil {
		t.Errorf("add error: %s", err)
		t.Fail()
	}

	cleanRoute(t, "10.96.0.0/12")
}

func TestDarwinRouteCleanupIdempontentIntegrationTest(t *testing.T) {

	cfg := &types.Route{
		Gateway: net.IPv4(127, 0, 0, 1),
		DestCIDR: &net.IPNet{
			IP:   net.IPv4(10, 96, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	cleanRoute(t, "10.96.0.0/12")
	addRoute(t, "10.96.0.0/12", "192.168.1.1")
	err := (&osRouter{}).Cleanup(cfg)
	if err != nil {
		t.Errorf("cleanup failed with %s", err)
		t.Fail()
	}
	err = (&osRouter{}).Cleanup(cfg)
	if err != nil {
		t.Errorf("cleanup failed with %s", err)
		t.Fail()
	}
}

func addRoute(t *testing.T, cidr string, gw string) {
	command := exec.Command("sudo", "Route", "-n", "add", cidr, gw)
	_, e := command.CombinedOutput()
	if e != nil {
		t.Logf("add Route error (should be ok): %s", e)
	}
}

func cleanRoute(t *testing.T, cidr string) {
	command := exec.Command("sudo", "Route", "-n", "delete", cidr)
	_, e := command.CombinedOutput()
	if e != nil {
		t.Logf("cleanup error (should be ok): %s", e)
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
				t.Fail()
			}
		})
	}
}
