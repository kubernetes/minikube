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

	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
)

func (router *osRouter) EnsureRouteIsAdded(route *types.Route) error {
	exists, e := router.CheckRoute(route)
	if e != nil {
		return e
	}
	if exists {
		return nil
	}

	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	logrus.Infof("Adding Route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "Route", "-n", "add", serviceCIDR, gatewayIP)
	logrus.Infof("About to run command: %s", command.Args)
	stdInAndOut, e := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	re := regexp.MustCompile(fmt.Sprintf("add net (.*): gateway %s\n", gatewayIP))
	if !re.MatchString(message) {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	logrus.Infof("%s", stdInAndOut)
	if e != nil {
		return e
	}
	return nil
}

func (router *osRouter) CheckRoute(r *types.Route) (bool, error) {
	stdInAndOut, e := exec.Command("netstat", "-nr", "-f", "inet").CombinedOutput()
	if e != nil {
		return false, e
	}
	routeTableString := fmt.Sprintf("%s", stdInAndOut)
	routeTable := strings.Split(routeTableString, "\n")
	skip := true
	exactMatch := false
	collision := ""
	for _, line := range routeTable {
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
			dstCIDR := router.padCIDR(fields[0])
			gatewayIP := fields[1]
			if dstCIDR == r.DestCIDR.String() {
				if gatewayIP == r.Gateway.String() {
					exactMatch = true
				} else {
					collision = line
				}
			} else if ip, ipNet, e := net.ParseCIDR(dstCIDR); e == nil {
				if ipNet.Contains(r.DestCIDR.IP) || r.DestCIDR.Contains(ip) {
					logrus.Warningf("overlapping CIDR (%s) detected in routing table with minikube tunnel (%s). It is advisable to remove this rule. Run: sudo Route -n delete %s", ipNet, r.DestCIDR, ipNet)
				}
			} else {
				logrus.Errorf("can't parse CIDR from routing table: %s", dstCIDR)
			}
		}
	}

	if exactMatch {
		return true, nil
	}

	if len(collision) > 0 {
		return false, fmt.Errorf("conflicting rule in routing table: %s", collision)
	}

	return false, nil
}

func (r *osRouter) padCIDR(origCIDR string) string {
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

func (router *osRouter) Cleanup(route *types.Route) error {
	exists, e := router.CheckRoute(route)
	if e != nil {
		return e
	}
	if !exists {
		return nil
	}
	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	fmt.Printf("Cleaning up Route for CIDR %s to gateway %s\n", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "Route", "-n", "delete", serviceCIDR)
	if stdInAndOut, e := command.CombinedOutput(); e != nil {
		return e
	} else {
		message := fmt.Sprintf("%s", stdInAndOut)
		logrus.Infof("%s", message)
		re := regexp.MustCompile(fmt.Sprintf("^delete net ([^:]*)$"))
		if !re.MatchString(message) {
			return fmt.Errorf("error deleting Route: %s, %d", message, len(strings.Split(message, "\n")))
		} else {
			return nil
		}
	}
}
