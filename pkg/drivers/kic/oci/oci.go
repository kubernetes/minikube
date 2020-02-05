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

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"

	"fmt"
	"os/exec"
	"strings"
)

// CreateContainerNode creates a new container node
func CreateContainerNode(p CreateParams) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}

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
		"--label", fmt.Sprintf("%s=%s", nodeRoleKey, p.Role),
	}

	for key, val := range p.Envs {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	// adds node specific args
	runArgs = append(runArgs, p.ExtraArgs...)

	enabled, err := isUsernsRemapEnabled(p.OCIBinary)
	if err != nil {
		glog.Warningf("Failed to detect if userns is enabled: %v", err)
	}
	if enabled {
		// We need this argument in order to make this command work
		// in systems that have userns-remap enabled on the docker daemon
		runArgs = append(runArgs, "--userns=host")
	}

	_, err = createContainer(p.OCIBinary,
		p.Image,
		withRunArgs(runArgs...),
		withMounts(p.Mounts),
		withPortMappings(p.PortMappings),
	)
	if err != nil {
		return errors.Wrap(err, "create a kic node")
	}
	return nil
}

// CreateContainer creates a container with "docker/podman run"
func createContainer(ociBinary string, image string, opts ...createOpt) ([]string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return nil, errors.Wrap(err, "point host docker-daemon")
	}

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

// Copy copies a local asset into the container
func Copy(ociBinary string, ociID string, targetDir string, fName string) error {
	if err := PointToHostDockerDaemon(); err != nil {
		return errors.Wrap(err, "point host docker-daemon")
	}
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
	if err := PointToHostDockerDaemon(); err != nil {
		return 0, errors.Wrap(err, "point host docker-daemon")
	}
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

// ContainerIPs returns ipv4,ipv6, error of a container by their name
func ContainerIPs(ociBinary string, name string) (string, string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return "", "", errors.Wrap(err, "point host docker-daemon")
	}
	// retrieve the IP address of the node using docker inspect
	lines, err := inspect(ociBinary, name, "{{range .NetworkSettings.Networks}}{{.IPAddress}},{{.GlobalIPv6Address}}{{end}}")
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

// ContainerID returns id of a container name
func ContainerID(ociBinary string, nameOrID string) (string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return "", errors.Wrap(err, "point host docker-daemon")
	}
	cmd := exec.Command(ociBinary, "inspect", "-f", "{{.Id}}", nameOrID)
	id, err := cmd.CombinedOutput()
	if err != nil {
		id = []byte{}
	}
	return string(id), err
}

// ListOwnedContainers lists all the containres that kic driver created on user's machine using a label
func ListOwnedContainers(ociBinary string) ([]string, error) {
	return listContainersByLabel(ociBinary, ClusterLabelKey)
}

// inspect return low-level information on containers
func inspect(ociBinary string, containerNameOrID, format string) ([]string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return nil, errors.Wrap(err, "point host docker-daemon")
	}
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

// isUsernsRemapEnabled checks if userns-remap is enabled in docker
func isUsernsRemapEnabled(ociBinary string) (bool, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return false, errors.Wrap(err, "point host docker-daemon")
	}
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
		return false, nil
	}
	if len(lines) > 0 {
		if strings.Contains(lines[0], "name=userns") {
			return true, nil
		}
	}
	return false, nil
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

// withRunArgs sets the args for docker run
// as in the args portion of `docker run args... image containerArgs...`
func withRunArgs(args ...string) createOpt {
	return func(r *createOpts) *createOpts {
		r.RunArgs = args
		return r
	}
}

// withMounts sets the container mounts
func withMounts(mounts []Mount) createOpt {
	return func(r *createOpts) *createOpts {
		r.Mounts = mounts
		return r
	}
}

// withPortMappings sets the container port mappings to the host
func withPortMappings(portMappings []PortMapping) createOpt {
	return func(r *createOpts) *createOpts {
		r.PortMappings = portMappings
		return r
	}
}

// listContainersByLabel lists all the containres that kic driver created on user's machine using a label
// io.x-k8s.kic.cluster
func listContainersByLabel(ociBinary string, label string) ([]string, error) {
	if err := PointToHostDockerDaemon(); err != nil {
		return nil, errors.Wrap(err, "point host docker-daemon")
	}
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

// PointToHostDockerDaemon will unset env variables that point to docker inside minikube
// to make sure it points to the docker daemon installed by user.
func PointToHostDockerDaemon() error {
	p := os.Getenv(constants.MinikubeActiveDockerdEnv)
	if p != "" {
		glog.Infof("shell is pointing to docker inside minikube. will unset to use host")
	}

	for i := range constants.DockerDaemonEnvs {
		e := constants.DockerDaemonEnvs[i]
		err := os.Setenv(e, "")
		if err != nil {
			return errors.Wrapf(err, "resetting %s env", e)
		}

	}
	return nil
}
