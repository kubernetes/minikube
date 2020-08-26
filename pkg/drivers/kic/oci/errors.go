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

	"github.com/golang/glog"
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

// ErrCPUCountLimit is thrown when docker daemon doesn't have enough CPUs for the requested container
var ErrCPUCountLimit = &FailFastError{errors.New("not enough CPUs is available for container")}

// ErrExitedUnexpectedly is thrown when container is created/started without error but later it exists and it's status is not running anymore.
var ErrExitedUnexpectedly = errors.New("container exited unexpectedly")

// ErrDaemonInfo is thrown when docker/podman info is failing or not responding
var ErrDaemonInfo = errors.New("daemon info not responding")

// ErrNetworkSubnetTaken is thrown when a subnet is taken by another network
var ErrNetworkSubnetTaken = errors.New("subnet is taken")

// LogContainerDebug will print relevant docker/podman infos after a container fails
func LogContainerDebug(ociBin string, name string) string {
	rr, err := containerInspect(ociBin, name)
	if err != nil {
		glog.Warningf("Filed to get postmortem inspect. %s :%v", rr.Command(), err)
	} else {
		glog.Infof("Postmortem inspect (%q): %s", rr.Command(), rr.Output())
	}

	rr, err = containerLogs(ociBin, name)
	if err != nil {
		glog.Warningf("Filed to get postmortem logs. %s :%v", rr.Command(), err)
	} else {
		glog.Infof("Postmortem logs (%q): %s", rr.Command(), rr.Output())
	}
	if ociBin == Docker {
		di, err := dockerSystemInfo()
		if err != nil {
			glog.Warningf("Failed to get postmortem docker info: %v", err)
		} else {
			glog.Infof("postmortem docker info: %+v", di)
		}
	} else {
		pi, err := podmanSystemInfo()
		if err != nil {
			glog.Warningf("couldn't get postmortem info, failed to to run podman info: %v", err)
		} else {
			glog.Infof("postmortem podman info: %+v", pi)
		}
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
