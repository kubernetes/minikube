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
	"fmt"
	"strings"

	"github.com/golang/glog"
)

//router manages the routing table on the host, implementations should cater for OS specific methods
type router interface {
	//Inspect checks if the given route exists or not in the routing table
	//conflict is defined as: same destination CIDR, different Gateway
	//overlaps are defined as: routes that have overlapping but not exactly matching destination CIDR
	Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error)

	//EnsureRouteIsAdded is an idempotent way to add a route to the routing table
	//it fails if there is a conflict
	EnsureRouteIsAdded(route *Route) error

	//Cleanup is an idempotent way to remove a route from the routing table
	//it fails if there is a conflict
	Cleanup(route *Route) error
}

type osRouter struct{}

type routingTableLine struct {
	route *Route
	line  string
}

func isValidToAddOrDelete(router router, r *Route) (bool, error) {
	exists, conflict, overlaps, err := router.Inspect(r)
	if err != nil {
		return false, err
	}

	if len(overlaps) > 0 {
		glog.Warningf("overlapping CIDR detected in routing table with minikube tunnel (CIDR: %s). It is advisable to remove these rules:\n%v", r.DestCIDR, strings.Join(overlaps, "\n"))
	}

	if exists {
		return true, nil
	}

	if len(conflict) > 0 {
		return false, fmt.Errorf("conflicting rule in routing table: %s", conflict)
	}

	return false, nil
}

//a partial representation of the routing table on the host
//tunnel only requires the destination CIDR, the gateway and the actual textual representation per line
type routingTable []routingTableLine

func (t *routingTable) Check(route *Route) (exists bool, conflict string, overlaps []string) {
	conflict = ""
	exists = false
	overlaps = []string{}
	for _, tableLine := range *t {

		if route.Equal(tableLine.route) {
			exists = true
		} else if route.DestCIDR.String() == tableLine.route.DestCIDR.String() &&
			route.Gateway.String() != tableLine.route.Gateway.String() {
			conflict = tableLine.line
		} else if route.DestCIDR.Contains(tableLine.route.DestCIDR.IP) || tableLine.route.DestCIDR.Contains(route.DestCIDR.IP) {
			overlaps = append(overlaps, tableLine.line)
		}
	}
	return
}

func (t *routingTable) String() string {
	result := fmt.Sprintf("table (%d routes)", len(*t))
	for _, l := range *t {
		result = fmt.Sprintf("%s\n  %s\t|%s", result, l.route.String(), l.line)
	}
	return result
}

func (t *routingTable) Equal(other *routingTable) bool {
	if other == nil || len(*t) != len(*other) {
		return false
	}

	for i := range *t {
		routesEqual := (*t)[i].route.Equal((*other)[i].route)
		linesEqual := (*t)[i].line == ((*other)[i].line)
		if !(routesEqual && linesEqual) {
			return false
		}
	}
	return true
}
