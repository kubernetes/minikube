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
	"k8s.io/klog"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

type containerdInstaller struct {
	Options              *engine.Options
	ContainerRuntimeName string
	Commander            runner.Runner
	Provider             string
	AuthOptions          *auth.Options
}

func NewContainerdInstaller(opts *engine.Options, cmd runner.Runner, provider string, authOpts *auth.Options) *containerdInstaller {
	return &containerdInstaller{
		Options:              opts,
		ContainerRuntimeName: "CRI-O",
		Commander:            cmd,
		Provider:             provider,
		AuthOptions:          authOpts,
	}
}

// x7NOTE: complete this
func (ci *containerdInstaller) InstallCRuntime() error {
	if err := ci.SetCRuntimeOptions(); err != nil {
		klog.Infof("Error setting container-runtime (%s) options during provisioning %v",
			ci.ContainerRuntimeName, err)
	}
	return nil
}

// x7NOTE: complete this
func (ci *containerdInstaller) SetCRuntimeOptions() error { return nil }
