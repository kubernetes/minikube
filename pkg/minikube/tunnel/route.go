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
	"github.com/golang/glog"
	"strings"
)

type router interface {
	Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error)
	EnsureRouteIsAdded(route *Route) error
	Cleanup(route *Route) error
}

type osRouter struct{}

type routingTableLine struct {
	route *Route
	line  string
}

func isValidToAddOrDelete(router router, r *Route) (bool, error) {
	exists, conflict, overlaps, e := router.Inspect(r)
	if e != nil {
		return false, e
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
	result := fmt.Sprintf("routingTable (%d routes)", len(*t))
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
