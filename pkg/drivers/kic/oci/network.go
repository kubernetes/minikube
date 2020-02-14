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
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// RoutableHostIPFromInside returns the ip/dns of the host that container lives on
// is routable from inside the container
func RoutableHostIPFromInside(ociBin string, containerName string) (string, error) {
	if ociBin != Docker {
		return "", fmt.Errorf("RoutableHostIPFromInside is currently only implemented for docker https://github.com/containers/libpod/issues/5205")
	}
	if runtime.GOOS == "linux" {
		return dockerGatewayBridgeIP()
	}
	// for winows and mac, the gateway ip is not routable so we use dns trick.
	return "host.docker.internal", nil
}

// dockerGatewayBridgeIP gets the default gateway ip for the docker bridge on the user's host machine
func dockerGatewayBridgeIP() (string, error) {
	cmd := exec.Command(Docker, "network", "ls", "--filter", "name=bridge", "--format", "{{.ID}}")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "get network bridge. output: %s", string(out))
	}

	bridgeID := strings.TrimSpace(string(out))
	cmd = exec.Command(Docker, "inspect",
		"--format", "{{(index .IPAM.Config 0).Gateway}}", bridgeID)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "inspect IP gatway for bridge network: %q. output: %s", string(out), bridgeID)
	}
	return strings.TrimSpace(string(out)), nil

}
