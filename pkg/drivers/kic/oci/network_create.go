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
const defaultFirstSubnetAddrv6 = "fd00::/64"

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

func firstSubnetAddrv6(subnet string) string {
    if subnet == "" {
        return defaultFirstSubnetAddrv6
    }
   return subnet
}

// CreateNetwork creates a network returns gateway and error, minikube creates one network per cluster
func CreateNetwork(ociBin, networkName, subnet, staticIP string) (net.IP, error) {
	return CreateNetworkWithIPFamily(ociBin, networkName, subnet, "", staticIP, "", "ipv4")
}

func CreateNetworkWithIPFamily(ociBin, networkName, subnet, subnetv6, staticIP, staticIPv6, ipFamily string) (net.IP, error) {
	bridgeName := defaultBridgeName(ociBin)
	if networkName == bridgeName {
		klog.Infof("skipping creating network since default network %s was specified", networkName)
		return nil, nil
	}

	// For IPv6-only or dual-stack networks, use the v6/dual creator
    	if ipFamily == "ipv6" || ipFamily == "dual" {
        	return createV6OrDualNetwork(ociBin, networkName, subnet, subnetv6, ipFamily)
    	}

	// check if the network already exists
	info, err := containerNetworkInspect(ociBin, networkName)
	if err == nil {
		klog.Infof("Found existing network %+v", info)
		return info.gateway, nil
	}

	// will try to get MTU from the docker network to avoid issue with systems with exotic MTU settings.
	// related issue #9528
	info, err = containerNetworkInspect(ociBin, bridgeName)
	if err != nil {
		klog.Warningf("failed to get mtu information from the %s's default network %q: %v", ociBin, bridgeName, err)
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

// createV6OrDualNetwork creates a user-defined bridge network with IPv6 enabled,
// and adds both subnets when ipFamily == "dual". Returns the gateway reported by inspect (may be nil for v6-only).
func createV6OrDualNetwork(ociBin, name, subnetV4, subnetV6, ipFamily string) (net.IP, error) {
    klog.Infof("creating %s network %q (family=%s, v4=%q, v6=%q)", ociBin, name, ipFamily, subnetV4, subnetV6)

    // If exists, reuse as-is
    if info, err := containerNetworkInspect(ociBin, name); err == nil {
        klog.Infof("found existing network %q", name)
        return info.gateway, nil
    }

    // Defaults
    if subnetV6 == "" {
        subnetV6 = firstSubnetAddrv6("")
    }

   // Build args
    args := []string{"network", "create", "--driver=bridge"}
    if ipFamily == "ipv6" || ipFamily == "dual" {
        args = append(args, "--ipv6")
    }
    if ipFamily == "dual" && subnetV4 != "" {
        args = append(args, "--subnet", subnetV4)
    }
    if subnetV6 != "" {
        args = append(args, "--subnet", subnetV6)
    }
    // Optional bridge knobs (safe to omit)
    // if ociBin == Docker {
    //     args = append(args, "-o", "com.docker.network.bridge.enable_ip_masquerade=true")
    //     args = append(args, "-o", "com.docker.network.bridge.enable_icc=true")
    // }
    args = append(args,
        fmt.Sprintf("--label=%s=%s", CreatedByLabelKey, "true"),
        fmt.Sprintf("--label=%s=%s", ProfileLabelKey, name),
        name,
    )

    if _, err := runCmd(exec.Command(ociBin, args...)); err != nil {
        klog.Warningf("failed to create %s network %q: %v", ociBin, name, err)
        return nil, fmt.Errorf("create %s network %q: %w", ociBin, name, err)
    }

    // Rely on inspect (gateway may be empty for IPv6)
    info, err := containerNetworkInspect(ociBin, name)
    if err != nil {
        // non-fatal: just return nil gateway
        klog.Warningf("post-create inspect failed for %s: %v", name, err)
        return nil, nil
    }
    return info.gateway, nil
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
	Name         string   `json:"Name"`
	Driver       string   `json:"Driver"`
	// Legacy single fields (older template)
	Subnet       string   `json:"Subnet"`
	Gateway      string   `json:"Gateway"`
	// Multi-family (new template)
	Subnets      []string `json:"Subnets"`
	Gateways     []string `json:"Gateways"`
	MTU          int      `json:"MTU"`
	ContainerIPs []string `json:"ContainerIPs"`
}


var dockerInspectGetter = func(name string) (*RunResult, error) {
    // keep the old workaround: avoid eq/== in templates; emit trailing commas then strip ",]"
	cmd := exec.Command(
		Docker, "network", "inspect", name, "--format",
		`{"Name":"{{.Name}}","Driver":"{{.Driver}}",` +
			`"Subnet":"{{range .IPAM.Config}}{{.Subnet}}{{end}}",` +
			`"Gateway":"{{range .IPAM.Config}}{{.Gateway}}{{end}}",` +
			`"Subnets":[{{range .IPAM.Config}}{{if .Subnet}}"{{.Subnet}}",{{end}}{{end}}],` +
			`"Gateways":[{{range .IPAM.Config}}{{if .Gateway}}"{{.Gateway}}",{{end}}{{end}}],` +
			`"MTU":{{if (index .Options "com.docker.network.driver.mtu")}}{{(index .Options "com.docker.network.driver.mtu")}}{{else}}0{{end}},` +
			`"ContainerIPs":[{{range $k,$v := .Containers}}"{{$v.IPv4Address}}",{{end}}]}`,
	)
	rr, err := runCmd(cmd)
    	// remove any trailing commas from arrays we just built
    	cleaned := bytes.ReplaceAll(rr.Stdout.Bytes(), []byte(",]"), []byte("]"))
	rr.Stdout = *bytes.NewBuffer(cleaned)
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

	// results look like:
	// {"Name":"bridge","Driver":"bridge",
	//  "Subnet":"172.17.0.0/16","Gateway":"172.17.0.1",
	//  "Subnets":["172.17.0.0/16","fd00::/64"],"Gateways":["172.17.0.1","fd00::1"],
	//  "MTU":1500,"ContainerIPs":[...]}
	if err := json.Unmarshal(rr.Stdout.Bytes(), &vals); err != nil {
		return info, fmt.Errorf("error parsing network inspect output: %q", rr.Stdout.String())
	}

	// Choose a subnet/gateway:
	// - Prefer an IPv4 entry (back-compat with existing IPv4 flows),
	// - else fall back to first entry,
	// - else use legacy single fields.
	pickSubnet := ""
	pickGateway := ""
	if len(vals.Subnets) > 0 {
		for i, s := range vals.Subnets {
			if ip, _, e := net.ParseCIDR(s); e == nil && ip.To4() != nil {
				pickSubnet = s
				if i < len(vals.Gateways) {
					pickGateway = vals.Gateways[i]
				}
				break
			}
		}
		if pickSubnet == "" {
			pickSubnet = vals.Subnets[0]
			if len(vals.Gateways) > 0 {
				pickGateway = vals.Gateways[0]
			}
		}
	}
	if pickSubnet == "" {
		pickSubnet = vals.Subnet
		pickGateway = vals.Gateway
	}
	if pickSubnet != "" {
		if _, info.subnet, err = net.ParseCIDR(pickSubnet); err != nil {
			return info, errors.Wrapf(err, "parse subnet for %s", name)
		}
	}
	if pickGateway != "" {
		info.gateway = net.ParseIP(pickGateway)
	}
	info.mtu = vals.MTU
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
