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

	"k8s.io/klog/v2"
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

	klog.Infof("Adding route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "ip", "route", "add", serviceCIDR, "via", gatewayIP)
	klog.Infof("About to run command: %s", command.Args)
	stdInAndOut, err := command.CombinedOutput()
	message := string(stdInAndOut)
	if len(message) > 0 {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	klog.Info(stdInAndOut)
	if err != nil {
		klog.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
		return err
	}
	return nil
}

func (router *osRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	cmd := exec.Command("ip", "r")
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
	for _, line := range strings.Split(string(table), "\n") {

		fields := strings.Fields(line)

		// don't care about the routes that 0.0.0.0
		if len(fields) == 0 ||
			len(fields) > 0 && (fields[0] == "default" || fields[0] == "0.0.0.0") {
			continue
		}

		if len(fields) > 2 {

			// assuming "10.96.0.0/12 via 192.168.39.47 dev virbr1"
			dstCIDRString := fields[0]
			gatewayIPString := fields[2]
			gatewayIP := net.ParseIP(gatewayIPString)

			// if not via format, then gateway is assumed to be 0.0.0.0
			// "1.2.3.0/24 dev eno1 proto kernel scope link src 1.2.3.54 metric 100"
			if fields[1] != "via" {
				gatewayIP = net.ParseIP("0.0.0.0")
			}

			_, ipNet, err := net.ParseCIDR(dstCIDRString)
			if err != nil {
				klog.V(4).Infof("skipping line: can't parse CIDR from routing table: %s", dstCIDRString)
			} else if gatewayIP == nil {
				klog.V(4).Infof("skipping line: can't parse IP from routing table: %s", gatewayIPString)
			} else {

				tableLine := routingTableLine{
					route: &Route{
						DestCIDR: ipNet,
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

	klog.Infof("Cleaning up route for CIDR %s to gateway %s\n", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "ip", "route", "delete", serviceCIDR)
	stdInAndOut, err := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	klog.Infof("%s", message)
	if err != nil {
		return fmt.Errorf("error deleting Route: %s, %s", message, err)
	}
	return nil
}
