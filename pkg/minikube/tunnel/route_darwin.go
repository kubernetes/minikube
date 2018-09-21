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
	"regexp"
	"strings"

	"github.com/golang/glog"
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
	gatewayIP := route.Gateway.String()

	glog.Infof("Adding Route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "route", "-n", "add", serviceCIDR, gatewayIP)
	glog.Infof("About to run command: %s", command.Args)
	stdInAndOut, e := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	re := regexp.MustCompile(fmt.Sprintf("add net (.*): gateway %s\n", gatewayIP))
	if !re.MatchString(message) {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	glog.Infof("%s", stdInAndOut)
	if e != nil {
		return e
	}
	return nil
}

func (router *osRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	command := exec.Command("netstat", "-nr", "-f", "inet")
	stdInAndOut, e := command.CombinedOutput()
	if e != nil {
		err = fmt.Errorf("error running '%v': %s", command, e)
		return
	}
	routeTableString := fmt.Sprintf("%s", stdInAndOut)

	rt := router.parseTable(routeTableString)

	exists, conflict, overlaps = rt.Check(route)

	return
}

func (router *osRouter) parseTable(table string) routingTable {
	t := routingTable{}
	skip := true
	for _, line := range strings.Split(table, "\n") {
		//header
		if strings.HasPrefix(line, "Destination") {
			skip = false
			continue
		}
		//don't care about the 0.0.0.0 routes
		if skip || strings.HasPrefix(line, "default") {
			continue
		}
		fields := strings.Fields(line)

		if len(fields) > 2 {
			dstCIDRString := router.padCIDR(fields[0])
			gatewayIPString := fields[1]
			gatewayIP := net.ParseIP(gatewayIPString)

			_, ipNet, e := net.ParseCIDR(dstCIDRString)
			if e != nil {
				glog.V(4).Infof("skipping line: can't parse CIDR from routing table: %s", dstCIDRString)
			} else if gatewayIP == nil {
				glog.V(4).Infof("skipping line: can't parse IP from routing table: %s", gatewayIPString)
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

func (router *osRouter) padCIDR(origCIDR string) string {
	s := ""
	dots := 0
	slash := false
	for i, c := range origCIDR {
		if c == '.' {
			dots++
		}
		if c == '/' {
			for dots < 3 {
				s += ".0"
				dots++
			}
			slash = true
		}
		if i == len(origCIDR)-1 {
			s += string(c)
			bits := 32 - 8*(3-dots)
			for dots < 3 {
				s += ".0"
				dots++
			}
			if !slash {
				s += fmt.Sprintf("/%d", bits)
			}
		} else {
			s += string(c)
		}
	}
	return s
}

func (router *osRouter) Cleanup(route *Route) error {
	glog.V(3).Infof("Cleaning up %s\n", route)
	exists, e := isValidToAddOrDelete(router, route)
	if e != nil {
		return e
	}
	if !exists {
		return nil
	}
	command := exec.Command("sudo", "route", "-n", "delete", route.DestCIDR.String())
	if stdInAndOut, e := command.CombinedOutput(); e != nil {
		return e
	} else {
		message := fmt.Sprintf("%s", stdInAndOut)
		glog.V(4).Infof("%s", message)
		re := regexp.MustCompile(fmt.Sprintf("^delete net ([^:]*)$"))
		if !re.MatchString(message) {
			return fmt.Errorf("error deleting route: %s, %d", message, len(strings.Split(message, "\n")))
		} else {
			return nil
		}
	}
}
