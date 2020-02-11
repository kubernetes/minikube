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
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/out"
)

// Start spins up a guest and starts the kubernetes node.
func Start(mc *config.MachineConfig, n *config.Node, primary bool) (*kubeconfig.Settings, error) {
	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheRequiredImages(&cacheGroup, mc.KubernetesConfig.ImageRepository, n.KubernetesVersion)

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.MachineProfile), mc); err != nil {
		exit.WithError("Failed to save config", err)
	}

	k8sVersion := mc.KubernetesConfig.KubernetesVersion
	// exits here in case of --download-only option.
	handleDownloadOnly(&cacheGroup, k8sVersion)
	mRunner, preExists, machineAPI, host := startMachine(mc, n)
	defer machineAPI.Close()
	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(mRunner, mc.Driver, mc.KubernetesConfig)
	showVersionInfo(k8sVersion, cr)
	waitCacheRequiredImages(&cacheGroup)

	// Must be written before bootstrap, otherwise health checks may flake due to stale IP
	kubeconfig, err := setupKubeconfig(host, &mc, &n, mc.Name)
	if err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}

	// setup kubeadm (must come after setupKubeconfig)
	bs := setupKubeAdm(machineAPI, mc, n)

	// pull images or restart cluster
	bootstrapCluster(bs, cr, mRunner, mc, preExists, isUpgrade)
	configureMounts()

	// enable addons, both old and new!
	existingAddons := map[string]bool{}
	if existing != nil && existing.Addons != nil {
		existingAddons = existing.Addons
	}
	addons.Start(viper.GetString(config.MachineProfile), existingAddons, addonList)

	if err = cacheAndLoadImagesInConfig(); err != nil {
		out.T(out.FailureType, "Unable to load cached images from config file.")
	}

	// special ops for none , like change minikube directory.
	if driverName == driver.None {
		prepareNone()
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForCluster(mc, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}

	return kubeconfig, nil
}
