/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// DefaultSubnet subnet to be used on first cluster
const defaultSubnet = "192.168.39.0/24"

// CreateNetwork creates a network returns gateway and error, minikube creates one network per cluster
func CreateNetwork(ociBin string, name string) (net.IP, error) {
	if ociBin != Docker {
		return nil, fmt.Errorf("%s network not implemented yet", ociBin)
	}
	return createDockerNetwork(name)
}

func createDockerNetwork(name string) (net.IP, error) {
	// check if the network already exists
	subnet, gateway, err := dockerNetworkInspect(name)
	if err == nil {
		glog.Infof("Found existing network with subnet %s and gateway %s.", subnet, gateway)
		return gateway, nil
	}
	// simple way to create networks, subnet is taken, try one bigger
	attempt := 0
	_, subnet, err = net.ParseCIDR(defaultSubnet)
	if err != nil {
		return nil, errors.Wrapf(err, "parse default subnet %s", defaultSubnet)
	}

	gateway, err = tryCreateDockerNetwork(subnet, name)
	if err != nil {
		if err != ErrNetworkSubnetTaken {
			return nil, errors.Wrapf(err, "error creating network")
		}
		// try up to 13 times
		// we can try up to 255
		for attempt < 13 {
			attempt++
			glog.Infof("Couldn't create network %q at %q subnet will try again with a new subnet ...", name, subnet)
			// increase 3nd digit by 10 each time
			// 13 times adding 10 defaultSubnet "192.168.39.0/24"
			// at most it will add up to 169 which is still less than max allowed 255
			// this is large enough to try more and not too small to not try enough
			// can be tuned in the next iterations
			subnet.IP.To4()[2] += 10
			gateway, err := tryCreateDockerNetwork(subnet, name)
			if err == nil {
				return gateway, nil
			}
			if err == ErrNetworkSubnetTaken {
				continue
			}
		}

	}
	return gateway, nil
}

func tryCreateDockerNetwork(subnet *net.IPNet, name string) (net.IP, error) {
	gateway := subnet.IP.To4()
	gateway[3]++ // first ip for gateway
	glog.Infof("attempt to create network %q with subnet: %s and gateway %s...", subnet, name, gateway)
	// options documentation https://docs.docker.com/engine/reference/commandline/network_create/#bridge-driver-options
	rr, err := runCmd(exec.Command(Docker, "network", "create", "--driver=bridge", fmt.Sprintf("--subnet=%s", subnet), fmt.Sprintf("--gateway=%s", gateway), "-o", "--ip-masq", "-o", "--icc", fmt.Sprintf("--label=%s=%s", CreatedByLabelKey, "true"), name))
	if err != nil {
		if strings.Contains(rr.Output(), "Pool overlaps with other one on this address space") {
			return nil, ErrNetworkSubnetTaken
		}
		if strings.Contains(rr.Output(), "failed to allocate gateway") && strings.Contains(rr.Output(), "Address already in use") {
			return nil, ErrNetworkGatewayTaken
		}
		return nil, errors.Wrapf(err, "error creating network")
	}
	return gateway, nil
}

// returns subnet and gate if exists
func dockerNetworkInspect(name string) (*net.IPNet, net.IP, error) {
	rr, err := runCmd(exec.Command(Docker, "network", "inspect", name, "--format", "{{(index .IPAM.Config 0).Subnet}},{{(index .IPAM.Config 0).Gateway}}"))
	if err != nil {
		if strings.Contains(rr.Output(), "No such network:") {
			return nil, nil, ErrNetworkNotFound
		}
		return nil, nil, err
	}
	// results looks like 172.17.0.0/16,172.17.0.1
	ips := strings.Split(strings.TrimSpace(rr.Stdout.String()), ",")
	if len(ips) == 0 {
		return nil, nil, fmt.Errorf("invalid network info")
	}

	_, subnet, err := net.ParseCIDR(ips[0])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "parse subnet for %s", name)
	}
	var gateway net.IP
	if len(ips) > 0 {
		gateway = net.ParseIP(ips[1])
	}
	return subnet, gateway, nil
}

// RemoveNetwork removes a network
func RemoveNetwork(name string) error {
	if !networkExists(name) {
		return nil
	}
	rr, err := runCmd(exec.Command(Docker, "network", "remove", name))
	if err != nil {
		if strings.Contains(rr.Output(), "No such network:") {
			return ErrNetworkNotFound
		}
		// Error response from daemon: error while removing network: network mynet123 id f9e1c50b89feb0b8f4b687f3501a81b618252c9907bc20666e386d0928322387 has active endpoints
		if strings.Contains(rr.Output(), "has active endpoints") {
			return ErrNetworkInUse
		}
	}

	return err
}

func networkExists(name string) bool {
	if _, _, err := dockerNetworkInspect(name); err != nil {
		if err == ErrNetworkNotFound {
			return false
		}
		glog.Warningf("error inspecting network %s: %v", name, err)
		return false
	}
	return true
}

// returns all network names created by a label
func allNetworkByLabel(ociBin string, label string) ([]string, error) {
	if ociBin != Docker {
		return nil, fmt.Errorf("%s not supported", ociBin)
	}

	// docker network ls --filter='label=created_by.minikube.sigs.k8s.io=true' --format '{{.Name}}
	rr, err := runCmd(exec.Command(Docker, "network", "ls", fmt.Sprintf("--filter=label=%s", label), "--format", "{{.Name}}"))
	if err != nil {
		return nil, err
	}
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(rr.Stdout.Bytes()))
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	return lines, nil
}

// DeleteAllNetworksByKIC deletes all networks created by kic
func DeleteAllNetworksByKIC() []error {
	var errs []error
	ns, err := allNetworkByLabel(Docker, CreatedByLabelKey+"=true")
	if err != nil {
		return []error{errors.Wrap(err, "list all volume")}
	}
	for _, n := range ns {
		err := RemoveNetwork(n)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
