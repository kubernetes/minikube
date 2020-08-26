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
	"runtime"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// RoutableHostIPFromInside returns the ip/dns of the host that container lives on
// is routable from inside the container
func RoutableHostIPFromInside(ociBin string, containerName string) (net.IP, error) {
	if ociBin == Docker {
		if runtime.GOOS == "linux" {
			_, gateway, err := dockerNetworkInspect(containerName)
			if err != nil {
				return gateway, errors.Wrap(err, "routable gateway")
			}
			return gateway, nil
		}
		// for windows and mac, the gateway ip is not routable so we use dns trick.
		return digDNS(ociBin, containerName, "host.docker.internal")
	}

	if runtime.GOOS == "linux" {
		return podmanGatewayIP(containerName)
	}

	return nil, fmt.Errorf("RoutableHostIPFromInside is currently only implemented for linux")
}

// digDNS will get the IP record for a dns
func digDNS(ociBin, containerName, dns string) (net.IP, error) {
	rr, err := runCmd(exec.Command(ociBin, "exec", "-t", containerName, "dig", "+short", dns))
	ip := net.ParseIP(strings.TrimSpace(rr.Stdout.String()))
	if err != nil {
		return ip, errors.Wrapf(err, "resolve dns to ip")
	}

	glog.Infof("got host ip for mount in container by digging dns: %s", ip.String())
	return ip, nil
}

// podmanGatewayIP gets the default gateway ip for the container
func podmanGatewayIP(containerName string) (net.IP, error) {
	rr, err := runCmd(exec.Command(Podman, "container", "inspect", "--format", "{{.NetworkSettings.Gateway}}", containerName))
	if err != nil {
		return nil, errors.Wrapf(err, "inspect gateway")
	}
	ip := net.ParseIP(strings.TrimSpace(rr.Stdout.String()))
	return ip, nil
}

// ForwardedPort will return port mapping for a container using cli.
// example : ForwardedPort("docker", "minikube", "22")
// will return the docker assigned port:
// 32769, nil
// only supports TCP ports
func ForwardedPort(ociBin string, ociID string, contPort int) (int, error) {
	var rr *RunResult
	var err error

	if ociBin == Podman {
		rr, err = runCmd(exec.Command(ociBin, "container", "inspect", "-f", fmt.Sprintf("{{range .NetworkSettings.Ports}}{{if eq .ContainerPort %s}}{{.HostPort}}{{end}}{{end}}", fmt.Sprint(contPort)), ociID))
		if err != nil {
			return 0, errors.Wrapf(err, "get port %d for %q", contPort, ociID)
		}
	} else {
		rr, err = runCmd(exec.Command(ociBin, "container", "inspect", "-f", fmt.Sprintf("'{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}'", contPort), ociID))
		if err != nil {
			return 0, errors.Wrapf(err, "get port %d for %q", contPort, ociID)
		}
	}

	o := strings.TrimSpace(rr.Stdout.String())
	o = strings.Trim(o, "'")
	p, err := strconv.Atoi(o)

	if err != nil {
		return p, errors.Wrapf(err, "convert host-port %q to number", p)
	}

	return p, nil
}

// ContainerIPs returns ipv4,ipv6, error of a container by their name
func ContainerIPs(ociBin string, name string) (string, string, error) {
	if ociBin == Podman {
		return podmanContainerIP(name)
	}
	return dockerContainerIP(name)
}

// podmanContainerIP returns ipv4, ipv6 of container or error
func podmanContainerIP(name string) (string, string, error) {
	rr, err := runCmd(exec.Command(Podman, "container", "inspect",
		"-f", "{{.NetworkSettings.IPAddress}}",
		name))
	if err != nil {
		return "", "", errors.Wrapf(err, "podman inspect ip %s", name)
	}
	output := strings.TrimSpace(rr.Stdout.String())
	if err == nil && output == "" { // podman returns empty for 127.0.0.1
		return DefaultBindIPV4, "", nil
	}
	return output, "", nil
}

// dockerContainerIP returns ipv4, ipv6 of container or error
func dockerContainerIP(name string) (string, string, error) {
	// retrieve the IP address of the node using docker inspect
	lines, err := inspect(Docker, name, "{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}")
	if err != nil {
		return "", "", errors.Wrap(err, "inspecting NetworkSettings.Networks")
	}

	if len(lines) != 1 {
		return "", "", errors.Errorf("IPs output should only be one line, got %d lines", len(lines))
	}

	ips := strings.Split(lines[0], ",")
	if len(ips) != 2 {
		return "", "", errors.Errorf("container addresses should have 2 values, got %d values: %+v", len(ips), ips)
	}

	return ips[0], ips[1], nil
}

// CreateNetwork creates a network returns subnet and error
func CreateNetwork(name string) (*net.IPNet, error) {
	// check if the network already exists
	subnet, _, err := dockerNetworkInspect(name)
	if err == nil {
		return subnet, nil
	}

	// simple way to create networks, subnet is taken, try one bigger
	attempt := 0
	_, subnet, err = net.ParseCIDR(defaultSubnet)
	if err != nil {
		return nil, errors.Wrapf(err, "parse default subnet %s", defaultSubnet)
	}

	err = attemptCreateNework(subnet, name)
	if err != nil {
		if err == ErrNetworkSubnetTaken {
			// try up tp 13 times
			for attempt < 13 {
				attempt++
				glog.Infof("Couldn't create network %q at %q subnet will try again a new subnet ...", name, subnet)
				// increase 2nd digit by 10 each time
				subnet.IP.To4()[2] += 10
				err := attemptCreateNework(subnet, name)
				if err == nil {
					return subnet, nil
				}
				if err == ErrNetworkSubnetTaken {
					continue
				}
			}
		}
		return nil, errors.Wrapf(err, "error creating network")

	}
	return subnet, nil
}

func attemptCreateNework(subnet *net.IPNet, name string) error {
	gateway := subnet.IP.To4()
	gateway[3]++ // first ip for gateway
	glog.Infof("attempt to create network %q with subnet: %s and gateway %s...", subnet, name, gateway)
	rr, err := runCmd(exec.Command(Docker, "network", "create", "--driver=bridge", fmt.Sprintf("--subnet=%s", subnet), fmt.Sprintf("--gateway=%s", gateway), "-o", "com.docker.network.bridge.enable_ip_masquerade=true", fmt.Sprintf("--label=%s=%s", CreatedByLabelKey, "true"), name))
	if err != nil {
		if strings.Contains(rr.Output(), "Pool overlaps with other one on this address space") {
			return ErrNetworkSubnetTaken
		}
		if strings.Contains(rr.Output(), "failed to allocate gateway") && strings.Contains(rr.Output(), "Address already in use") {
			return ErrNetworkGatewayTaken
		}
		return errors.Wrapf(err, "error creating network")
	}
	return nil
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

// DeleteAllNetworksByKIC delets all networks created by kic
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
