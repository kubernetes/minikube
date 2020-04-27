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
	"time"

	"bufio"
	"bytes"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util/retry"

	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// DeleteContainersByLabel deletes all containers that have a specific label
// if there no containers found with the given 	label, it will return nil
func DeleteContainersByLabel(prefix string, ociBin string, label string) []error {
	var deleteErrs []error

	cs, err := ListContainersByLabel(prefix, ociBin, label)
	if err != nil {
		return []error{fmt.Errorf("listing containers by label %q", label)}
	}

	if len(cs) == 0 {
		return nil
	}

	for _, c := range cs {
		_, err := ContainerStatus(prefix, ociBin, c)
		// only try to delete if docker/podman inspect returns
		// if it doesn't it means docker daemon is stuck and needs restart
		if err != nil {
			deleteErrs = append(deleteErrs, errors.Wrapf(err, "delete container %s: %s daemon is stuck. please try again!", c, ociBin))
			glog.Errorf("%s daemon seems to be stuck. Please try restarting your %s. :%v", ociBin, ociBin, err)
			continue
		}
		if err := ShutDown(prefix, ociBin, c); err != nil {
			glog.Infof("couldn't shut down %s (might be okay): %v ", c, err)
		}

		if _, err := runCmd(exec.Command(prefix, ociBin, "rm", "-f", "-v", c)); err != nil {
			deleteErrs = append(deleteErrs, errors.Wrapf(err, "delete container %s: output %s", c, err))
		}

	}
	return deleteErrs
}

// DeleteContainer deletes a container by ID or Name
func DeleteContainer(prefix string, ociBin string, name string) error {

	_, err := ContainerStatus(prefix, ociBin, name)
	if err != nil {
		glog.Errorf("%s daemon seems to be stuck. Please try restarting your %s. Will try to delete anyways: %v", ociBin, ociBin, err)
	}
	// try to delete anyways
	if err := ShutDown(prefix, ociBin, name); err != nil {
		glog.Infof("couldn't shut down %s (might be okay): %v ", name, err)
	}

	if _, err := runCmd(exec.Command(prefix, ociBin, "rm", "-f", "-v", name)); err != nil {
		return errors.Wrapf(err, "delete %s", name)
	}
	return nil
}

// PrepareContainerNode sets up the container node before CreateContainerNode is called.
// For the docker runtime, it creates a docker volume which will be mounted into kic
func PrepareContainerNode(p CreateParams) error {
	if p.OCIBinary != Docker {
		return nil
	}
	if err := createDockerVolume(p.Name, p.Name); err != nil {
		return errors.Wrapf(err, "creating volume for %s container", p.Name)
	}
	glog.Infof("Successfully created a docker volume %s", p.Name)
	return nil
}

// CreateContainerNode creates a new container node
func CreateContainerNode(p CreateParams) error {
	runArgs := []string{
		"-d", // run the container detached
		"-t", // allocate a tty for entrypoint logs
		// running containers in a container requires privileged
		// NOTE: we could try to replicate this with --cap-add, and use less
		// privileges, but this flag also changes some mounts that are necessary
		// including some ones docker would otherwise do by default.
		// for now this is what we want. in the future we may revisit this.
		"--privileged",
		"--security-opt", "seccomp=unconfined", //  ignore seccomp
		// ignore apparmore github actions docker: https://github.com/kubernetes/minikube/issues/7624
		"--security-opt", "apparmor=unconfined",
		"--tmpfs", "/tmp", // various things depend on working /tmp
		"--tmpfs", "/run", // systemd wants a writable /run
		// logs,pods be stroed on  filesystem vs inside container,
		// some k8s things want /lib/modules
		"-v", "/lib/modules:/lib/modules:ro",
		"--hostname", p.Name, // make hostname match container name
		"--name", p.Name, // ... and set the container name
		"--label", fmt.Sprintf("%s=%s", CreatedByLabelKey, "true"),
		// label the node with the cluster ID
		"--label", p.ClusterLabel,
		// label the node with the role ID
		"--label", fmt.Sprintf("%s=%s", nodeRoleLabelKey, p.Role),
		// label th enode wuth the node ID
		"--label", p.NodeLabel,
	}

	if p.OCIBinary == Podman { // enable execing in /var
		// podman mounts var/lib with no-exec by default  https://github.com/containers/libpod/issues/5103
		runArgs = append(runArgs, "--volume", fmt.Sprintf("%s:/var:exec", p.Name))
	}
	if p.OCIBinary == Docker {
		runArgs = append(runArgs, "--volume", fmt.Sprintf("%s:/var", p.Name))
	}

	runArgs = append(runArgs, fmt.Sprintf("--cpus=%s", p.CPUs))

	memcgSwap := true
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/sys/fs/cgroup/memory/memsw.limit_in_bytes"); os.IsNotExist(err) {
			// requires CONFIG_MEMCG_SWAP_ENABLED or cgroup_enable=memory in grub
			glog.Warning("Your kernel does not support swap limit capabilities or the cgroup is not mounted.")
			memcgSwap = false
		}
	}

	if p.OCIBinary == Podman && memcgSwap { // swap is required for memory
		runArgs = append(runArgs, fmt.Sprintf("--memory=%s", p.Memory))
	}
	if p.OCIBinary == Docker { // swap is only required for --memory-swap
		runArgs = append(runArgs, fmt.Sprintf("--memory=%s", p.Memory))
	}

	// https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
	var virtualization string
	if p.OCIBinary == Podman {
		virtualization = "podman" // VIRTUALIZATION_PODMAN
	}
	if p.OCIBinary == Docker {
		virtualization = "docker" // VIRTUALIZATION_DOCKER
	}
	runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", "container", virtualization))

	for key, val := range p.Envs {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	// adds node specific args
	runArgs = append(runArgs, p.ExtraArgs...)

	if enabled := isUsernsRemapEnabled(p.OCIPrefix, p.OCIBinary); enabled {
		// We need this argument in order to make this command work
		// in systems that have userns-remap enabled on the docker daemon
		runArgs = append(runArgs, "--userns=host")
	}

	if err := createContainer(p.OCIPrefix, p.OCIBinary, p.Image, withRunArgs(runArgs...), withMounts(p.Mounts), withPortMappings(p.PortMappings)); err != nil {
		return errors.Wrap(err, "create container")
	}

	checkRunning := func() error {
		r, err := ContainerRunning(p.OCIPrefix, p.OCIBinary, p.Name)
		if err != nil {
			return fmt.Errorf("temporary error checking running for %q : %v", p.Name, err)
		}
		if !r {
			return fmt.Errorf("temporary error created container %q is not running yet", p.Name)
		}
		s, err := ContainerStatus(p.OCIPrefix, p.OCIBinary, p.Name)
		if err != nil {
			return fmt.Errorf("temporary error checking status for %q : %v", p.Name, err)
		}
		if s != state.Running {
			return fmt.Errorf("temporary error created container %q is not running yet", p.Name)
		}
		glog.Infof("the created container %q has a running status.", p.Name)
		return nil
	}

	// retry up to up 13 seconds to make sure the created container status is running.
	if err := retry.Expo(checkRunning, 13*time.Millisecond, time.Second*13); err != nil {
		return errors.Wrapf(err, "check container %q running", p.Name)
	}

	return nil
}

// CreateContainer creates a container with "docker/podman run"
func createContainer(prefix string, ociBin string, image string, opts ...createOpt) error {
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
	args := []string{ociBin, "run"}

	// to run nested container from privileged container in podman https://bugzilla.redhat.com/show_bug.cgi?id=1687713
	if ociBin == Podman {
		args = append(args, "--cgroup-manager", "cgroupfs")
	}

	args = append(args, runArgs...)
	args = append(args, image)
	args = append(args, o.ContainerArgs...)

	if _, err := runCmd(exec.Command(prefix, args...)); err != nil {
		return err
	}

	return nil
}

// ContainerID returns id of a container name
func ContainerID(prefix, ociBin string, nameOrID string) (string, error) {
	rr, err := runCmd(exec.Command(prefix, ociBin, "inspect", "-f", "{{.Id}}", nameOrID))
	if err != nil { // don't return error if not found, only return empty string
		if strings.Contains(rr.Stdout.String(), "Error: No such object:") || strings.Contains(rr.Stdout.String(), "unable to find") {
			err = nil
		}
		return "", err
	}
	return rr.Stdout.String(), nil
}

// ContainerExists checks if container name exists (either running or exited)
func ContainerExists(prefix string, ociBin string, name string, warnSlow ...bool) (bool, error) {
	rr, err := runCmd(exec.Command(ociBin, "ps", "-a", "--format", "{{.Names}}"), warnSlow...)
	if err != nil {
		return false, err
	}

	containers := strings.Split(rr.Stdout.String(), "\n")
	for _, c := range containers {
		if strings.TrimSpace(c) == name {
			return true, nil
		}
	}

	return false, nil
}

// IsCreatedByMinikube returns true if the container was created by minikube
// with default assumption that it is not created by minikube when we don't know for sure
func IsCreatedByMinikube(prefix, ociBin string, nameOrID string) bool {
	rr, err := runCmd(exec.Command(prefix, ociBin, "inspect", nameOrID, "--format", "{{.Config.Labels}}"))
	if err != nil {
		return false
	}

	if strings.Contains(rr.Stdout.String(), fmt.Sprintf("%s:true", CreatedByLabelKey)) {
		return true
	}

	return false
}

func ListOwnedContainers(prefix string, ociBin string) ([]string, error) {
	return ListContainersByLabel(prefix, ociBin, ProfileLabelKey)
}

// inspect return low-level information on containers
func inspect(prefix string, ociBin string, containerNameOrID, format string) ([]string, error) {
	cmd := exec.Command(prefix, ociBin, "inspect",
		"-f", format,
		containerNameOrID) // ... against the "node" container
	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	_, err := runCmd(cmd)
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
func isUsernsRemapEnabled(prefix string, ociBin string) bool {
	cmd := exec.Command(prefix, ociBin, "info", "--format", "'{{json .SecurityOptions}}'")
	var buff bytes.Buffer
	cmd.Stdout = &buff
	cmd.Stderr = &buff

	if _, err := runCmd(cmd); err != nil {
		return false
	}

	scanner := bufio.NewScanner(&buff)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
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

// ListContainersByLabel returns all the container names with a specified label
func ListContainersByLabel(prefix string, ociBin string, label string, warnSlow ...bool) ([]string, error) {
	rr, err := runCmd(exec.Command(prefix, ociBin, "ps", "-a", "--filter", fmt.Sprintf("label=%s", label), "--format", "{{.Names}}"), warnSlow...)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(bytes.NewReader(rr.Stdout.Bytes()))
	var names []string
	for s.Scan() {
		n := strings.TrimSpace(s.Text())
		if n != "" {
			names = append(names, n)
		}
	}
	return names, err
}

// PointToHostDockerDaemon will unset env variables that point to docker inside minikube
// to make sure it points to the docker daemon installed by user.
func PointToHostDockerDaemon() error {
	p := os.Getenv(constants.MinikubeActiveDockerdEnv)
	if p != "" {
		glog.Infof("shell is pointing to dockerd inside minikube. will unset to use host")
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

// ContainerRunning returns running state of a container
func ContainerRunning(prefix string, ociBin string, name string, warnSlow ...bool) (bool, error) {
	rr, err := runCmd(exec.Command(prefix, ociBin, "inspect", name, "--format={{.State.Running}}"), warnSlow...)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(rr.Stdout.String()))
}

// ContainerStatus returns status of a container running,exited,...
func ContainerStatus(prefix string, ociBin string, name string, warnSlow ...bool) (state.State, error) {
	cmd := exec.Command(prefix, ociBin, "inspect", name, "--format={{.State.Status}}")
	rr, err := runCmd(cmd, warnSlow...)
	o := strings.TrimSpace(rr.Stdout.String())
	switch o {
	case "configured":
		return state.Stopped, nil
	case "running":
		return state.Running, nil
	case "exited":
		return state.Stopped, nil
	case "paused":
		return state.Paused, nil
	case "restarting":
		return state.Starting, nil
	case "dead":
		return state.Error, nil
	default:
		return state.None, errors.Wrapf(err, "unknown state %q", name)
	}
}

// ShutDown will run command to shut down the container
// to ensure the containers process and networking bindings are all closed
// to avoid containers getting stuck before delete https://github.com/kubernetes/minikube/issues/7657
func ShutDown(prefix string, ociBin string, name string) error {
	if _, err := runCmd(exec.Command(prefix, ociBin, "exec", "--privileged", "-t", name, "/bin/bash", "-c", "sudo init 0")); err != nil {
		glog.Infof("error shutdown %s: %v", name, err)
	}
	// helps with allowing docker realize the container is exited and report its status correctly.
	time.Sleep(time.Second * 1)
	// wait till it is stoped
	stopped := func() error {
		st, err := ContainerStatus(prefix, ociBin, name)
		if st == state.Stopped {
			glog.Infof("container %s status is %s", name, st)
			return nil
		}
		if err != nil {
			glog.Infof("temporary error verifying shutdown: %v", err)
		}
		glog.Infof("temporary error: container %s status is %s but expect it to be exited", name, st)
		return errors.Wrap(err, "couldn't verify cointainer is exited. %v")
	}
	if err := retry.Expo(stopped, time.Millisecond*500, time.Second*20); err != nil {
		return errors.Wrap(err, "verify shutdown")
	}
	glog.Infof("Successfully shutdown container %s", name)
	return nil
}
