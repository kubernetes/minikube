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
	"os"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

// Start spins up a guest and starts the kubernetes node.
func Start(cc config.ClusterConfig, n config.Node, existingAddons map[string]bool) error {
	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheRequiredImages(&cacheGroup, cc.KubernetesConfig.ImageRepository, n.KubernetesVersion)

	runner, preExists, mAPI, _ := cluster.StartMachine(&cc, &n)
	defer mAPI.Close()

	bs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), n.Name)
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

	// special ops for none , like change minikube directory.
	// multinode super doesn't work on the none driver
	if driverName == driver.None && len(cc.Nodes) == 1 {
		prepareNone()
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForCluster(cc, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}

	bs.SetupCerts(cc.KubernetesConfig, n)

	cp, err := config.PrimaryControlPlane(cc)
	if err != nil {
		exit.WithError("Getting primary control plane", err)
	}
	cpBs, err := cluster.Bootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper), cp.Name)
	if err != nil {
		exit.WithError("Getting bootstrapper", err)
	}
	joinCmd, err := cpBs.GenerateToken(cc.KubernetesConfig)
	return bs.JoinCluster(cc, n, joinCmd)
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	out.T(out.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(config.WantNoneDriverWarning) {
		out.T(out.Empty, "")
		out.WarningT("The 'none' driver provides limited isolation and may reduce system security and reliability.")
		out.WarningT("For more information, see:")
		out.T(out.URL, "https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		out.T(out.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may need to relocate them. For example, to overwrite your own settings, run:")

		out.T(out.Empty, "")
		out.T(out.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.T(out.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.T(out.Empty, "")

		out.T(out.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := util.MaybeChownDirRecursiveToMinikubeUser(localpath.MiniPath()); err != nil {
		exit.WithCodeT(exit.Permissions, "Failed to change permissions for {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": localpath.MiniPath(), "error": err})
	}
}
