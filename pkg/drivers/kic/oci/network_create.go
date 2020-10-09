/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// firstSubnetAddr subnet to be used on first kic cluster
// it is one octet more than the one used by KVM to avoid possible conflict
const firstSubnetAddr = "192.168.49.0"

// big enough for a cluster of 254 nodes
const defaultSubnetMask = 24

// CreateNetwork creates a network returns gateway and error, minikube creates one network per cluster
func CreateNetwork(ociBin string, name string) (net.IP, error) {
	if ociBin != Docker {
		return nil, fmt.Errorf("%s network not implemented yet", ociBin)
	}
	return createDockerNetwork(name)
}

func createDockerNetwork(clusterName string) (net.IP, error) {
	// check if the network already exists
	subnet, gateway, err := dockerNetworkInspect(clusterName)
	if err == nil {
		glog.Infof("Found existing network with subnet %s and gateway %s.", subnet, gateway)
		return gateway, nil
	}

	attempts := 0
	subnetAddr := firstSubnetAddr
	// Rather than iterate through all of the valid subnets, give up at 20 to avoid a lengthy user delay for something that is unlikely to work.
	// will be like 192.168.49.0/24 ,...,192.168.239.0/24
	for attempts < 20 {
		gateway, err = tryCreateDockerNetwork(subnetAddr, defaultSubnetMask, clusterName)
		if err == nil {
			return gateway, nil
		}

		// don't retry if error is not adddress is taken
		if !(errors.Is(err, ErrNetworkSubnetTaken) || errors.Is(err, ErrNetworkGatewayTaken)) {
			glog.Errorf("error while trying to create network %v", err)
			return nil, errors.Wrap(err, "un-retryable")
		}
		attempts++
		// Find an open subnet by incrementing the 3rd octet by 10 for each try
		// 13 times adding 10 firstSubnetAddr "192.168.49.0/24"
		// at most it will add up to 169 which is still less than max allowed 255
		// this is large enough to try more and not too small to not try enough
		// can be tuned in the next iterations
		newSubnet := net.ParseIP(subnetAddr).To4()
		newSubnet[2] += byte(9 + attempts)
		subnetAddr = newSubnet.String()
	}
	return gateway, fmt.Errorf("failed to create network after 20 attempts")
}

func tryCreateDockerNetwork(subnetAddr string, subnetMask int, name string) (net.IP, error) {
	gateway := net.ParseIP(subnetAddr)
	gateway.To4()[3]++ // first ip for gateway
	glog.Infof("attempt to create network %s/%d with subnet: %s and gateway %s...", subnetAddr, subnetMask, name, gateway)
	// options documentation https://docs.docker.com/engine/reference/commandline/network_create/#bridge-driver-options
	rr, err := runCmd(exec.Command(Docker, "network", "create", "--driver=bridge", fmt.Sprintf("--subnet=%s", fmt.Sprintf("%s/%d", subnetAddr, subnetMask)), fmt.Sprintf("--gateway=%s", gateway), "-o", "--ip-masq", "-o", "--icc", fmt.Sprintf("--label=%s=%s", CreatedByLabelKey, "true"), name))
	if err != nil {
		// Pool overlaps with other one on this address space
		if strings.Contains(rr.Output(), "Pool overlaps") {
			return nil, ErrNetworkSubnetTaken
		}
		if strings.Contains(rr.Output(), "failed to allocate gateway") && strings.Contains(rr.Output(), "Address already in use") {
			return nil, ErrNetworkGatewayTaken
		}
		return nil, errors.Wrapf(err, "create network %s", fmt.Sprintf("%s %s/%d", name, subnetAddr, subnetMask))
	}
	return gateway, nil
}

// returns subnet and gate if exists
func dockerNetworkInspect(name string) (*net.IPNet, net.IP, error) {
	cmd := exec.Command(Docker, "network", "inspect", name, "--format", "{{(index .IPAM.Config 0).Subnet}},{{(index .IPAM.Config 0).Gateway}}")
	rr, err := runCmd(cmd)
	if err != nil {
		logDockerNetworkInspect(name)
		if strings.Contains(rr.Output(), "No such network") {
			return nil, nil, ErrNetworkNotFound
		}
		return nil, nil, err
	}
	// results looks like 172.17.0.0/16,172.17.0.1
	ips := strings.Split(strings.TrimSpace(rr.Stdout.String()), ",")
	if len(ips) == 0 {
		return nil, nil, fmt.Errorf("empty IP list parsed from: %q", rr.Output())
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

func logDockerNetworkInspect(name string) {
	cmd := exec.Command(Docker, "network", "inspect", name)
	glog.Infof("running %v to gather additional debugging logs...", cmd.Args)
	rr, err := runCmd(cmd)
	if err != nil {
		glog.Infof("error running %v: %v", rr.Args, err)
	}
	glog.Infof("output of %v: %v", rr.Args, rr.Output())
}

// RemoveNetwork removes a network
func RemoveNetwork(name string) error {
	if !networkExists(name) {
		return nil
	}
	rr, err := runCmd(exec.Command(Docker, "network", "remove", name))
	if err != nil {
		if strings.Contains(rr.Output(), "No such network") {
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
	_, _, err := dockerNetworkInspect(name)
	if err != nil && !errors.Is(err, ErrNetworkNotFound) { // log unexpected error
		glog.Warningf("Error inspecting docker network %s: %v", name, err)
	}
	return err == nil
}

// networkNamesByLabel returns all network names created by a label
func networkNamesByLabel(ociBin string, label string) ([]string, error) {
	if ociBin != Docker {
		return nil, fmt.Errorf("%s not supported", ociBin)
	}

	// docker network ls --filter='label=created_by.minikube.sigs.k8s.io=true' --format '{{.Name}}'
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

// DeleteKICNetworks deletes all networks created by kic
func DeleteKICNetworks() []error {
	var errs []error
	ns, err := networkNamesByLabel(Docker, CreatedByLabelKey+"=true")
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
