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

	"github.com/sirupsen/logrus"

)

type OSRouter struct {
	config Route
}

func (r *OSRouter) EnsureRouteIsAdded() error {
	exists, e := checkRoute(r.config.RoutedCIDR, r.config.TargetGateway)
	if e != nil {
		return e
	}
	if exists {
		return nil
	}

	serviceCIDR := r.config.RoutedCIDR.String()
	destinationIP := r.config.RoutedCIDR.IP.String()
	destinationMask := fmt.Sprintf("%d.%d.%d.%d",
		r.config.RoutedCIDR.Mask[0],
		r.config.RoutedCIDR.Mask[1],
		r.config.RoutedCIDR.Mask[2],
		r.config.RoutedCIDR.Mask[3])

	gatewayIP := r.config.TargetGateway.String()

	logrus.Infof("Adding Route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("Route", "ADD", destinationIP, "MASK", destinationMask, gatewayIP)
	logrus.Infof("About to run command: %s", command.Args)
	stdInAndOut, e := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	if message != " OK!\r\n" {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	logrus.Infof("%s", stdInAndOut)
	if e != nil {
		logrus.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
		return e
	}
	return nil
}

func checkRoute(cidr *net.IPNet, gateway net.IP) (bool, error) {
	stdInAndOut, e := exec.Command("Route", "print", "-4").CombinedOutput()
	if e != nil {
		return false, e
	}
	routeTableString := fmt.Sprintf("%s", stdInAndOut)
	return checkRouteTable(cidr, gateway, routeTableString)
}

func checkRouteTable(cidr *net.IPNet, gateway net.IP, routeTableString string) (bool, error) {
	routeTable := strings.Split(routeTableString, "\n")
	skip := true
	exactMatch := false
	collision := ""
	for _, line := range routeTable {
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
			gatewayIP := fields[2]
			dstCIDR := &net.IPNet{
				IP:   dstCIDRIP,
				Mask: net.IPMask(dstMaskIP.To4()),
			}
			if dstCIDR.String() == cidr.String() {
				if gatewayIP == gateway.String() {
					exactMatch = true
				} else {
					collision = line
				}
			} else if dstCIDR != nil {
				if dstCIDR.Contains(cidr.IP) || cidr.Contains(dstCIDRIP) {
					logrus.Warningf("overlapping CIDR (%s) detected in routing table with minikube tunnel (%s). It is advisable to remove this rule. Run: sudo Route -n delete %s", dstCIDR.String(), cidr, dstCIDR.String())
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

func (r *OSRouter) Cleanup() error {
	exists, e := checkRoute(r.config.RoutedCIDR, r.config.TargetGateway)
	if e != nil {
		return e
	}
	if !exists {
		return nil
	}
	serviceCIDR := r.config.RoutedCIDR.String()
	gatewayIP := r.config.TargetGateway.String()

	fmt.Printf("Cleaning up Route for CIDR %s to gateway %s\n", serviceCIDR, gatewayIP)
	command := exec.Command("Route", "delete", serviceCIDR)
	if stdInAndOut, e := command.CombinedOutput(); e != nil {
		return e
	} else {
		message := fmt.Sprintf("%s", stdInAndOut)
		logrus.Infof("'%s'", message)
		if message != " OK!\r\n" {
			return fmt.Errorf("error deleting Route: %s, %d", message, len(strings.Split(message, "\n")))
		} else {
			return nil
		}
	}
}

func (r *OSRouter) GetConfig() Route {
	return r.config
}
