/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package cruntimeInstaller

import (
	"strings"

	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

// Intaller is the abstraction that we use in order to install and configure
// a generic Container Runtime inside the linux machine.
// Configuring authentication is still provision.Provisioner's responsibility.
type Installer interface {
	// InstallCRuntime downloads and installs the container runtime into the machine
	InstallCRuntime() error

	// SetCRuntimeOptions gives the configuration to the container runtime
	SetCRuntimeOptions() error
}

// DetectCRuntimeInstaller should be the entrypoint to this package.
// By passing the ngineConf we should be able to tell which container runtime we want to install.
// Also we're passing all the requirements for the installation/configuration.
func DetectCRuntimeInstaller(ngineConf *engine.Options, commander runner.Runner, provider string, authOpts *auth.Options) Installer {
	switch strings.ToLower(ngineConf.EngineName) {
	case "crio", "cri-o":
		return NewCRIOInstaller(ngineConf, commander, provider, authOpts)
	case "containerd":
		return NewContainerdInstaller(ngineConf, commander, provider, authOpts)
	default:
		return NewDockerInstaller(ngineConf, commander, provider, authOpts)
	}
}
