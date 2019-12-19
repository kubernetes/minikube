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

// UpdateCluster generated kubeadm.yaml and other configs and loads the imags
func (k *Bootstrapper) UpdateCluster(cfg config.MachineConfig) error {
	// images, err := images.Kubeadm(cfg.KubernetesConfig.ImageRepository, cfg.KubernetesConfig.KubernetesVersion)
	// if err != nil {
	// 	return errors.Wrap(err, "kubeadm images")
	// }

	// if cfg.KubernetesConfig.ShouldLoadCachedImages {
	// 	if err := machine.LoadImages(&cfg, k.c, images, constants.ImageCacheDir); err != nil {
	// 		out.FailureT("Unable to load cached images: {{.error}}", out.V{"error": err})
	// 	}
	// }
	// r, err := cruntime.New(cruntime.Config{Type: cfg.ContainerRuntime, Socket: cfg.KubernetesConfig.CRISocket})
	// if err != nil {
	// 	return errors.Wrap(err, "runtime")
	// }
	// kubeadmCfg, err := bsutil.GenerateKubeadmYAML(cfg.KubernetesConfig, r)
	// if err != nil {
	// 	return errors.Wrap(err, "generating kubeadm cfg")
	// }

	// kubeletCfg, err := NewKubeletConfig(cfg.KubernetesConfig, r)
	// if err != nil {
	// 	return errors.Wrap(err, "generating kubelet config")
	// }

	// kubeletService, err := NewKubeletService(cfg.KubernetesConfig)
	// if err != nil {
	// 	return errors.Wrap(err, "generating kubelet service")
	// }

	// glog.Infof("kubelet %s config:\n%+v", kubeletCfg, cfg.KubernetesConfig)

	// stopCmd := exec.Command("/bin/bash", "-c", "pgrep kubelet && sudo systemctl stop kubelet")
	// // stop kubelet to avoid "Text File Busy" error
	// if rr, err := k.c.RunCmd(stopCmd); err != nil {
	// 	glog.Warningf("unable to stop kubelet: %s command: %q output: %q", err, rr.Command(), rr.Output())
	// }

	// if err := transferBinaries(cfg.KubernetesConfig, k.c); err != nil {
	// 	return errors.Wrap(err, "downloading binaries")
	// }
	// files := configFiles(cfg.KubernetesConfig, kubeadmCfg, kubeletCfg, kubeletService)
	// if err := addAddons(&files, assets.GenerateTemplateData(cfg.KubernetesConfig)); err != nil {
	// 	return errors.Wrap(err, "adding addons")
	// }
	// for _, f := range files {
	// 	if err := k.c.Copy(f); err != nil {
	// 		return errors.Wrapf(err, "copy")
	// 	}
	// }

	// if _, err := k.c.RunCmd(exec.Command("/bin/bash", "-c", "sudo systemctl daemon-reload && sudo systemctl start kubelet")); err != nil {
	// 	return errors.Wrap(err, "starting kubelet")
	// }
	return nil
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
