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
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	DockerEnv []string
	DockerOpt []string
)

// configureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(runner cruntime.CommandRunner, drvName string, k8s config.KubernetesConfig) cruntime.Manager {
	config := cruntime.Config{Type: viper.GetString(containerRuntime), Runner: runner, ImageRepository: k8s.ImageRepository, KubernetesVersion: k8s.KubernetesVersion}
	cr, err := cruntime.New(config)
	if err != nil {
		exit.WithError("Failed runtime", err)
	}

	disableOthers := true
	if driver.BareMetal(drvName) {
		disableOthers = false
	}
	err = cr.Enable(disableOthers)
	if err != nil {
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

func showVersionInfo(k8sVersion string, cr cruntime.Manager) {
	version, _ := cr.Version()
	out.T(cr.Style(), "Preparing Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}} ...", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
	for _, v := range DockerOpt {
		out.T(out.Option, "opt {{.docker_option}}", out.V{"docker_option": v})
	}
	for _, v := range DockerEnv {
		out.T(out.Option, "env {{.docker_env}}", out.V{"docker_env": v})
	}
}
