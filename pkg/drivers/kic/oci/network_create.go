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
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/network"
)

// defaultFirstSubnetAddr is a first subnet to be used on first kic cluster
// it is one octet more than the one used by KVM to avoid possible conflict
const defaultFirstSubnetAddr = "192.168.49.0"

// name of the default bridge network, used to lookup the MTU (see #9528)
const dockerDefaultBridge = "bridge"

// name of the default bridge network
const podmanDefaultBridge = "podman"

func defaultBridgeName(ociBin string) string {
	switch ociBin {
	case Docker:
		return dockerDefaultBridge
	case Podman:
		return podmanDefaultBridge
	default:
		klog.Warningf("Unexpected oci:  %v", ociBin)
		return dockerDefaultBridge
	}
}

func firstSubnetAddr(subnet string) string {
	if subnet == "" {
		return defaultFirstSubnetAddr
	}

	return subnet
}

// CreateNetwork creates a network returns gateway and error, minikube creates one network per cluster
func CreateNetwork(ociBin, networkName, subnet, staticIP string) (net.IP, error) {
	defaultBridgeName := defaultBridgeName(ociBin)
	if networkName == defaultBridgeName {
		klog.Infof("skipping creating network since default network %s was specified", networkName)
		return nil, nil
	}

	// check if the network already exists
	info, err := containerNetworkInspect(ociBin, networkName)
	if err == nil {
		klog.Infof("Found existing network %+v", info)
		return info.gateway, nil
	}

	// will try to get MTU from the docker network to avoid issue with systems with exotic MTU settings.
	// related issue #9528
	info, err = containerNetworkInspect(ociBin, defaultBridgeName)
	if err != nil {
		klog.Warningf("failed to get mtu information from the %s's default network %q: %v", ociBin, defaultBridgeName, err)
	}

	tries := 20

	// we don't want to increment the subnet IP on network creation failure if the user specifies a static IP, so set tries to 1
	if staticIP != "" {
		tries = 1
		subnet = staticIP
	}

	// retry up to 5 times to create container network
	for attempts, subnetAddr := 0, firstSubnetAddr(subnet); attempts < 5; attempts++ {
		// Rather than iterate through all of the valid subnets, give up at 20 to avoid a lengthy user delay for something that is unlikely to work.
		// will be like 192.168.49.0/24,..., 192.168.220.0/24 (in increment steps of 9)
		var subnet *network.Parameters
		subnet, err = network.FreeSubnet(subnetAddr, 9, tries)
		if err != nil {
			klog.Errorf("failed to find free subnet for %s network %s after %d attempts: %v", ociBin, networkName, 20, err)
			return nil, fmt.Errorf("un-retryable: %w", err)
		}
		info.gateway, err = tryCreateDockerNetwork(ociBin, subnet, info.mtu, networkName)
		if err == nil {
			klog.Infof("%s network %s %s created", ociBin, networkName, subnet.CIDR)
			return info.gateway, nil
		}
		// don't retry if error is not address is taken
		if !errors.Is(err, ErrNetworkSubnetTaken) && !errors.Is(err, ErrNetworkGatewayTaken) {
			klog.Errorf("error while trying to create %s network %s %s: %v", ociBin, networkName, subnet.CIDR, err)
			return nil, fmt.Errorf("un-retryable: %w", err)
		}
		klog.Warningf("failed to create %s network %s %s, will retry: %v", ociBin, networkName, subnet.CIDR, err)
		subnetAddr = subnet.IP
	}
	return info.gateway, fmt.Errorf("failed to create %s network %s: %w", ociBin, networkName, err)
}

func tryCreateDockerNetwork(ociBin string, subnet *network.Parameters, mtu int, name string) (net.IP, error) {
	gateway := net.ParseIP(subnet.Gateway)
	klog.Infof("attempt to create %s network %s %s with gateway %s and MTU of %d ...", ociBin, name, subnet.CIDR, subnet.Gateway, mtu)
	args := []string{
		"network",
		"create",
		"--driver=bridge",
		fmt.Sprintf("--subnet=%s", subnet.CIDR),
		fmt.Sprintf("--gateway=%s", subnet.Gateway),
	}
	if ociBin == Docker {
		// options documentation https://docs.docker.com/engine/reference/commandline/network_create/#bridge-driver-options
		args = append(args, "-o")
		args = append(args, "--ip-masq")
		args = append(args, "-o")
		args = append(args, "--icc")

		// adding MTU option because #9528
		if mtu > 0 {
			args = append(args, "-o")
			args = append(args, fmt.Sprintf("com.docker.network.driver.mtu=%d", mtu))
		}
	}
	args = append(args, fmt.Sprintf("--label=%s=%s", CreatedByLabelKey, "true"), fmt.Sprintf("--label=%s=%s", ProfileLabelKey, name), name)

	rr, err := runCmd(exec.Command(ociBin, args...))
	if err != nil {
		klog.Warningf("failed to create %s network %s %s with gateway %s and mtu of %d: %v", ociBin, name, subnet.CIDR, subnet.Gateway, mtu, err)
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
		if strings.Contains(rr.Output(), "is already used on the host or by another config") {
			return nil, ErrNetworkGatewayTaken
		}
		return nil, fmt.Errorf("create %s network %s %s with gateway %s and MTU of %d: %w", ociBin, name, subnet.CIDR, subnet.Gateway, mtu, err)
	}
	return gateway, nil
}

// netInfo holds part of a docker or podman network information relevant to kic drivers
type netInfo struct {
	name    string
	subnet  *net.IPNet
	gateway net.IP
	mtu     int
}

func containerNetworkInspect(ociBin string, name string) (netInfo, error) {
	if ociBin == Docker {
		return dockerNetworkInspect(name)
	}
	if ociBin == Podman {
		return podmanNetworkInspect(name)
	}
	return netInfo{}, fmt.Errorf("%s unknown", ociBin)
}

// networkInspect is only used to unmarshal the docker network inspect output and translate it to netInfo
type networkInspect struct {
	Name         string
	Driver       string
	Subnet       string
	Gateway      string
	MTU          int
	ContainerIPs []string
}

var dockerInspectGetter = func(name string) (*RunResult, error) {
	// hack -- 'support ancient versions of docker again (template parsing issue) #10362' and resolve 'Template parsing error: template: :1: unexpected "=" in operand' / 'exit status 64'
	// note: docker v18.09.7 and older use go v1.10.8 and older, whereas support for '=' operator in go templates came in go v1.11
	cmd := exec.Command(Docker, "network", "inspect", name, "--format", `{"Name": "{{.Name}}","Driver": "{{.Driver}}","Subnet": "{{range .IPAM.Config}}{{.Subnet}}{{end}}","Gateway": "{{range .IPAM.Config}}{{.Gateway}}{{end}}","MTU": {{if (index .Options "com.docker.network.driver.mtu")}}{{(index .Options "com.docker.network.driver.mtu")}}{{else}}0{{end}}, "ContainerIPs": [{{range $k,$v := .Containers }}"{{$v.IPv4Address}}",{{end}}]}`)
	rr, err := runCmd(cmd)
	// remove extra ',' after the last element in the ContainerIPs slice
	rr.Stdout = *bytes.NewBuffer(bytes.ReplaceAll(rr.Stdout.Bytes(), []byte(",]"), []byte("]")))
	return rr, err
}

// if exists returns subnet, gateway and mtu
func dockerNetworkInspect(name string) (netInfo, error) {
	var vals networkInspect
	var info = netInfo{name: name}

	rr, err := dockerInspectGetter(name)
	if err != nil {
		logDockerNetworkInspect(Docker, name)
		if isNetworkNotFound(rr.Output()) {
			return info, ErrNetworkNotFound
		}
		return info, err
	}

	// results looks like {"Name": "bridge","Driver": "bridge","Subnet": "172.17.0.0/16","Gateway": "172.17.0.1","MTU": 1500, "ContainerIPs": ["172.17.0.3/16", "172.17.0.2/16"]}
	if err := json.Unmarshal(rr.Stdout.Bytes(), &vals); err != nil {
		return info, fmt.Errorf("error parsing network inspect output: %q", rr.Stdout.String())
	}

	info.gateway = net.ParseIP(vals.Gateway)
	info.mtu = vals.MTU

	_, info.subnet, err = net.ParseCIDR(vals.Subnet)
	if err != nil {
		return info, errors.Wrapf(err, "parse subnet for %s", name)
	}

	return info, nil
}

var podmanInspectGetter = func(name string) (*RunResult, error) {
	v, err := podmanVersion()
	if err != nil {
		return nil, errors.Wrapf(err, "podman version")
	}
	format := `{{range .}}{{if eq .Driver "bridge"}}{{(index .Subnets 0).Subnet}},{{(index .Subnets 0).Gateway}}{{end}}{{end}}`
	if v.LT(semver.Version{Major: 4, Minor: 0, Patch: 0}) {
		// format was changed in Podman 4.0.0: https://github.com/kubernetes/minikube/issues/13861#issuecomment-1082639236
		format = `{{range .plugins}}{{if eq .type "bridge"}}{{(index (index .ipam.ranges 0) 0).subnet}},{{(index (index .ipam.ranges 0) 0).gateway}}{{end}}{{end}}`
	}
	cmd := exec.Command(Podman, "network", "inspect", name, "--format", format)
	return runCmd(cmd)
}

func podmanNetworkInspect(name string) (netInfo, error) {
	var info = netInfo{name: name}
	rr, err := podmanInspectGetter(name)
	if err != nil {
		logDockerNetworkInspect(Podman, name)
		if strings.Contains(rr.Output(), "no such network") {

			return info, ErrNetworkNotFound
		}
		return info, err
	}

	output := strings.TrimSpace(rr.Stdout.String())
	if output == "" {
		return info, fmt.Errorf("no bridge network found for %s", name)
	}

	// results looks like 172.17.0.0/16,172.17.0.1,1500
	vals := strings.Split(output, ",")

	if len(vals) >= 2 {
		info.gateway = net.ParseIP(vals[1])
	}

	_, info.subnet, err = net.ParseCIDR(vals[0])
	if err != nil {
		return info, errors.Wrapf(err, "parse subnet for %s", name)
	}

	return info, nil
}

func logDockerNetworkInspect(ociBin string, name string) {
	cmd := exec.Command(ociBin, "network", "inspect", name)
	klog.Infof("running %v to gather additional debugging logs...", cmd.Args)
	rr, err := runCmd(cmd)
	if err != nil {
		klog.Infof("error running %v: %v", rr.Args, err)
	}
	klog.Infof("output of %v: %v", rr.Args, rr.Output())
}

// RemoveNetwork removes a network
func RemoveNetwork(ociBin string, name string) error {
	if !networkExists(ociBin, name) {
		return nil
	}
	rr, err := runCmd(exec.Command(ociBin, "network", "rm", name))
	if err != nil {
		if isNetworkNotFound(rr.Output()) {
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
	_, err := containerNetworkInspect(ociBin, name)
	if err != nil && !errors.Is(err, ErrNetworkNotFound) { // log unexpected error
		klog.Warningf("Error inspecting docker network %s: %v", name, err)
	}
	return err == nil
}

// networkNamesByLabel returns all network names created by a label
func networkNamesByLabel(ociBin string, label string) ([]string, error) {
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

	return lines, scanner.Err()
}

// DeleteKICNetworksByLabel deletes all networks that have a specific label
func DeleteKICNetworksByLabel(ociBin string, label string) []error {
	var errs []error
	ns, err := networkNamesByLabel(ociBin, label)
	if err != nil {
		return []error{errors.Wrap(err, "list all volume")}
	}
	for _, n := range ns {
		err := RemoveNetwork(ociBin, n)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func isNetworkNotFound(output string) bool {
	// "No such network" on Docker 20.X.X and before, "network %s not found" on Docker 23.X.X and later
	re := regexp.MustCompile(`(No such network)|(network .+ not found)`)
	return re.MatchString(output)
}
