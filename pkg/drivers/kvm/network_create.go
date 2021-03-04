/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/network"
)

// firstSubnetAddr subnet to be used on first kvm cluster
// it is one octet more than the one used by KVM to avoid possible conflict
const firstSubnetAddr = "192.168.49.0"

// big enough for a cluster of 254 nodes
const defaultSubnetMask = 24

// name of the default bridge network
const kvmDefaultBridge = "kvm"

// CreateNetwork creates a network returns gateway and error, minikube creates one network per cluster
func CreateKvmNetwork(ociBin string, networkName string) (net.IP, error) {
	var defaultBridgeName string
	if ociBin == kvm {
		defaultBridgeName = kvmDefaultBridge
	}
	
	
	if networkName == defaultBridgeName {
		klog.Infof("skipping creating network since default network %s was specified", networkName)
		return nil, nil
	}

	// check if the network already exists
	info, err := ContainerNetworkInspect(ociBin, networkName)
	if err == nil {
		klog.Infof("Found existing network %+v", info)
		return info.gateway, nil
	}

	// will try to get MTU from the docker network to avoid issue with systems with exotic MTU settings.
	// related issue #9528
	info, err = ContainerNetworkInspect(ociBin, defaultBridgeName)
	if err != nil {
		klog.Warningf("failed to get mtu information from the %s's default network %q: %v", ociBin, defaultBridgeName, err)
	}
	// Rather than iterate through all of the valid subnets, give up at 20 to avoid a lengthy user delay for something that is unlikely to work.
	// will be like 192.168.49.0/24 ,...,192.168.239.0/24
	subnet, err := network.FreeSubnet(firstSubnetAddr, 10, 20)
	if err != nil {
		klog.Errorf("error while trying to create network: %v", err)
		return nil, errors.Wrap(err, "un-retryable")
	}
	info.gateway, err = tryCreateKvmNetwork(ociBin, subnet.IP, defaultSubnetMask, info.mtu, networkName)
	if err != nil {
		return info.gateway, fmt.Errorf("failed to create network after 20 attempts")
	}
	return info.gateway, nil
}

func tryCreateKvmNetwork(ociBin string, subnetAddr string, subnetMask int, mtu int, name string) (net.IP, error) {
	gateway := net.ParseIP(subnetAddr)
	gateway.To4()[3]++ // first ip for gateway
	klog.Infof("attempt to create network %s/%d with subnet: %s and gateway %s and MTU of %d ...", subnetAddr, subnetMask, name, gateway, mtu)
	args := []string{
		"network",
		"create",
		"--driver=kvm",
		fmt.Sprintf("--subnet=%s", fmt.Sprintf("%s/%d", subnetAddr, subnetMask)),
		fmt.Sprintf("--gateway=%s", gateway),
	}

	args = append(args, name)
	rr, err := runCmd(exec.Command(ociBin, args...))
	if err != nil {
		// Pool overlaps with other one on this address space
		if strings.Contains(rr.Output(), "Pool overlaps") {
			return nil, ErrNetworkSubnetTaken
		}
		if strings.Contains(rr.Output(), "failed to allocate gateway") && strings.Contains(rr.Output(), "Address already in use") {
			return nil, ErrNetworkGatewayTaken
		}
		if strings.Contains(rr.Output(), "is being used by a network interface") {
			return nil, ErrNetworkGatewayTaken
		}
		return nil, errors.Wrapf(err, "create network %s", fmt.Sprintf("%s %s/%d", name, subnetAddr, subnetMask))
	}
	return gateway, nil
}

// netInfo holds part of a kvm network information relevant to kvm drivers
type netInfo struct {
	name    string
	subnet  *net.IPNet
	gateway net.IP
	mtu     int
}

func ContainerNetworkInspect(ociBin string, name string) (netInfo, error) {
	if ociBin == kvm {
		return kvmNetworkInspect(name)
	}
	return netInfo{}, fmt.Errorf("%s unknown", ociBin)
}

// networkInspect is only used to unmarshal the kvm network inspect output and translate it to netInfo
type networkInspect struct {
	Name         string
	Driver       string
	Subnet       string
	Gateway      string
	MTU          int
	ContainerIPs []string
}

func kvmNetworkInspect(name string) (netInfo, error) {
	var info = netInfo{name: name}
	cmd := exec.Command(kvm, "network", "inspect", name, "--format", `{{range .plugins}}{{if eq .type "bridge"}}{{(index (index .ipam.ranges 0) 0).subnet}},{{(index (index .ipam.ranges 0) 0).gateway}}{{end}}{{end}}`)
	rr, err := runCmd(cmd)
	if err != nil {
		logKvmNetworkInspect(kvm, name)
		if strings.Contains(rr.Output(), "No such network") {

			return info, ErrNetworkNotFound
		}
		return info, err
	}

	output := rr.Stdout.String()
	if output == "" {
		return info, fmt.Errorf("no bridge network found for %s", name)
	}

	// results looks like 172.17.0.0/16,172.17.0.1,1500
	vals := strings.Split(strings.TrimSpace(output), ",")
	if len(vals) == 0 {
		return info, fmt.Errorf("empty list network inspect: %q", rr.Output())
	}

	if len(vals) > 0 {
		info.gateway = net.ParseIP(vals[1])
	}

	_, info.subnet, err = net.ParseCIDR(vals[0])
	if err != nil {
		return info, errors.Wrapf(err, "parse subnet for %s", name)
	}

	return info, nil
}

func logKvmNetworkInspect(ociBin string, name string) {
	cmd := exec.Command(ociBin, "network", "inspect", name)
	klog.Infof("running %v to gather additional debugging logs...", cmd.Args)
	rr, err := runCmd(cmd)
	if err != nil {
		klog.Infof("error running %v: %v", rr.Args, err)
	}
	klog.Infof("output of %v: %v", rr.Args, rr.Output())
}

// RemovesNetwork removes a network
func RemovesNetwork(ociBin string, name string) error {
	if !networkExists(ociBin, name) {
		return nil
	}
	rr, err := runCmd(exec.Command(ociBin, "network", "rm", name))
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

func networkExists(ociBin string, name string) bool {
	_, err := ContainerNetworkInspect(ociBin, name)
	if err != nil && !errors.Is(err, ErrNetworkNotFound) { // log unexpected error
		klog.Warningf("Error inspecting docker network %s: %v", name, err)
	}
	return err == nil
}

// networkNameByLabel returns all network names created by a label
func networkNameByLabel(ociBin string, label string) ([]string, error) {
	// docker network ls --filter='label=created_by.minikube.sigs.k8s.io=true' --format '{{.Name}}'
	rr, err := runCmd(exec.Command(ociBin, "network", "ls", fmt.Sprintf("--filter=label=%s", label), "--format", "{{.Name}}"))
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

// DeleteKvmNetworks deletes all networks created by kvm
func DeleteKvmNetworks(ociBin string) []error {
	var errs []error
	ns, err := networkNameByLabel(ociBin, CreatedByLabelKey)
	if err != nil {
		return []error{errors.Wrap(err, "list all volume")}
	}
	for _, n := range ns {
		err := RemovesNetwork(ociBin, n)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
