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
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

// Start spins up a guest and starts the kubernetes node.
func Start(mc config.ClusterConfig, n config.Node, primary bool, existingAddons map[string]bool) (*kubeconfig.Settings, error) {
	k8sVersion := mc.KubernetesConfig.KubernetesVersion
	driverName := mc.Driver

	// If using kic, make sure we download the kic base image
	var kicGroup errgroup.Group
	if driver.IsKIC(driverName) {
		beginDownloadKicArtifacts(&kicGroup)
	}

	var cacheGroup errgroup.Group
	// Adding a second layer of cache does not make sense for the none driver
	if !driver.BareMetal(driverName) {
		beginCacheKubernetesImages(&cacheGroup, mc.KubernetesConfig.ImageRepository, k8sVersion, mc.KubernetesConfig.ContainerRuntime)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveProfile must be called before startHost, and again afterwards when we know the IP.
	if err := config.SaveProfile(viper.GetString(config.ProfileName), &mc); err != nil {
		exit.WithError("Failed to save config", err)
	}

	// exits here in case of --download-only option.
	handleDownloadOnly(&cacheGroup, &kicGroup, k8sVersion)
	waitDownloadKicArtifacts(&kicGroup)

	mRunner, preExists, machineAPI, host := startMachine(&mc, &n)
	defer machineAPI.Close()

	// wait for preloaded tarball to finish downloading before configuring runtimes
	waitCacheRequiredImages(&cacheGroup)

	sv, err := util.ParseKubernetesVersion(mc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(mRunner, driverName, mc.KubernetesConfig, sv)
	showVersionInfo(k8sVersion, cr)

	//TODO(sharifelgamal): Part out the cluster-wide operations, perhaps using the "primary" param

	// Must be written before bootstrap, otherwise health checks may flake due to stale IP
	kubeconfig, err := setupKubeconfig(host, &mc, &n, mc.Name)
	if err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}

	// setup kubeadm (must come after setupKubeconfig)
	bs := setupKubeAdm(machineAPI, mc, n)

	// pull images or restart cluster
	out.T(out.Launch, "Launching Kubernetes ... ")
	if err := bs.StartCluster(mc); err != nil {
		exit.WithLogEntries("Error starting cluster", err, logs.FindProblems(cr, bs, mRunner))
	}
	configureMounts()

	// enable addons, both old and new!
	if existingAddons != nil {
		addons.Start(viper.GetString(config.ProfileName), existingAddons, AddonList)
	}

	if err = CacheAndLoadImagesInConfig(); err != nil {
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
