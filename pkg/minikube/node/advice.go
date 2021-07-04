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

package node

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// ExitIfFatal before exiting will try to check for different error types and provide advice if we know for sure what the error is
func ExitIfFatal(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, oci.ErrWindowsContainers) {
		exit.Message(reason.Kind{
			ID:       "PROVIDER_DOCKER_CONTAINER_OS",
			ExitCode: reason.ExProviderConflict,
			Style:    style.Conflict,
			URL:      "https://docs.docker.com/docker-for-windows/#switch-between-windows-and-linux-containers",
			Advice:   "From the Docker Desktop menu, select 'Switch to Linux containers'",
		}, "Docker Desktop is configured for Windows containers, but Linux containers are required for minikube")
	}

	if errors.Is(err, oci.ErrCPUCountLimit) {
		if runtime.GOOS == "darwin" {
			exit.Message(reason.RsrcInsufficientDarwinDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		}
		if runtime.GOOS == "windows" {
			exit.Message(reason.RsrcInsufficientWindowsDockerCores, "Docker Desktop has less than 2 CPUs configured, but Kubernetes requires at least 2 to be available")
		}
		exit.Message(reason.RsrcInsufficientCores, "Docker has less than 2 CPUs available, but Kubernetes requires at least 2 to be available")
	}

	if errors.Is(err, kubeadm.ErrNoExecLinux) {
		exit.Message(reason.Kind{
			ID:       "PROVIDER_DOCKER_NOEXEC",
			ExitCode: reason.ExProviderPermission,
			Style:    style.Permissions,
			Issues:   []int{8327},
			Advice:   "Ensure that your Docker mountpoints do not have the 'noexec' flag set",
		}, "The kubeadm binary within the Docker container is not executable")
	}

	if rtErr, ok := err.(*cruntime.ErrServiceVersion); ok {
		exit.Message(reason.Kind{
			ID:       "PROVIDER_INVALID_VERSION",
			ExitCode: reason.ExGuestConfig,
			Style:    style.Unsupported,
			Advice:   "Try to start minikube with '--delete-on-failure=true' option",
		}, fmt.Sprintf("Your existing minikube instance has version %s of service %v which is too old. "+
			"Please try to start minikube with --delete-on-failure=true option", rtErr.Installed, rtErr.Service))
	}
}
