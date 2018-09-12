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
	"net"
	"os/exec"
	"strings"
)

func (router *osRouter) EnsureRouteIsAdded(route *Route) error {
	exists, e := isValidToAddOrDelete(router, route)
	if e != nil {
		return e
	}
	if exists {
		return nil
	}

	serviceCIDR := route.DestCIDR.String()
	destinationIP := route.DestCIDR.IP.String()
	destinationMask := fmt.Sprintf("%d.%d.%d.%d",
		route.DestCIDR.Mask[0],
		route.DestCIDR.Mask[1],
		route.DestCIDR.Mask[2],
		route.DestCIDR.Mask[3])

	gatewayIP := route.Gateway.String()

	glog.Infof("Adding Route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("Route", "ADD", destinationIP, "MASK", destinationMask, gatewayIP)
	glog.Infof("About to run command: %s", command.Args)
	stdInAndOut, e := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	if message != " OK!\r\n" {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	glog.Infof("%s", stdInAndOut)
	if e != nil {
		glog.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
		return e
	}
	return nil
}

func (router *osRouter) parseTable(table string) routingTable {
	t := routingTable{}
	skip := true
	for _, line := range strings.Split(table, "\n") {
		//after first line of header we can start consuming
		if strings.HasPrefix(line, "Network Destination") {
			skip = false
			continue
		}

		fields := strings.Fields(line)
		//don't care about the 0.0.0.0 routes
		if skip || len(fields) == 0 || len(fields) > 0 && (fields[0] == "default" || fields[0] == "0.0.0.0") {
			continue
		}
		if len(fields) > 2 {
			dstCIDRIP := net.ParseIP(fields[0])
			dstCIDRMask := fields[1]
			dstMaskIP := net.ParseIP(dstCIDRMask)
			gatewayIP := net.ParseIP(fields[2])
			if dstCIDRIP == nil || dstMaskIP == nil || gatewayIP == nil {
				glog.V(4).Infof("skipping line: can't parse all IPs from routing table: %s", line)
			} else {
				tableLine := routingTableLine{
					route: &Route{
						DestCIDR: &net.IPNet{
							IP:   dstCIDRIP,
							Mask: net.IPMask(dstMaskIP.To4()),
						},
						Gateway: gatewayIP,
					},
					line: line,
				}
				glog.V(4).Infof("adding line %s", tableLine)
				t = append(t, tableLine)
			}
		}
	}

	return t
}

func (router *osRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	command := exec.Command("Route", "print", "-4")
	stdInAndOut, e := command.CombinedOutput()
	if e != nil {
		err = fmt.Errorf("error running '%s': %s", command.Args, e)
		return
	}
	routeTableString := fmt.Sprintf("%s", stdInAndOut)

	rt := router.parseTable(routeTableString)

	exists, conflict, overlaps = rt.Check(route)

	return
}

func (router *osRouter) Cleanup(route *Route) error {
	exists, e := isValidToAddOrDelete(router, route)
	if e != nil {
		return e
	}
	if !exists {
		return nil
	}
	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	fmt.Printf("Cleaning up Route for CIDR %s to gateway %s\n", serviceCIDR, gatewayIP)
	command := exec.Command("Route", "delete", serviceCIDR)
	if stdInAndOut, e := command.CombinedOutput(); e != nil {
		return e
	} else {
		message := fmt.Sprintf("%s", stdInAndOut)
		glog.Infof("'%s'", message)
		if message != " OK!\r\n" {
			return fmt.Errorf("error deleting Route: %s, %d", message, len(strings.Split(message, "\n")))
		} else {
			return nil
		}
	}
}
