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

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/assets"

	"bufio"
	"bytes"

	"github.com/pkg/errors"

	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
)

// Stop stops a container
func Stop(ociBinary, ociID string) error {
	cmd := exec.Command(ociBinary, "stop", ociID)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "error stop node %s", ociID)
	}

	return nil
}

// Status returns the status of the container
func Status(ociBinary string, ociID string) (state.State, error) {
	cmd := exec.Command(ociBinary, "inspect", "-f", "{{.State.Status}}", ociID)
	out, err := cmd.CombinedOutput()
	o := strings.Trim(string(out), "\n")
	s := state.Error
	switch o {
	case "running":
		s = state.Running
	case "exited":
		s = state.Stopped
	case "paused":
		s = state.Paused
	case "restaring":
		s = state.Starting
	}

	if err != nil {
		return state.Error, errors.Wrapf(err, "error getting node %s status", ociID)
	}
	return s, nil
}

// SystemStatus checks if the oci container engine is running
func SystemStatus(ociBinary string, ociID string) (state.State, error) {
	_, err := exec.LookPath(ociBinary)
	if err != nil {
		return state.Error, err
	}

	err = exec.Command("docker", "info").Run()
	if err != nil {
		return state.Error, err
	}

	return state.Running, nil
}

// Remove removes a container
func Remove(ociBinary string, ociID string) error {
	// TODO: force remove should be an option
	cmd := exec.Command(ociBinary, "rm", "-f", "-v", ociID)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error removing node %s", ociID)
	}

	return nil
}

// Pause pauses a container
func Pause(ociBinary string, ociID string) error {
	cmd := exec.Command(ociBinary, "pause", ociID)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "error pausing node %s", ociID)
	}

	return nil
}

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

// NetworkInspect displays detailed information on one or more networks
func NetworkInspect(networkNames []string, format string) ([]string, error) {
	cmd := exec.Command("docker", "network", "inspect",
		"-f", format,
		strings.Join(networkNames, " "))
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

// GetSubnets returns a slice of subnets for a specified network name
// For example the command : docker network inspect -f '{{range (index (index . "IPAM") "Config")}}{{index . "Subnet"}} {{end}}' bridge
// returns 172.17.0.0/16
func GetSubnets(networkName string) ([]string, error) {
	format := `{{range (index (index . "IPAM") "Config")}}{{index . "Subnet"}} {{end}}`
	lines, err := NetworkInspect([]string{networkName}, format)
	if err != nil {
		return nil, err
	}
	return strings.Split(lines[0], " "), nil
}

// ImageInspect return low-level information on containers images
func ImageInspect(containerNameOrID, format string) ([]string, error) {
	cmd := exec.Command("docker", "image", "inspect",
		"-f", format,
		containerNameOrID,
	)
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

// ImageID return the Id of the container image
func ImageID(containerNameOrID string) (string, error) {
	lines, err := ImageInspect(containerNameOrID, "{{ .Id }}")
	if err != nil {
		return "", err
	}
	if len(lines) != 1 {
		return "", fmt.Errorf("docker image ID should only be one line, got %d lines", len(lines))
	}
	return lines[0], nil
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

// PullIfNotPresent pulls docker image if not present back off exponentially
func PullIfNotPresent(ociBinary string, image string, forceUpdate bool, maxWait time.Duration) error {
	cmd := exec.Command(ociBinary, "inspect", "--type=image", image)
	err := cmd.Run()
	if err == nil && !forceUpdate {
		return nil // if presents locally and not force
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxWait
	f := func() error {
		return pull(ociBinary, image)
	}
	return backoff.Retry(f, b)
}

// Pull pulls an image, retrying up to retries times
func pull(ociBinary string, image string) error {
	cmd := exec.Command(ociBinary, "pull", image)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error pull image %s : %v", image, err)
	}
	return err
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
		var hostPortBinding string
		if pm.ListenAddress != "" {
			hostPortBinding = net.JoinHostPort(pm.ListenAddress, fmt.Sprintf("%d", pm.HostPort))
		} else {
			hostPortBinding = fmt.Sprintf("%d", pm.HostPort)
		}
		publish := fmt.Sprintf("--publish=%s:%d", hostPortBinding, pm.ContainerPort)
		result = append(result, publish)
	}
	return result
}

// Save saves an image archive "docker/podman save"
func Save(ociBinary string, image, dest string) error {
	cmd := exec.Command(ociBinary, "save", "-o", dest, image)
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
		return errors.Wrapf(err, "saving image to tar failed, output %s", lines[0])
	}
	return nil
}

// CreateOpt is an option for Create
type CreateOpt func(*createOpts) *createOpts

// actual options struct
type createOpts struct {
	RunArgs       []string
	ContainerArgs []string
	Mounts        []Mount
	PortMappings  []PortMapping
}

// CreateContainer creates a container with "docker/podman run"
func CreateContainer(ociBinary string, image string, opts ...CreateOpt) ([]string, error) {
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
func WithRunArgs(args ...string) CreateOpt {
	return func(r *createOpts) *createOpts {
		r.RunArgs = args
		return r
	}
}

// WithMounts sets the container mounts
func WithMounts(mounts []Mount) CreateOpt {
	return func(r *createOpts) *createOpts {
		r.Mounts = mounts
		return r
	}
}

// WithPortMappings sets the container port mappings to the host
func WithPortMappings(portMappings []PortMapping) CreateOpt {
	return func(r *createOpts) *createOpts {
		r.PortMappings = portMappings
		return r
	}
}

// Copy copies a local asset into the container
func Copy(ociBinary string, ociID string, asset assets.CopyableFile) error {
	if _, err := os.Stat(asset.GetAssetName()); os.IsNotExist(err) {
		return errors.Wrapf(err, "error source %s does not exist", asset.GetAssetName())
	}
	destination := fmt.Sprintf("%s:%s", ociID, asset.GetTargetDir())
	cmd := exec.Command(ociBinary, "cp", asset.GetAssetName(), destination)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "error copying %s into node", asset.GetAssetName())
	}
	return nil
}
