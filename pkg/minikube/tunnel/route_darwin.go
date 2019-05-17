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
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
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
	if err := writeResolverFile(route); err != nil {
		return fmt.Errorf("could not write /etc/resolver/{cluster_domain} file: %s", err)
	}

	serviceCIDR := route.DestCIDR.String()
	gatewayIP := route.Gateway.String()

	glog.Infof("Adding route for CIDR %s to gateway %s", serviceCIDR, gatewayIP)
	command := exec.Command("sudo", "route", "-n", "add", serviceCIDR, gatewayIP)
	glog.Infof("About to run command: %s", command.Args)
	stdInAndOut, err := command.CombinedOutput()
	message := fmt.Sprintf("%s", stdInAndOut)
	re := regexp.MustCompile(fmt.Sprintf("add net (.*): gateway %s\n", gatewayIP))
	if !re.MatchString(message) {
		return fmt.Errorf("error adding Route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	glog.Infof("%s", stdInAndOut)

	return err
}

func (router *osRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	cmd := exec.Command("netstat", "-nr", "-f", "inet")
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

		if len(fields) <= 2 {
			continue
		}
		dstCIDRString := router.padCIDR(fields[0])
		gatewayIPString := fields[1]
		gatewayIP := net.ParseIP(gatewayIPString)

		_, ipNet, err := net.ParseCIDR(dstCIDRString)
		if err != nil {
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
	exists, err := isValidToAddOrDelete(router, route)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	command := exec.Command("sudo", "route", "-n", "delete", route.DestCIDR.String())
	stdInAndOut, err := command.CombinedOutput()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s", stdInAndOut)
	glog.V(4).Infof("%s", message)
	re := regexp.MustCompile("^delete net ([^:]*)$")
	if !re.MatchString(message) {
		return fmt.Errorf("error deleting route: %s, %d", message, len(strings.Split(message, "\n")))
	}
	// idempotent removal of cluster domain dns
	resolverFile := fmt.Sprintf("/etc/resolver/%s", route.ClusterDomain)
	command = exec.Command("sudo", "rm", "-f", resolverFile)
	if err := command.Run(); err != nil {
		return fmt.Errorf("could not remove %s: %s", resolverFile, err)
	}
	return nil
}

func writeResolverFile(route *Route) error {
	resolverFile := "/etc/resolver/" + route.ClusterDomain
	content := fmt.Sprintf("nameserver %s\nsearch_order 1\n", route.ClusterDNSIP)
	// write resolver content into tmpFile, then copy it to /etc/resolver/clusterDomain
	tmpFile, err := ioutil.TempFile("", "minikube-tunnel-resolver-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err = tmpFile.WriteString(content); err != nil {
		return err
	}
	if err = tmpFile.Close(); err != nil {
		return err
	}
	command := exec.Command("sudo", "mkdir", "-p", "/etc/resolver")
	if err := command.Run(); err != nil {
		return err
	}
	command = exec.Command("sudo", "cp", "-f", tmpFile.Name(), resolverFile)
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}
