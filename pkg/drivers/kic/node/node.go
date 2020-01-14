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
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"

	"github.com/pkg/errors"
)

const (
	// Docker default bridge network is named "bridge" (https://docs.docker.com/network/bridge/#use-the-default-bridge-network)
	DefaultNetwork  = "bridge"
	ClusterLabelKey = "io.x-k8s.kic.cluster" // ClusterLabelKey is applied to each node docker container for identification
	NodeRoleKey     = "io.k8s.sigs.kic.role"
)

// Node represents a handle to a kic node
// This struct must be created by one of: CreateControlPlane
type Node struct {
	id        string         // container id
	name      string         // container name
	r         command.Runner // Runner
	ociBinary string
}

type CreateConfig struct {
	Name         string            // used for container name and hostname
	Image        string            // container image to use to create the node.
	ClusterLabel string            // label the containers we create using minikube so we can clean up
	Role         string            // currently only role supported is control-plane
	Mounts       []oci.Mount       // volume mounts
	PortMappings []oci.PortMapping // ports to map to container from host
	CPUs         string            // number of cpu cores assign to container
	Memory       string            // memory (mbs) to assign to the container
	Envs         map[string]string // environment variables to pass to the container
	ExtraArgs    []string          // a list of any extra option to pass to oci binary during creation time, for example --expose 8080...
	OCIBinary    string            // docker or podman
}

// CreateNode creates a new container node
func CreateNode(p CreateConfig) (*Node, error) {
	cmder := command.NewKICRunner(p.Name, p.OCIBinary)
	runArgs := []string{
		fmt.Sprintf("--cpus=%s", p.CPUs),
		fmt.Sprintf("--memory=%s", p.Memory),
		"-d", // run the container detached
		"-t", // allocate a tty for entrypoint logs
		// running containers in a container requires privileged
		// NOTE: we could try to replicate this with --cap-add, and use less
		// privileges, but this flag also changes some mounts that are necessary
		// including some ones docker would otherwise do by default.
		// for now this is what we want. in the future we may revisit this.
		"--privileged",
		"--security-opt", "seccomp=unconfined", // also ignore seccomp
		"--tmpfs", "/tmp", // various things depend on working /tmp
		"--tmpfs", "/run", // systemd wants a writable /run
		// logs,pods be stroed on  filesystem vs inside container,
		"--volume", "/var",
		// some k8s things want /lib/modules
		"-v", "/lib/modules:/lib/modules:ro",
		"--hostname", p.Name, // make hostname match container name
		"--name", p.Name, // ... and set the container name
		// label the node with the cluster ID
		"--label", p.ClusterLabel,
		// label the node with the role ID
		"--label", fmt.Sprintf("%s=%s", NodeRoleKey, p.Role),
	}

	for key, val := range p.Envs {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	// adds node specific args
	runArgs = append(runArgs, p.ExtraArgs...)

	if oci.UsernsRemap(p.OCIBinary) {
		// We need this argument in order to make this command work
		// in systems that have userns-remap enabled on the docker daemon
		runArgs = append(runArgs, "--userns=host")
	}

	_, err := oci.CreateContainer(p.OCIBinary,
		p.Image,
		oci.WithRunArgs(runArgs...),
		oci.WithMounts(p.Mounts),
		oci.WithPortMappings(p.PortMappings),
	)

	if err != nil {
		return nil, errors.Wrap(err, "oci create ")
	}

	// we should return a handle so the caller can clean it up
	node, err := Find(p.OCIBinary, p.Name, cmder)
	if err != nil {
		return node, errors.Wrap(err, "find node")
	}

	return node, nil
}

// Find finds a node
func Find(ociBinary string, name string, cmder command.Runner) (*Node, error) {
	n, err := oci.Inspect(ociBinary, name, "{{.Id}}")
	if err != nil {
		return nil, fmt.Errorf("can't find node %v", err)
	}
	return &Node{
		ociBinary: ociBinary,
		id:        n[0],
		name:      name,
		r:         cmder,
	}, nil
}

// WriteFile writes content to dest on the node
func (n *Node) WriteFile(dest, content string, perm string) error {
	// create destination directory
	cmd := exec.Command("mkdir", "-p", filepath.Dir(dest))
	rr, err := n.r.RunCmd(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to create directory %s cmd: %v output:%q", cmd.Args, dest, rr.Output())
	}

	cmd = exec.Command("cp", "/dev/stdin", dest)
	cmd.Stdin = strings.NewReader(content)

	if rr, err := n.r.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "failed to run: cp /dev/stdin %s cmd: %v output:%q", dest, cmd.Args, rr.Output())
	}

	cmd = exec.Command("chmod", perm, dest)
	_, err = n.r.RunCmd(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to run: chmod %s %s", perm, dest)
	}
	return nil
}

// IP returns the IP address of the node
func (n *Node) IP() (ipv4 string, ipv6 string, err error) {
	// retrieve the IP address of the node using docker inspect
	lines, err := oci.Inspect(n.ociBinary, n.name, "{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}")
	if err != nil {
		return "", "", errors.Wrap(err, "node ips")
	}
	if len(lines) != 1 {
		return "", "", errors.Errorf("file should only be one line, got %d lines", len(lines))
	}
	ips := strings.Split(lines[0], ",")
	if len(ips) != 2 {
		return "", "", errors.Errorf("container addresses should have 2 values, got %d values: %+v", len(ips), ips)
	}
	return ips[0], ips[1], nil
}

// Copy copies a local asset into the node
func (n *Node) Copy(ociBinary string, asset assets.CopyableFile) error {
	if err := oci.Copy(ociBinary, n.name, asset); err != nil {
		return errors.Wrap(err, "failed to copy file/folder")
	}

	cmd := exec.Command("chmod", asset.GetPermissions(), asset.GetTargetName())
	if _, err := n.r.RunCmd(cmd); err != nil {
		return errors.Wrap(err, "failed to chmod file permissions")
	}
	return nil
}

// Remove removes the node
func (n *Node) Remove() error {
	return oci.Remove(n.ociBinary, n.name)
}
