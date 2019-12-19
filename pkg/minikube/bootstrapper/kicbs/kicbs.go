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

// bootstrapper for kic
package kicbs

import (
	"fmt"
	"net"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

// Bootstrapper is a bootstrapper using kicbs
type Bootstrapper struct {
	c           command.Runner
	contextName string
}

// NewKICBSBootstrapper creates a new kicbs.Bootstrapper
func NewKICBSBootstrapper(api libmachine.API) (*Bootstrapper, error) {
	name := viper.GetString(config.MachineProfile)
	h, err := api.Load(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &Bootstrapper{c: runner, contextName: name}, nil
}

func (k *Bootstrapper) PullImages(config.KubernetesConfig) error {
	return fmt.Errorf("the PullImages is not implemented in kicbs yet")
}
func (k *Bootstrapper) StartCluster(config.KubernetesConfig) error {
	return fmt.Errorf("the StartCluster is not implemented in kicbs yet")
}
func (k *Bootstrapper) UpdateCluster(config.MachineConfig) error {
	return fmt.Errorf("the UpdateCluster is not implemented in kicbs yet")
}
func (k *Bootstrapper) DeleteCluster(config.KubernetesConfig) error {
	return fmt.Errorf("the DeleteCluster is not implemented in kicbs yet")
}
func (k *Bootstrapper) WaitForCluster(config.KubernetesConfig, time.Duration) error {
	return fmt.Errorf("the WaitForCluster is not implemented in kicbs yet")
}
func (k *Bootstrapper) LogCommands(bootstrapper.LogOptions) map[string]string {
	return map[string]string{}
}
func (k *Bootstrapper) SetupCerts(cfg config.KubernetesConfig) error {
	return fmt.Errorf("the SetupCerts is not implemented in kicbs yet")
}
func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	return "", fmt.Errorf("the GetKubeletStatus is not implemented in kicbs yet")
}
func (k *Bootstrapper) GetAPIServerStatus(net.IP, int) (string, error) {
	return "", fmt.Errorf("the GetAPIServerStatus is not implemented in kicbs yet")
}
