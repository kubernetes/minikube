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
	"net"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

func (router *osRouter) EnsureRouteIsAdded(route *Route) error {
	exists, err := isValidToAddOrDelete(router, route)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	glog.Infof("Adding route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "ip", "route", "add", serviceCIDR, "via", gatewayIP)
	glog.Infof("About to run command: %s", command.Args)
	stdInAndOut, err := command.CombinedOutput()
	message := string(stdInAndOut)
	if len(message) > 0 {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	glog.Info(stdInAndOut)
	if err != nil {
		glog.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
		return err
	}
	return nil
}

func (router *osRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	cmd := exec.Command("route", "-n")
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	stdInAndOut, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("error running '%v': %s", cmd, err)
		return
	}
	rt := router.parseTable(stdInAndOut)

	exists, conflict, overlaps = rt.Check(route)

	return
}

func (router *osRouter) parseTable(table []byte) routingTable {
	t := routingTable{}
	skip := true
	for _, line := range strings.Split(string(table), "\n") {
		//after first line of header we can start consuming
		if strings.HasPrefix(line, "Destination") {
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
			dstCIDRMask := fields[2]
			dstMaskIP := net.ParseIP(dstCIDRMask)
			gatewayIP := net.ParseIP(fields[1])

			if dstCIDRIP == nil || gatewayIP == nil || dstMaskIP == nil {
				glog.V(8).Infof("skipping line: can't parse: %s", line)
			} else {

				dstCIDR := &net.IPNet{
					IP:   dstCIDRIP,
					Mask: net.IPv4Mask(dstMaskIP[12], dstMaskIP[13], dstMaskIP[14], dstMaskIP[15]),
				}

				tableLine := routingTableLine{
					route: &Route{
						DestCIDR: dstCIDR,
						Gateway:  gatewayIP,
					},
					line: line,
				}
				t = append(t, tableLine)
			}
		}
	}

	return t
}

func (router *osRouter) Cleanup(route *Route) error {
	exists, err := isValidToAddOrDelete(router, route)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	glog.Infof("Cleaning up route for CIDR %s to gateway %s\n", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "ip", "route", "delete", serviceCIDR)
	stdInAndOut, err := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	glog.Infof("%s", message)
	if err != nil {
		return fmt.Errorf("error deleting Route: %s, %s", message, err)
	}
	return nil
}
