/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"errors"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

// FailFastError type is an error that could not be solved by trying again
type FailFastError struct {
	Err error
}

func (f *FailFastError) Error() string {
	return f.Err.Error()
}

// ErrWindowsContainers is thrown when docker been configured to run windows containers instead of Linux
var ErrWindowsContainers = &FailFastError{errors.New("docker container type is windows")}

// ErrMinDockerVersion is thrown when docker version is less than minimum requited by Minikube
var ErrMinDockerVersion = &FailFastError{errors.New("docker version is less than the minimum required")}

// ErrCPUCountLimit is thrown when docker daemon doesn't have enough CPUs for the requested container
var ErrCPUCountLimit = &FailFastError{errors.New("not enough CPUs is available for container")}

// ErrIPinUse is thrown when the container been given an IP used by another container
var ErrIPinUse = &FailFastError{errors.New("can't create with that IP, address already in use")}

// ErrExitedUnexpectedly is thrown when container is created/started without error but later it exists and it's status is not running anymore.
var ErrExitedUnexpectedly = errors.New("container exited unexpectedly")

// ErrDaemonInfo is thrown when docker/podman info is failing or not responding
var ErrDaemonInfo = errors.New("daemon info not responding")

// ErrInsufficientDockerStorage is thrown when there is not more storage for docker
var ErrInsufficientDockerStorage = &FailFastError{errors.New("insufficient docker storage, no space left on device")}

// ErrVolumeNotFound is when given volume was not found
var ErrVolumeNotFound = errors.New("kic volume not found")

// ErrNetworkSubnetTaken is thrown when a subnet is taken by another network
var ErrNetworkSubnetTaken = errors.New("subnet is taken")

// ErrNetworkNotFound is when given network was not found
var ErrNetworkNotFound = errors.New("kic network not found")

// ErrNetworkGatewayTaken is when given network gatway is taken
var ErrNetworkGatewayTaken = errors.New("network gateway is taken")

// ErrNetworkInUse is when trying to delete a network which is attached to another container
var ErrNetworkInUse = errors.New("unable to delete a network that is attached to a running container")

// ErrGetSSHPortContainerNotRunning happens when you try to inspect a container (in order to get SSH port) that "exists" but is no longer running
var ErrGetSSHPortContainerNotRunning = errors.New("unable to inspect a not running container to get SSH port")

// ErrGetPortContainerNotRunning happens when you try to inspect a container (in order to get Port) that "exists" but is no longer running
var ErrGetPortContainerNotRunning = errors.New("unable to inspect a not running container to get port")

// LogContainerDebug will print relevant docker/podman infos after a container fails
func LogContainerDebug(ociBin string, name string) string {
	rr, err := containerInspect(ociBin, name)
	if err != nil {
		klog.Warningf("Failed to get postmortem inspect. %s :%v", rr.Command(), err)
	} else {
		klog.Infof("Postmortem inspect (%q): %s", rr.Command(), rr.Output())
	}

	rr, err = containerLogs(ociBin, name)
	if err != nil {
		klog.Warningf("Failed to get postmortem logs. %s :%v", rr.Command(), err)
	} else {
		klog.Infof("Postmortem logs (%q): %s", rr.Command(), rr.Output())
	}
	if ociBin == Docker {
		di, err := dockerSystemInfo()
		if err != nil {
			klog.Warningf("Failed to get postmortem docker info: %v", err)
		} else {
			klog.Infof("postmortem docker info: %+v", di)
		}
		logDockerNetworkInspect(ociBin, name)
	} else {
		pi, err := podmanSystemInfo()
		if err != nil {
			klog.Warningf("couldn't get postmortem podman info: %v", err)
		} else {
			klog.Infof("postmortem podman info: %+v", pi)
		}
		logDockerNetworkInspect(ociBin, name)
	}

	if rr.Stdout.Len() == 0 {
		return ""
	}

	// If available, return an excerpt of the post-mortem logs for inclusion in error message
	excerpt := strings.Split(strings.TrimSpace(rr.Stdout.String()), "\n")
	if len(excerpt) > 4 {
		excerpt = excerpt[len(excerpt)-4:]
	}
	return strings.Join(excerpt, "\n")
}

// containerLogs will return out the logs for a container
func containerLogs(ociBin string, name string) (*RunResult, error) {
	if ociBin == Docker {
		return runCmd(exec.Command(ociBin, "logs", "--timestamps", "--details", name))
	}
	// podman doesn't have --details
	return runCmd(exec.Command(ociBin, "logs", "--timestamps", name))
}

// containerInspect will return the inspect for a container
func containerInspect(ociBin string, name string) (*RunResult, error) {
	return runCmd(exec.Command(ociBin, "inspect", name))
}
