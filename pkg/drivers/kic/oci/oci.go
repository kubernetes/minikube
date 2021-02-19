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
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
)

// DeleteContainersByLabel deletes all containers that have a specific label
// if there no containers found with the given 	label, it will return nil
func DeleteContainersByLabel(ociBin string, label string) []error {
	var deleteErrs []error
	ctx := context.Background()
	cs, err := ListContainersByLabel(ctx, ociBin, label)
	if err != nil {
		return []error{fmt.Errorf("listing containers by label %q", label)}
	}

	if len(cs) == 0 {
		return nil
	}

	for _, c := range cs {
		_, err := ContainerStatus(ociBin, c)
		// only try to delete if docker/podman inspect returns
		// if it doesn't it means docker daemon is stuck and needs restart
		if err != nil {
			deleteErrs = append(deleteErrs, errors.Wrapf(err, "delete container %s: %s daemon is stuck. please try again!", c, ociBin))
			klog.Errorf("%s daemon seems to be stuck. Please try restarting your %s. :%v", ociBin, ociBin, err)
			continue
		}
		if err := ShutDown(ociBin, c); err != nil {
			klog.Infof("couldn't shut down %s (might be okay): %v ", c, err)
		}

		if _, err := runCmd(exec.Command(ociBin, "rm", "-f", "-v", c)); err != nil {
			deleteErrs = append(deleteErrs, errors.Wrapf(err, "delete container %s: output %s", c, err))
		}

	}
	return deleteErrs
}

// DeleteContainer deletes a container by ID or Name
func DeleteContainer(ctx context.Context, ociBin string, name string) error {
	_, err := ContainerStatus(ociBin, name)
	if err == context.DeadlineExceeded {
		out.WarningT("{{.ocibin}} is taking an unsually long time to respond, consider restarting {{.ocibin}}", out.V{"ociBin": ociBin})
	} else if err != nil {
		klog.Warningf("error getting container status, will try to delete anyways: %v", err)
	}
	// try to delete anyways
	if err := ShutDown(ociBin, name); err != nil {
		klog.Infof("couldn't shut down %s (might be okay): %v ", name, err)
	}

	if _, err := runCmd(exec.CommandContext(ctx, ociBin, "rm", "-f", "-v", name)); err != nil {
		return errors.Wrapf(err, "delete %s", name)
	}
	return nil
}

// PrepareContainerNode sets up the container node before CreateContainerNode is called.
// For the container runtime, it creates a volume which will be mounted into kic
func PrepareContainerNode(p CreateParams) error {
	if err := createVolume(p.OCIBinary, p.Name, p.Name); err != nil {
		return errors.Wrapf(err, "creating volume for %s container", p.Name)
	}
	klog.Infof("Successfully created a %s volume %s", p.OCIBinary, p.Name)
	if err := prepareVolumeSideCar(p.OCIBinary, p.Image, p.Name); err != nil {
		return errors.Wrapf(err, "preparing volume for %s container", p.Name)
	}
	klog.Infof("Successfully prepared a %s volume %s", p.OCIBinary, p.Name)
	return nil
}

func hasMemoryCgroup() bool {
	memcg := true
	if runtime.GOOS == "linux" {
		var memory string
		if cgroup2, err := IsCgroup2UnifiedMode(); err == nil && cgroup2 {
			memory = "/sys/fs/cgroup/memory/memsw.limit_in_bytes"
		}
		if _, err := os.Stat(memory); os.IsNotExist(err) {
			klog.Warning("Your kernel does not support memory limit capabilities or the cgroup is not mounted.")
			memcg = false
		}
	}
	return memcg
}

func hasMemorySwapCgroup() bool {
	memcgSwap := true
	if runtime.GOOS == "linux" {
		var memoryswap string
		if cgroup2, err := IsCgroup2UnifiedMode(); err == nil && cgroup2 {
			memoryswap = "/sys/fs/cgroup/memory/memory.swap.max"
		}
		if _, err := os.Stat(memoryswap); os.IsNotExist(err) {
			// requires CONFIG_MEMCG_SWAP_ENABLED or cgroup_enable=memory in grub
			klog.Warning("Your kernel does not support swap limit capabilities or the cgroup is not mounted.")
			memcgSwap = false
		}
	}
	return memcgSwap
}

// CreateContainerNode creates a new container node
func CreateContainerNode(p CreateParams) error {
	// on windows os, if docker desktop is using Windows Containers. Exit early with error
	if p.OCIBinary == Docker && runtime.GOOS == "windows" {
		info, err := DaemonInfo(p.OCIBinary)
		if info.OSType == "windows" {
			return ErrWindowsContainers
		}
		if err != nil {
			klog.Warningf("error getting dameon info: %v", err)
			return errors.Wrap(err, "daemon info")
		}
	}

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
	// to provide a static IP
	if p.Network != "" && p.IP != "" {
		runArgs = append(runArgs, "--network", p.Network)
		runArgs = append(runArgs, "--ip", p.IP)
	}

	memcgSwap := hasMemorySwapCgroup()
	memcg := hasMemoryCgroup()

	if !memcgSwap || !memcg {
		out.WarningT("Cgroup v2 does not allow setting memory, if you want to set memory, please modify your Grub as instructed in https://docs.docker.com/engine/install/linux-postinstall/#your-kernel-does-not-support-cgroup-swap-limit-capabilities")
	}

	// https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
	var virtualization string
	if p.OCIBinary == Podman { // enable execing in /var
		// podman mounts var/lib with no-exec by default  https://github.com/containers/libpod/issues/5103
		runArgs = append(runArgs, "--volume", fmt.Sprintf("%s:/var:exec", p.Name))

		if memcgSwap {
			runArgs = append(runArgs, fmt.Sprintf("--memory-swap=%s", p.Memory))
		}

		if memcg {
			runArgs = append(runArgs, fmt.Sprintf("--memory=%s", p.Memory))
		}

		virtualization = "podman" // VIRTUALIZATION_PODMAN
	}
	if p.OCIBinary == Docker {
		runArgs = append(runArgs, "--volume", fmt.Sprintf("%s:/var", p.Name))
		// ignore apparmore github actions docker: https://github.com/kubernetes/minikube/issues/7624
		runArgs = append(runArgs, "--security-opt", "apparmor=unconfined")

		if memcg {
			runArgs = append(runArgs, fmt.Sprintf("--memory=%s", p.Memory))
		}
		if memcgSwap {
			// Disable swap by setting the value to match
			runArgs = append(runArgs, fmt.Sprintf("--memory-swap=%s", p.Memory))
		}

		virtualization = "docker" // VIRTUALIZATION_DOCKER
	}

	cpuCfsPeriod := true
	cpuCfsQuota := true
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/sys/fs/cgroup/cpu/cpu.cfs_period_us"); os.IsNotExist(err) {
			cpuCfsPeriod = false
		}
		if _, err := os.Stat("/sys/fs/cgroup/cpu/cpu.cfs_quota_us"); os.IsNotExist(err) {
			cpuCfsQuota = false
		}
		if !cpuCfsPeriod || !cpuCfsQuota {
			// requires CONFIG_CFS_BANDWIDTH
			klog.Warning("Your kernel does not support CPU cfs period/quota or the cgroup is not mounted.")
		}
	}

	if cpuCfsPeriod && cpuCfsQuota {
		runArgs = append(runArgs, fmt.Sprintf("--cpus=%s", p.CPUs))
	}

	runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", "container", virtualization))

	for key, val := range p.Envs {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	// adds node specific args
	runArgs = append(runArgs, p.ExtraArgs...)

	if enabled := isUsernsRemapEnabled(p.OCIBinary); enabled {
		// We need this argument in order to make this command work
		// in systems that have userns-remap enabled on the docker daemon
		runArgs = append(runArgs, "--userns=host")
	}

	if err := createContainer(p.OCIBinary, p.Image, withRunArgs(runArgs...), withMounts(p.Mounts), withPortMappings(p.PortMappings)); err != nil {
		return errors.Wrap(err, "create container")
	}

	checkRunning := func() error {
		r, err := ContainerRunning(p.OCIBinary, p.Name)
		if err != nil {
			return fmt.Errorf("temporary error checking running for %q : %v", p.Name, err)
		}
		if !r {
			return fmt.Errorf("temporary error created container %q is not running yet", p.Name)
		}
		s, err := ContainerStatus(p.OCIBinary, p.Name)
		if err != nil {
			return fmt.Errorf("temporary error checking status for %q : %v", p.Name, err)
		}
		if s != state.Running {
			return fmt.Errorf("temporary error created container %q is not running yet", p.Name)
		}
		if !iptablesFileExists(p.OCIBinary, p.Name) {
			return fmt.Errorf("iptables file doesn't exist, see #8179")
		}
		klog.Infof("the created container %q has a running status.", p.Name)
		return nil
	}

	if err := retry.Expo(checkRunning, 15*time.Millisecond, 25*time.Second); err != nil {
		excerpt := LogContainerDebug(p.OCIBinary, p.Name)
		_, err := DaemonInfo(p.OCIBinary)
		if err != nil {
			return errors.Wrapf(ErrDaemonInfo, "container name %q", p.Name)
		}

		return errors.Wrapf(ErrExitedUnexpectedly, "container name %q: log: %s", p.Name, excerpt)
	}

	return nil
}

// CreateContainer creates a container with "docker/podman run"
func createContainer(ociBin string, image string, opts ...createOpt) error {
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

	// to run nested container from privileged container in podman https://bugzilla.redhat.com/show_bug.cgi?id=1687713
	// only add when running locally (linux), when running remotely it needs to be configured on server in libpod.conf
	if ociBin == Podman && runtime.GOOS == "linux" {
		args = append(args, "--cgroup-manager", "cgroupfs")
	}

	args = append(args, runArgs...)
	args = append(args, image)
	args = append(args, o.ContainerArgs...)

	if rr, err := runCmd(exec.Command(ociBin, args...)); err != nil {
		// full error: docker: Error response from daemon: Range of CPUs is from 0.01 to 8.00, as there are only 8 CPUs available.
		if strings.Contains(rr.Output(), "Range of CPUs is from") && strings.Contains(rr.Output(), "CPUs available") { // CPUs available
			return ErrCPUCountLimit
		}
		// example: docker: Error response from daemon: Address already in use.
		if strings.Contains(rr.Output(), "Address already in use") {
			return ErrIPinUse
		}
		return err
	}

	return nil
}

// StartContainer starts a container with "docker/podman start"
func StartContainer(ociBin string, container string) error {
	// construct the actual docker start argv
	args := []string{"start"}

	// to run nested container from privileged container in podman https://bugzilla.redhat.com/show_bug.cgi?id=1687713
	// only add when running locally (linux), when running remotely it needs to be configured on server in libpod.conf
	if ociBin == Podman && runtime.GOOS == "linux" {
		args = append(args, "--cgroup-manager", "cgroupfs")
	}

	args = append(args, container)

	if _, err := runCmd(exec.Command(ociBin, args...)); err != nil {
		return err
	}

	return nil
}

// ContainerID returns id of a container name
func ContainerID(ociBin string, nameOrID string) (string, error) {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", "-f", "{{.Id}}", nameOrID))
	if err != nil { // don't return error if not found, only return empty string
		if strings.Contains(rr.Stdout.String(), "Error: No such object:") ||
			strings.Contains(rr.Stdout.String(), "Error: No such container:") ||
			strings.Contains(rr.Stdout.String(), "unable to find") ||
			strings.Contains(rr.Stdout.String(), "Error: error inspecting object") ||
			strings.Contains(rr.Stdout.String(), "Error: error looking up container") ||
			strings.Contains(rr.Stdout.String(), "no such container") {
			err = nil
		}
		return "", err
	}
	return rr.Stdout.String(), nil
}

// ContainerExists checks if container name exists (either running or exited)
func ContainerExists(ociBin string, name string, warnSlow ...bool) (bool, error) {
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
func IsCreatedByMinikube(ociBin string, nameOrID string) bool {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", nameOrID, "--format", "{{.Config.Labels}}"))
	if err != nil {
		return false
	}

	if strings.Contains(rr.Stdout.String(), fmt.Sprintf("%s:true", CreatedByLabelKey)) {
		return true
	}

	return false
}

// ListOwnedContainers lists all the containres that kic driver created on user's machine using a label
func ListOwnedContainers(ociBin string) ([]string, error) {
	return ListContainersByLabel(context.Background(), ociBin, ProfileLabelKey)
}

// inspect return low-level information on containers
func inspect(ociBin string, containerNameOrID, format string) ([]string, error) {
	cmd := exec.Command(ociBin, "container", "inspect",
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
func isUsernsRemapEnabled(ociBin string) bool {
	cmd := exec.Command(ociBin, "info", "--format", "'{{json .SecurityOptions}}'")
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
func ListContainersByLabel(ctx context.Context, ociBin string, label string, warnSlow ...bool) ([]string, error) {
	rr, err := runCmd(exec.CommandContext(ctx, ociBin, "ps", "-a", "--filter", fmt.Sprintf("label=%s", label), "--format", "{{.Names}}"), warnSlow...)
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
	if p := os.Getenv(constants.MinikubeActiveDockerdEnv); p != "" {
		klog.Infof("shell is pointing to dockerd inside minikube. will unset to use host")
		for _, e := range constants.DockerDaemonEnvs {
			if err := resetEnv(e); err != nil {
				return err
			}
		}
	}
	return nil
}

func resetEnv(key string) error {
	v := os.Getenv(constants.MinikubeExistingPrefix + key)
	if v == "" {
		if err := os.Unsetenv(key); err != nil {
			return errors.Wrapf(err, "resetting %s env", key)
		}
		return nil
	}
	if err := os.Setenv(key, v); err != nil {
		return errors.Wrapf(err, "resetting %s env", key)
	}
	return nil
}

// PointToHostPodman will unset env variables that point to podman inside minikube
func PointToHostPodman() error {
	if p := os.Getenv(constants.MinikubeActivePodmanEnv); p != "" {
		klog.Infof("shell is pointing to podman inside minikube. will unset to use host")
		for _, e := range constants.PodmanRemoteEnvs {
			if err := resetEnv(e); err != nil {
				return err
			}
		}
	}
	return nil
}

// ContainerRunning returns running state of a container
func ContainerRunning(ociBin string, name string, warnSlow ...bool) (bool, error) {
	rr, err := runCmd(exec.Command(ociBin, "container", "inspect", name, "--format={{.State.Running}}"), warnSlow...)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(rr.Stdout.String()))
}

// ContainerStatus returns status of a container running,exited,...
func ContainerStatus(ociBin string, name string, warnSlow ...bool) (state.State, error) {
	cmd := exec.Command(ociBin, "container", "inspect", name, "--format={{.State.Status}}")
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
func ShutDown(ociBin string, name string) error {
	if _, err := runCmd(exec.Command(ociBin, "exec", "--privileged", "-t", name, "/bin/bash", "-c", "sudo init 0")); err != nil {
		klog.Infof("error shutdown %s: %v", name, err)
	}
	// helps with allowing docker realize the container is exited and report its status correctly.
	time.Sleep(time.Second * 1)
	// wait till it is stoped
	stopped := func() error {
		st, err := ContainerStatus(ociBin, name)
		if st == state.Stopped {
			klog.Infof("container %s status is %s", name, st)
			return nil
		}
		if err != nil {
			klog.Infof("temporary error verifying shutdown: %v", err)
		}
		klog.Infof("temporary error: container %s status is %s but expect it to be exited", name, st)
		return errors.Wrap(err, "couldn't verify container is exited. %v")
	}
	if err := retry.Expo(stopped, time.Millisecond*500, time.Second*20); err != nil {
		return errors.Wrap(err, "verify shutdown")
	}
	klog.Infof("Successfully shutdown container %s", name)
	return nil
}

// iptablesFileExists checks if /var/lib/dpkg/alternatives/iptables exists in minikube
// this file is necessary for the entrypoint script to pass
// TODO: https://github.com/kubernetes/minikube/issues/8179
func iptablesFileExists(ociBin string, nameOrID string) bool {
	file := "/var/lib/dpkg/alternatives/iptables"
	_, err := runCmd(exec.Command(ociBin, "exec", nameOrID, "stat", file), false)
	if err != nil {
		klog.Warningf("error checking if %s exists: %v", file, err)
		return false
	}
	return true
}

// DaemonHost returns the ip/hostname where OCI daemon service for driver is running
// For Podman return the host part of CONTAINER_HOST environment variable if set
// For Docker return the host part of DOCKER_HOST environment variable if set
// or DefaultBindIPV4 otherwise
func DaemonHost(driver string) string {
	if driver == Podman {
		if dh := os.Getenv(constants.PodmanContainerHostEnv); dh != "" {
			if u, err := url.Parse(dh); err == nil {
				if u.Host != "" {
					return u.Hostname()
				}
			}
		}
	}
	if driver == Docker {
		if dh := os.Getenv(constants.DockerHostEnv); dh != "" {
			if u, err := url.Parse(dh); err == nil {
				if u.Host != "" {
					return u.Hostname()
				}
			}
		}
	}
	return DefaultBindIPV4
}

// IsExternalDaemonHost returns whether or not the OCI runtime is running on an external/virtual host
// For Podman driver return true if CONTAINER_HOST is set to a URI, and the URI contains a host item
// For Docker driver return true if DOCKER_HOST is set to a URI, and the URI contains a host item
func IsExternalDaemonHost(driver string) bool {
	if driver == Podman {
		if dh := os.Getenv(constants.PodmanContainerHostEnv); dh != "" {
			if u, err := url.Parse(dh); err == nil {
				return u.Host != ""
			}
		}
	}
	if driver == Docker {
		if dh := os.Getenv(constants.DockerHostEnv); dh != "" {
			if u, err := url.Parse(dh); err == nil {
				return u.Host != ""
			}
		}
	}
	return false
}
