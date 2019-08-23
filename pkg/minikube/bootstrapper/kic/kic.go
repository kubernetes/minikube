/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package kic

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

// Bootstrapper is a bootstrapper using kubeadm
type Bootstrapper struct {
	c command.Runner
}

// NewKicBootstrapper creates a new kic.Bootstrapper
func NewKicBootstrapper(api libmachine.API) (*Bootstrapper, error) {
	h, err := api.Load(config.GetMachineName())
	if err != nil {
		return nil, errors.Wrap(err, "getting api client")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "command runner")
	}
	return &Bootstrapper{c: runner}, nil
}

// GetKubeletStatus returns the kubelet status
func (k *Bootstrapper) GetKubeletStatus() (string, error) {
	statusCmd := `sudo systemctl is-active kubelet`
	status, err := k.c.CombinedOutput(statusCmd)
	if err != nil {
		return "", errors.Wrap(err, "getting status")
	}
	s := strings.TrimSpace(status)
	switch s {
	case "active":
		return state.Running.String(), nil
	case "inactive":
		return state.Stopped.String(), nil
	case "activating":
		return state.Starting.String(), nil
	}
	return state.Error.String(), nil
}

// GetAPIServerStatus returns the api-server status
func (k *Bootstrapper) GetAPIServerStatus(ip net.IP, apiserverPort int) (string, error) {
	url := fmt.Sprintf("https://%s:%d/healthz", ip, apiserverPort)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // To avoid connectiv issue if http(s)_proxy is set.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	glog.Infof("%s response: %v %+v", url, err, resp)
	// Connection refused, usually.
	if err != nil {
		return state.Stopped.String(), nil
	}
	if resp.StatusCode != http.StatusOK {
		return state.Error.String(), nil
	}
	return state.Running.String(), nil
}

// LogCommands returns a map of log type to a command which will display that log.
func (k *Bootstrapper) LogCommands(o bootstrapper.LogOptions) map[string]string {
	var kubelet strings.Builder
	kubelet.WriteString("journalctl -u kubelet")
	if o.Lines > 0 {
		kubelet.WriteString(fmt.Sprintf(" -n %d", o.Lines))
	}
	if o.Follow {
		kubelet.WriteString(" -f")
	}

	var dmesg strings.Builder
	dmesg.WriteString("sudo dmesg -PH -L=never --level warn,err,crit,alert,emerg")
	if o.Follow {
		dmesg.WriteString(" --follow")
	}
	if o.Lines > 0 {
		dmesg.WriteString(fmt.Sprintf(" | tail -n %d", o.Lines))
	}
	return map[string]string{
		"kubelet": kubelet.String(),
		"dmesg":   dmesg.String(),
	}
}

// StartCluster starts the cluster
func (k *Bootstrapper) StartCluster(k8s config.KubernetesConfig) error {
	return nil
}

// WaitCluster blocks until Kubernetes appears to be healthy.
func (k *Bootstrapper) WaitCluster(k8s config.KubernetesConfig) error {
	// TODOO: Later
	time.Sleep(10 * time.Second)
	return nil
}

// RestartCluster restarts the Kubernetes cluster configured by kubeadm
func (k *Bootstrapper) RestartCluster(k8s config.KubernetesConfig) error {
	// docker restart
	return nil
}

// DeleteCluster removes the components that were started earlier
func (k *Bootstrapper) DeleteCluster(k8s config.KubernetesConfig) error {
	cmd := fmt.Sprintf("sudo kubeadm reset --force")
	out, err := k.c.CombinedOutput(cmd)
	if err != nil {
		return errors.Wrapf(err, "kubeadm reset: %s\n%s\n", cmd, out)
	}

	return nil
}

// PullImages downloads images that will be used by RestartCluster
func (k *Bootstrapper) PullImages(k8s config.KubernetesConfig) error {
	// pull the image
	return nil
}

// SetupCerts sets up certificates within the cluster.
func (k *Bootstrapper) SetupCerts(k8s config.KubernetesConfig) error {
	return bootstrapper.SetupCerts(k.c, k8s)
}

// UpdateCluster updates generates kubeadm/kubelet,... configs
// also transgers config and binaries to the bootstrapper's control-plane node
func (k *Bootstrapper) UpdateCluster(cfg config.KubernetesConfig) error {
	// TODO:medyagh investigate if loading cached images is needed for kic

	// r, err := cruntime.New(cruntime.Config{Type: cfg.ContainerRuntime, Socket: cfg.CRISocket})
	// if err != nil {
	// 	return errors.Wrap(err, "update cluster new runtim")
	// }

	// MEDYA:Todo genrate kubeadmCfg []byte
	// TODO:medyagh investigate if we could genrate kubeletCfg for kic

	// stop kubelet to avoid "Text File Busy" error'
	// TODO:medyagh investigate if needed in kic
	// err = k.c.Run(`pgrep kubelet && sudo systemctl stop kubelet`)
	// if err != nil {
	// 	glog.Warningf("unable to stop kubelet: %s", err)
	// }

	// kubeadmCfg := []byte{}

	// files := []assets.CopyableFile{
	// 	assets.NewMemoryAssetTarget(kubeadmCfg, yamlConfigPath, "0640"),
	// }

	// if err := addAddons(&files, assets.GenerateTemplateData(cfg)); err != nil {
	// 	return errors.Wrap(err, "adding addons")
	// }
	// for _, f := range files {
	// 	if err := k.c.Copy(f); err != nil {
	// 		return errors.Wrapf(err, "copy")
	// 	}
	// }

	// if err := k.c.Run(`sudo systemctl daemon-reload && sudo systemctl start kubelet`); err != nil {
	// 	return errors.Wrap(err, "starting kubelet")
	// }
	return nil
}
