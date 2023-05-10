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

package cruntimeinstaller

import (
	"strings"

	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

// Intaller is the abstraction that we use in order to install and configure
// a generic Container Runtime inside the linux machine.
// Configuring authentication is still provision.Provisioner's responsibility.
type installer interface {
	// InstallCRuntime downloads and installs the container runtime into the machine
	InstallCRuntime() error

	// SetCRuntimeOptions gives the configuration to the container runtime
	SetCRuntimeOptions() error
}

// InstalLCRuntime is the entrypoint to this package.
// we're detecting the installer from the engine.Options.EngineName field, and passing
// all the requirements to the installer.
// we're then calling the install method, which installs and configures
// the container runtime into the machine.
func InstallCRuntime(ngineConf *engine.Options, commander runner.Runner, provider string, authOpts *auth.Options) error {
	instllr := detectInstaller(ngineConf, commander, provider, authOpts)
	return instllr.InstallCRuntime()
}

// detectInstaller takes all the requirements as parameters and returns
// an initialized struct for the correct container runtime installer
func detectInstaller(ngineConf *engine.Options, commander runner.Runner, provider string, authOpts *auth.Options) installer {
	switch strings.ToLower(ngineConf.EngineName) {
	case "crio", "cri-o":
		return newCRIOInstaller(ngineConf, commander, provider, authOpts)
	case "containerd":
		return newContainerdInstaller(ngineConf, commander, provider, authOpts)
	default:
		return newDockerInstaller(ngineConf, commander, provider, authOpts)
	}
}
