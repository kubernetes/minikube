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
	if ociBin != Docker {
		return nil, fmt.Errorf("RoutableHostIPFromInside is currently only implemented for docker https://github.com/containers/libpod/issues/5205")
	}
	if runtime.GOOS == "linux" {
		return dockerGatewayIP()
	}
	// for windows and mac, the gateway ip is not routable so we use dns trick.
	return digDNS(ociBin, containerName, "host.docker.internal")
}

// digDNS will get the IP record for a dns
func digDNS(ociBin, containerName, dns string) (net.IP, error) {
	cmd := exec.Command(ociBin, "exec", "-t", containerName, "dig", "+short", dns)
	out, err := cmd.CombinedOutput()
	ip := net.ParseIP(strings.TrimSpace(string(out)))

	if err != nil {
		return ip, errors.Wrapf(err, "resolve dns to ip: %s", string(out))
	}

	glog.Infof("got host ip for mount in container by digging dns: %s", ip.String())
	return ip, nil
}

// dockerGatewayIP gets the default gateway ip for the docker bridge on the user's host machine
// gets the ip from user's host docker
func dockerGatewayIP() (net.IP, error) {
	cmd := exec.Command(Docker, "network", "ls", "--filter", "name=bridge", "--format", "{{.ID}}")
	out, err := cmd.CombinedOutput()

	if err != nil {
		return nil, errors.Wrapf(err, "get network bridge. output: %s", string(out))
	}

	bridgeID := strings.TrimSpace(string(out))
	cmd = exec.Command(Docker, "inspect",
		"--format", "{{(index .IPAM.Config 0).Gateway}}", bridgeID)
	out, err = cmd.CombinedOutput()

	if err != nil {
		return nil, errors.Wrapf(err, "inspect IP gatway for bridge network: %q. output: %s", string(out), bridgeID)
	}

	ip := net.ParseIP(strings.TrimSpace(string(out)))
	glog.Infof("got host ip for mount in container by inspect docker network: %s", ip.String())
	return ip, nil
}

// ForwardedPort will return port mapping for a container using cli.
// example : ForwardedPort("docker", "minikube", "22")
// will return the docker assigned port:
// 32769, nil
// only supports TCP ports
func ForwardedPort(ociBinary string, ociID string, contPort int) (int, error) {
	var out []byte
	var err error

	if ociBinary == Podman {
		//podman inspect -f "{{range .NetworkSettings.Ports}}{{if eq .ContainerPort "80"}}{{.HostPort}}{{end}}{{end}}"
		cmd := exec.Command(ociBinary, "inspect", "-f", fmt.Sprintf("{{range .NetworkSettings.Ports}}{{if eq .ContainerPort %s}}{{.HostPort}}{{end}}{{end}}", fmt.Sprint(contPort)), ociID)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return 0, errors.Wrapf(err, "get host-bind port %d for %q, output %s", contPort, ociID, out)
		}
	} else {
		cmd := exec.Command(ociBinary, "inspect", "-f", fmt.Sprintf("'{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}'", contPort), ociID)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return 0, errors.Wrapf(err, "get host-bind port %d for %q, output %s", contPort, ociID, out)
		}
	}

	o := strings.TrimSpace(string(out))
	o = strings.Trim(o, "'")
	p, err := strconv.Atoi(o)

	if err != nil {
		return p, errors.Wrapf(err, "convert host-port %q to number", p)
	}

	return p, nil
}

// ContainerIPs returns ipv4,ipv6, error of a container by their name
func ContainerIPs(ociBinary string, name string) (string, string, error) {
	if ociBinary == Podman {
		return podmanConttainerIP(name)
	}
	return dockerContainerIP(name)
}

// podmanConttainerIP returns ipv4, ipv6 of container or error
func podmanConttainerIP(name string) (string, string, error) {
	cmd := exec.Command(Podman, "inspect",
		"-f", "{{.NetworkSettings.IPAddress}}",
		name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.Wrapf(err, "podman inspect ip %s", name)
	}
	output := strings.TrimSpace(string(out))
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
