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
	"os"
	"strconv"

	"bufio"
	"bytes"

	"github.com/pkg/errors"

	"fmt"
	"os/exec"
	"strings"
)

// Inspect return low-level information on containers
func Inspect(ociBinary string, containerNameOrID, format string) ([]string, error) {
	cmd := exec.Command(ociBinary, "inspect",
		"-f", format,
		containerNameOrID) // ... against the "node" container
	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	err := cmd.Run()
	scanner := bufio.NewScanner(&buff)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, err
}

/*
This is adapated from:
https://github.com/kubernetes/kubernetes/blob/07a5488b2a8f67add543da72e8819407d8314204/pkg/kubelet/dockershim/helpers.go#L115-L155
*/
// generateMountBindings converts the mount list to a list of strings that
// can be understood by docker
// '<HostPath>:<ContainerPath>[:options]', where 'options'
// is a comma-separated list of the following strings:
// 'ro', if the path is read only
// 'Z', if the volume requires SELinux relabeling
func generateMountBindings(mounts ...Mount) []string {
	result := make([]string, 0, len(mounts))
	for _, m := range mounts {
		bind := fmt.Sprintf("%s:%s", m.HostPath, m.ContainerPath)
		var attrs []string
		if m.Readonly {
			attrs = append(attrs, "ro")
		}
		// Only request relabeling if the pod provides an SELinux context. If the pod
		// does not provide an SELinux context relabeling will label the volume with
		// the container's randomly allocated MCS label. This would restrict access
		// to the volume to the container which mounts it first.
		if m.SelinuxRelabel {
			attrs = append(attrs, "Z")
		}
		switch m.Propagation {
		case MountPropagationNone:
			// noop, private is default
		case MountPropagationBidirectional:
			attrs = append(attrs, "rshared")
		case MountPropagationHostToContainer:
			attrs = append(attrs, "rslave")
		default:
			// Falls back to "private"
		}

		if len(attrs) > 0 {
			bind = fmt.Sprintf("%s:%s", bind, strings.Join(attrs, ","))
		}
		// our specific modification is the following line: make this a docker flag
		bind = fmt.Sprintf("--volume=%s", bind)
		result = append(result, bind)
	}
	return result
}

// UsernsRemap checks if userns-remap is enabled in dockerd
func UsernsRemap(ociBinary string) bool {
	cmd := exec.Command(ociBinary, "info", "--format", "'{{json .SecurityOptions}}'")
	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	err := cmd.Run()
	scanner := bufio.NewScanner(&buff)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err != nil {
		return false
	}
	if len(lines) > 0 {
		if strings.Contains(lines[0], "name=userns") {
			return true
		}
	}
	return false
}

func generatePortMappings(portMappings ...PortMapping) []string {
	result := make([]string, 0, len(portMappings))
	for _, pm := range portMappings {
		// let docker pick a host port by leaving it as ::
		// example --publish=127.0.0.17::8443 will get a random host port for 8443
		publish := fmt.Sprintf("--publish=%s::%d", pm.ListenAddress, pm.ContainerPort)
		result = append(result, publish)
	}
	return result
}

// createOpt is an option for Create
type createOpt func(*createOpts) *createOpts

// actual options struct
type createOpts struct {
	RunArgs       []string
	ContainerArgs []string
	Mounts        []Mount
	PortMappings  []PortMapping
}

// CreateContainer creates a container with "docker/podman run"
func CreateContainer(ociBinary string, image string, opts ...createOpt) ([]string, error) {
	o := &createOpts{}
	for _, opt := range opts {
		o = opt(o)
	}
	// convert mounts to container run args
	runArgs := o.RunArgs
	for _, mount := range o.Mounts {
		runArgs = append(runArgs, generateMountBindings(mount)...)
	}
	for _, portMapping := range o.PortMappings {
		runArgs = append(runArgs, generatePortMappings(portMapping)...)
	}
	// construct the actual docker run argv
	args := []string{"run"}
	args = append(args, runArgs...)
	args = append(args, image)
	args = append(args, o.ContainerArgs...)
	cmd := exec.Command(ociBinary, args...)
	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	err := cmd.Run()
	scanner := bufio.NewScanner(&buff)
	var output []string
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	if err != nil {
		return output, errors.Wrapf(err, "args: %v  output: %s ", args, output)
	}
	return output, nil
}

// WithRunArgs sets the args for docker run
// as in the args portion of `docker run args... image containerArgs...`
func WithRunArgs(args ...string) createOpt {
	return func(r *createOpts) *createOpts {
		r.RunArgs = args
		return r
	}
}

// WithMounts sets the container mounts
func WithMounts(mounts []Mount) createOpt {
	return func(r *createOpts) *createOpts {
		r.Mounts = mounts
		return r
	}
}

// WithPortMappings sets the container port mappings to the host
func WithPortMappings(portMappings []PortMapping) createOpt {
	return func(r *createOpts) *createOpts {
		r.PortMappings = portMappings
		return r
	}
}

// Copy copies a local asset into the container
func Copy(ociBinary string, ociID string, targetDir string, fName string) error {
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		return errors.Wrapf(err, "error source %s does not exist", fName)
	}
	destination := fmt.Sprintf("%s:%s", ociID, targetDir)
	cmd := exec.Command(ociBinary, "cp", fName, destination)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "error copying %s into node", fName)
	}
	return nil
}

// HostPortBinding will return port mapping for a container using cli.
// example : HostPortBinding("docker", "minikube", "22")
// will return the docker assigned port:
// 32769, nil
// only supports TCP ports
func HostPortBinding(ociBinary string, ociID string, contPort int) (int, error) {
	cmd := exec.Command(ociBinary, "inspect", "-f", fmt.Sprintf("'{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}'", contPort), ociID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "getting host-bind port %d for container ID %q, output %s", contPort, ociID, out)
	}
	o := strings.Trim(string(out), "\n")
	o = strings.Trim(o, "'")
	p, err := strconv.Atoi(o)
	if err != nil {
		return p, errors.Wrapf(err, "convert host-port %q to number", p)
	}
	return p, nil
}

// listContainersByLabel lists all the containres that kic driver created on user's machine using a label
// io.x-k8s.kic.cluster
func listContainersByLabel(ociBinary string, label string) ([]string, error) {
	cmd := exec.Command(ociBinary, "ps", "-a", "--filter", fmt.Sprintf("label=%s", label), "--format", "{{.Names}}")
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	var lines []string
	sc := bufio.NewScanner(&b)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, err
}

// ContainerIPs returns ipv4,ipv6, error of a container by their name
func ContainerIPs(ociBinary string, name string) (string, string, error) {
	// retrieve the IP address of the node using docker inspect
	lines, err := Inspect(ociBinary, name, "{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}")
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

// ListOwnedContainers lists all the containres that kic driver created on user's machine using a label
func ListOwnedContainers(ociBinary string) ([]string, error) {
	return listContainersByLabel(ociBinary, ClusterLabelKey)
}
