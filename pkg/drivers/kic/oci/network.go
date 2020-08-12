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
			return dockerGatewayIP(containerName)
		}
		// for windows and mac, the gateway ip is not routable so we use dns trick.
		return digDNS(ociBin, containerName, "host.docker.internal")
	}

	if runtime.GOOS == "linux" {
		return containerGatewayIP(ociBin, containerName)
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

// dockerGatewayIP gets the default gateway ip for the docker bridge on the user's host machine
// gets the ip from user's host docker
func dockerGatewayIP(profile string) (net.IP, error) {
	// check if using custom network first
	if networkExists(profile) {
		ip := net.ParseIP(DefaultGateway)
		return ip, nil
	}
	rr, err := runCmd(exec.Command(Docker, "network", "ls", "--filter", "name=bridge", "--format", "{{.ID}}"))
	if err != nil {
		return nil, errors.Wrapf(err, "get network bridge")
	}

	bridgeID := strings.TrimSpace(rr.Stdout.String())
	rr, err = runCmd(exec.Command(Docker, "network", "inspect",
		"--format", "{{(index .IPAM.Config 0).Gateway}}", bridgeID))
	if err != nil {
		return nil, errors.Wrapf(err, "inspect IP bridge network %q.", bridgeID)
	}

	ip := net.ParseIP(strings.TrimSpace(rr.Stdout.String()))
	glog.Infof("got host ip for mount in container by inspect docker network: %s", ip.String())
	return ip, nil
}

// containerGatewayIP gets the default gateway ip for the container
func containerGatewayIP(ociBin, containerName string) (net.IP, error) {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", "--format", "{{.NetworkSettings.Gateway}}", containerName))
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

// CreateNetwork creates a network
func CreateNetwork(name, ipRange, gateway string) error {
	// check if the network already exists
	if networkExists(name) {
		return nil
	}

	subnet := fmt.Sprintf("--subnet=%s", ipRange)
	_, err := runCmd(exec.Command(Docker, "network", "create", "--driver=bridge", subnet, "--gateway", gateway, name))
	if err != nil {
		return errors.Wrapf(err, "error creating network")
	}

	return nil
}

// removeNetwork removes a network
func removeNetwork(name string) error {
	if !networkExists(name) {
		return nil
	}
	_, err := runCmd(exec.Command(Docker, "network", "remove", name))
	return err
}

func networkExists(name string) bool {
	rr, err := runCmd(exec.Command(Docker, "network", "ls", "--format", "{{.Name}}"))
	if err != nil {
		glog.Warningf("error listing networks: %v", err)
		return false
	}
	networks := strings.Split(rr.Output(), "\n")
	for _, n := range networks {
		if strings.Trim(n, "\n") == name {
			return true
		}
	}
	return false
}
