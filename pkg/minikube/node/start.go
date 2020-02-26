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
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

// Start spins up a guest and starts the kubernetes node.
func Start(cc config.ClusterConfig, n config.Node, existingAddons map[string]bool) error {
	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheRequiredImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion)

	// Why do we need this?
	if cc.Downloader == nil {
		cc.Downloader = util.DefaultDownloader{}
	}

	runner, preExists, mAPI, _ := cluster.StartMachine(&cc, &n)
	defer mAPI.Close()

	bs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cc.Name, n.Name)
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}

	k8sVersion := cc.KubernetesConfig.KubernetesVersion
	driverName := cc.Driver
	// exits here in case of --download-only option.
	handleDownloadOnly(&cacheGroup, k8sVersion)
	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(runner, driverName, cc.KubernetesConfig)
	showVersionInfo(k8sVersion, cr)
	waitCacheRequiredImages(&cacheGroup)

	configureMounts()

	// enable addons, both old and new!
	if existingAddons != nil {
		addons.Start(viper.GetString(config.MachineProfile), existingAddons, config.AddonList)
	}

	if err := bs.UpdateNode(cc, n, cr); err != nil {
		exit.WithError("Failed to update node", err)
	}

	if err := CacheAndLoadImagesInConfig(); err != nil {
		out.T(out.FailureType, "Unable to load cached images from config file.")
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForCluster(cc, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}

	err = bs.SetupCerts(cc.KubernetesConfig, n)
	if err != nil {
		exit.WithError("setting up certs", err)
	}

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		exit.WithError("Getting primary control plane", err)
	}
	cpBs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cc.Name, cp.Name)
	if err != nil {
		exit.WithError("Getting bootstrapper", err)
	}
	joinCmd, err := cpBs.GenerateToken(cc.KubernetesConfig)
	if err != nil {
		exit.WithError("generating join token", err)
	}
	return bs.JoinCluster(cc, n, joinCmd)
}
