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
	"reflect"
	"testing"

	"k8s.io/minikube/pkg/util"
)

func TestRoutingTable(t *testing.T) {
	tcs := []struct {
		name  string
		table routingTable
		route *Route

		exists   bool
		conflict string
		overlaps []string
	}{
		{
			name: "doesn't exist, no complication",
			table: routingTable{
				{
					route: unsafeParseRoute("127.0.0.1", "10.96.0.0/12"),
					line:  "line1",
				},
			},
			route: unsafeParseRoute("127.0.0.1", "10.112.0.0/12"),

			exists:   false,
			conflict: "",
			overlaps: []string{},
		},

		{
			name: "doesn't exist, and has overlap and a conflict",
			table: routingTable{
				{
					route: unsafeParseRoute("127.0.0.1", "10.96.0.0/12"),
					line:  "conflicting line",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "10.98.0.0/8"),
					line:  "overlap line1",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "10.100.0.0/24"),
					line:  "overlap line2",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "192.96.0.0/12"),
					line:  "no overlap",
				},
			},
			route: unsafeParseRoute("192.168.1.1", "10.96.0.0/12"),

			exists:   false,
			conflict: "conflicting line",
			overlaps: []string{
				"overlap line1",
				"overlap line2",
			},
		},

		{
			name: "exists, and has overlap and no conflict",
			table: routingTable{
				{
					route: unsafeParseRoute("127.0.0.1", "10.96.0.0/12"),
					line:  "same",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "10.98.0.0/8"),
					line:  "overlap line1",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "10.100.0.0/24"),
					line:  "overlap line2",
				},
				{
					route: unsafeParseRoute("127.0.0.1", "192.96.0.0/12"),
					line:  "no overlap",
				},
			},
			route: unsafeParseRoute("127.0.0.1", "10.96.0.0/12"),

			exists:   true,
			conflict: "",
			overlaps: []string{
				"overlap line1",
				"overlap line2",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			exists, conflict, overlaps := tc.table.Check(tc.route)
			if tc.exists != exists || tc.conflict != conflict || !reflect.DeepEqual(tc.overlaps, overlaps) {
				t.Errorf(`expected
  exists: %v
  conflict: %s
  overlaps: %s
got
  exists: %v
  conflict: %s
  overlaps: %s
`, tc.exists, tc.conflict, tc.overlaps,
					exists, conflict, overlaps)
			}
		})
	}
}

func unsafeParseRoute(gatewayIP string, destCIDR string) *Route {
	ip := net.ParseIP(gatewayIP)
	_, ipNet, _ := net.ParseCIDR(destCIDR)
	dnsIP, _ := util.GetDNSIP(ipNet.String())

	expectedRoute := &Route{
		Gateway:      ip,
		DestCIDR:     ipNet,
		ClusterDNSIP: dnsIP,
	}
	return expectedRoute
}
