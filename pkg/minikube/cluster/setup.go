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

package cluster

import (
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
	waitTimeout      = "wait-timeout"
	waitUntilHealthy = "wait"
	embedCerts       = "embed-certs"
	keepContext      = "keep-context"
)

// InitialSetup performs all necessary operations on the initial control plane node when first spinning up a cluster
func InitialSetup(cc config.ClusterConfig, n config.Node, cr cruntime.Manager) (*kubeconfig.Settings, error) {
	mRunner, preExists, machineAPI, host := StartMachine(&cc, &n)
	defer machineAPI.Close()

	// Must be written before bootstrap, otherwise health checks may flake due to stale IP
	kubeconfig, err := setupKubeconfig(host, &cc, &n, cc.Name)
	if err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}

	// setup kubeadm (must come after setupKubeconfig)
	bs := setupKubeAdm(machineAPI, cc, n)

	// pull images or restart cluster
	out.T(out.Launch, "Launching Kubernetes ... ")
	if err := bs.StartCluster(cc); err != nil {
		exit.WithLogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, mRunner))
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForCluster(cc, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}

	return kubeconfig, nil

}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, cfg config.ClusterConfig, n config.Node) bootstrapper.Bootstrapper {
	bs, err := Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), n.Name)
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range ExtraOptions {
		out.T(out.Option, "{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	if err := bs.UpdateCluster(cfg); err != nil {
		exit.WithError("Failed to update cluster", err)
	}
	if err := bs.SetupCerts(cfg.KubernetesConfig, n); err != nil {
		exit.WithError("Failed to setup certs", err)
	}
	return bs
}

func setupKubeconfig(h *host.Host, c *config.ClusterConfig, n *config.Node, clusterName string) (*kubeconfig.Settings, error) {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	if !driver.IsKIC(h.DriverName) {
		addr = strings.Replace(addr, "tcp://", "https://", -1)
		addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(n.Port), -1)
	}

	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, n.IP, c.KubernetesConfig.APIServerName, -1)
	}
	kcs := &kubeconfig.Settings{
		ClusterName:          clusterName,
		ClusterServerAddress: addr,
		ClientCertificate:    localpath.MakeMiniPath("client.crt"),
		ClientKey:            localpath.MakeMiniPath("client.key"),
		CertificateAuthority: localpath.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
		EmbedCerts:           viper.GetBool(embedCerts),
	}

	kcs.SetPath(kubeconfig.PathFromEnv())
	if err := kubeconfig.Update(kcs); err != nil {
		return kcs, err
	}
	return kcs, nil
}
