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
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
)

// MaybeExitWithAdvice before exiting will try to check for different error types and provide advice if we know for sure what the error is
func MaybeExitWithAdvice(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, oci.ErrWindowsContainers) {
		out.ErrLn("")
		out.ErrT(out.Conflict, "Your Docker Desktop container OS type is Windows but Linux is required.")
		out.T(out.Warning, "Please change Docker settings to use Linux containers instead of Windows containers.")
		out.T(out.Documentation, "https://minikube.sigs.k8s.io/docs/drivers/docker/#verify-docker-container-type-is-linux")
		exit.UsageT(`You can verify your Docker container type by running:
{{.command}}
	`, out.V{"command": "docker info --format '{{.OSType}}'"})
	}

	if errors.Is(err, oci.ErrCPUCountLimit) {
		out.ErrLn("")
		out.ErrT(out.Conflict, "{{.name}} doesn't have enough CPUs. ", out.V{"name": driver.FullName(viper.GetString("driver"))})
		if runtime.GOOS != "linux" && viper.GetString("driver") == "docker" {
			out.T(out.Warning, "Please consider changing your Docker Desktop's resources.")
			out.T(out.Documentation, "https://docs.docker.com/config/containers/resource_constraints/")
		} else {
			cpuCount := viper.GetInt(cpus)
			if cpuCount == 2 {
				out.T(out.Tip, "Please ensure your system has {{.cpu_counts}} CPU cores.", out.V{"cpu_counts": viper.GetInt(cpus)})
			} else {
				out.T(out.Tip, "Please ensure your {{.driver_name}} system has access to {{.cpu_counts}} CPU cores or reduce the number of the specified CPUs", out.V{"driver_name": driver.FullName(viper.GetString("driver")), "cpu_counts": viper.GetInt(cpus)})
			}
		}

		// TODO: Use a problem
		exit.WithCodeT(exit.InsufficientCores, "Ensure your {{.driver_name}} system has enough CPUs. The minimum allowed is 2 CPUs.", out.V{"driver_name": viper.GetString("driver")})
	}

	if errors.Is(err, kubeadm.ErrNoExecLinux) {
		out.ErrLn("")
		out.ErrT(out.Conflict, "kubeadm binary is not executable !")
		out.T(out.Documentation, "Try the solution in this link: https://github.com/kubernetes/minikube/issues/8327#issuecomment-651288459")
		exit.UsageT(`Ensure the binaries are not mounted with "noexec" option. To check run:

	$ findmnt

`)
	}

}
