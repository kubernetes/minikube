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

package node

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/cri"
	"k8s.io/minikube/pkg/minikube/command"
)

// Spec describes a node to create purely from the container aspect
// this does not inlude eg starting kubernetes (see actions for that)
type Spec struct {
	Name              string
	Profile           string
	Role              string
	Image             string // for example  4000mb based on https://docs.docker.com/config/containers/resource_constraints/
	CPUs              string // for example 2
	Memory            string
	ExtraMounts       []cri.Mount
	ExtraPortMappings []cri.PortMapping
	APIServerPort     int32
	APIServerAddress  string
	IPv6              bool
	Envs              map[string]string // environment variables to be passsed to passed to create nodes
}

func (d *Spec) Create(cmder command.Runner) (node *Node, err error) {
	params := CreateParams{
		Name:         d.Name,
		Image:        d.Image,
		ClusterLabel: ClusterLabelKey + "=" + d.Profile,
		Mounts:       d.ExtraMounts,
		PortMappings: d.ExtraPortMappings,
		Cpus:         d.CPUs,
		Memory:       d.Memory,
		Envs:         d.Envs,
		ExtraArgs:    []string{"--expose", fmt.Sprintf("%d", d.APIServerPort)},
	}

	switch d.Role {
	case "control-plane":
		params.PortMappings = append(params.PortMappings, cri.PortMapping{
			ListenAddress: d.APIServerAddress,
			HostPort:      d.APIServerPort,
			ContainerPort: 6443,
		})
		node, err = CreateNode(
			params,
			cmder,
		)
		if err != nil {
			return node, err
		}

		// stores the port mapping into the node internal state
		node.cache.set(func(cache *nodeCache) {
			cache.ports = map[int32]int32{6443: d.APIServerPort}
		})
		return node, nil

	default:
		return nil, fmt.Errorf("unknown node role: %s", d.Role)
	}
}

// ListNodes lists all the nodes (containers) created by kic on the system
func (d *Spec) ListNodes() ([]string, error) {
	args := []string{
		"ps",
		"-q",         // quiet output for parsing
		"-a",         // show stopped nodes
		"--no-trunc", // don't truncate
		// filter for nodes with the cluster label
		"--filter", "label=" + ClusterLabelKey,
		// format to include friendly name and the cluster name
		"--format", fmt.Sprintf(`{{.Names}}\t{{.Label "%s"}}`, ClusterLabelKey),
	}
	cmd := exec.Command("docker", args...)

	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to list containers for %s", d.Profile))

	}

	lines := []string{}
	scanner := bufio.NewScanner(&buff)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	names := []string{}
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			return nil, errors.Errorf("invalid output when listing containers: %s", line)

		}
		ns := strings.Split(parts[0], ",")
		names = append(names, ns...)
	}
	return names, nil

}
