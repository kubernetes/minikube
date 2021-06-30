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

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
)

// RoutableHostIPFromInside returns the ip/dns of the host that container lives on
// is routable from inside the container
func RoutableHostIPFromInside(ociBin string, clusterName string, containerName string) (net.IP, error) {
	if ociBin == Docker {
		if runtime.GOOS == "linux" {
			info, err := containerNetworkInspect(ociBin, clusterName)
			if err != nil {
				if errors.Is(err, ErrNetworkNotFound) {
					klog.Infof("The container %s is not attached to a network, this could be because the cluster was created by minikube <v1.14, will try to get the IP using container gatway", containerName)

					return containerGatewayIP(Docker, containerName)
				}
				return info.gateway, errors.Wrap(err, "network inspect")
			}
			return info.gateway, nil
		}
		// for windows and mac, the gateway ip is not routable so we use dns trick.
		return digDNS(ociBin, containerName, "host.docker.internal")
	}
	// podman
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

	klog.Infof("got host ip for mount in container by digging dns: %s", ip.String())
	return ip, nil
}

// gatewayIP inspects oci container to find a gateway IP string
func gatewayIP(ociBin, containerName string) (string, error) {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", "--format", "{{.NetworkSettings.Gateway}}", containerName))
	if err != nil {
		return "", errors.Wrapf(err, "inspect gateway")
	}
	if gatewayIP := strings.TrimSpace(rr.Stdout.String()); gatewayIP != "" {
		return gatewayIP, nil
	}

	// https://github.com/kubernetes/minikube/issues/11293
	// need to check nested network
	// check .NetworkSettings.Networks["cluster-name"].Gateway and then
	// .NetworkSettings.Networks["bridge"|"podman"].Gateway
	for _, network := range []string{containerName, defaultBridgeName(ociBin)} {
		gatewayIP, err := networkGateway(ociBin, containerName, network)
		// err == nil here doesn't mean we get a valid gateway IP, it still can be an empty string
		if err != nil {
			return "", err
		}
		if gatewayIP != "" {
			return gatewayIP, nil
		}
	}
	klog.Infof("Couldn't find gateway for container %s", containerName)
	return "", nil
}

func networkGateway(ociBin, container, network string) (string, error) {
	format := fmt.Sprintf(`
{{ if index .NetworkSettings.Networks %q}} 
	{{(index .NetworkSettings.Networks %q).Gateway}}
{{ end }}
`, network, network)
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", "--format", format, container))
	if err != nil {
		return "", errors.Wrapf(err, "inspect gateway")
	}
	return strings.TrimSpace(rr.Stdout.String()), nil
}

// containerGatewayIP gets the default gateway ip for the container
func containerGatewayIP(ociBin string, containerName string) (net.IP, error) {
	gatewayIP, err := gatewayIP(ociBin, containerName)
	if err != nil {
		return nil, errors.Wrapf(err, "inspect gateway")
	}
	return net.ParseIP(gatewayIP), nil
}

// ForwardedPort will return port mapping for a container using cli.
// example : ForwardedPort("docker", "minikube", "22")
// will return the docker assigned port:
// 32769, nil
// only supports TCP ports
func ForwardedPort(ociBin string, ociID string, contPort int) (int, error) {
	var rr *RunResult
	var err error
	var v semver.Version

	if ociBin == Podman {
		rr, err = runCmd(exec.Command(Podman, "version", "--format", "{{.Version}}"))
		if err != nil {
			return 0, errors.Wrapf(err, "podman version")
		}
		output := strings.TrimSpace(rr.Stdout.String())
		v, err = semver.Make(output)
		if err != nil {
			return 0, errors.Wrapf(err, "podman version")
		}
	}

	// podman 2.0.1 introduced docker syntax for .NetworkSettings.Ports (podman#5380)
	if ociBin == Podman && v.LT(semver.Version{Major: 2, Minor: 0, Patch: 1}) {
		rr, err = runCmd(exec.Command(ociBin, "container", "inspect", "-f", fmt.Sprintf("{{range .NetworkSettings.Ports}}{{if eq .ContainerPort %s}}{{.HostPort}}{{end}}{{end}}", fmt.Sprint(contPort)), ociID))
		if err != nil {
			return 0, errors.Wrapf(err, "get port %d for %q", contPort, ociID)
		}
	} else {
		rr, err = runCmd(exec.Command(ociBin, "container", "inspect", "-f", fmt.Sprintf("'{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}'", contPort), ociID))
		if err != nil {

			// Error: "Template parsing error: template: :1:3: executing "" at <index (index .NetworkSettings.Ports "22/tcp") 0>: error calling index: index of untyped nil"
			if strings.Contains(rr.Output(), `<index (index .NetworkSettings.Ports "22/tcp") 0>: error calling index: index of untyped nil`) && contPort == constants.SSHPort {
				return 0, ErrGetSSHPortContainerNotRunning
			}
			if strings.Contains(rr.Output(), "error calling index: index of untyped nil") {
				return 0, ErrGetPortContainerNotRunning
			}
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
		return podmanContainerIP(ociBin, name)
	}
	return dockerContainerIP(ociBin, name)
}

// podmanContainerIP returns ipv4, ipv6 of container or error
func podmanContainerIP(ociBin string, name string) (string, string, error) {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect",
		"-f", "{{.NetworkSettings.IPAddress}}",
		name))
	if err != nil {
		return "", "", errors.Wrapf(err, "podman inspect ip %s", name)
	}
	output := strings.TrimSpace(rr.Stdout.String())
	if output == "" { // podman returns empty for 127.0.0.1
		// check network, if the ip address is missing
		ipv4, ipv6, err := dockerContainerIP(ociBin, name)
		if err == nil {
			return ipv4, ipv6, nil
		}
		return DefaultBindIPV4, "", nil
	}
	return output, "", nil
}

// dockerContainerIP returns ipv4, ipv6 of container or error
func dockerContainerIP(ociBin string, name string) (string, string, error) {
	// retrieve the IP address of the node using docker inspect
	lines, err := inspect(ociBin, name, "{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}")
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
