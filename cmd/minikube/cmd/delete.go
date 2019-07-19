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

package cmd

import (
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	pkgutil "k8s.io/minikube/pkg/util"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a local kubernetes cluster",
	Long: `Deletes a local kubernetes cluster. This command deletes the VM, and removes all
associated files.`,
	Run: runDelete,
}

// runDelete handles the executes the flow of "minikube delete"
func runDelete(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		exit.UsageT("usage: minikube delete")
	}
	profile := viper.GetString(pkg_config.MachineProfile)
	api, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Error getting client", err)
	}
	defer api.Close()

	cc, err := pkg_config.Load()
	if err != nil && !os.IsNotExist(err) {
		console.ErrT(console.Sad, "Error loading profile config: {{.error}}", console.Arg{"name": profile})
	}

	// In the case of "none", we want to uninstall Kubernetes as there is no VM to delete
	if err == nil && cc.MachineConfig.VMDriver == constants.DriverNone {
		uninstallKubernetes(api, cc.KubernetesConfig, viper.GetString(cmdcfg.Bootstrapper))
	}

	if err = cluster.DeleteHost(api); err != nil {
		switch err := errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			console.OutT(console.Meh, `"{{.name}}" cluster does not exist`, console.Arg{"name": profile})
		default:
			exit.WithError("Failed to delete cluster", err)
		}
	}

	if err := cmdUtil.KillMountProcess(); err != nil {
		console.FatalT("Failed to kill mount process: {{.error}}", console.Arg{"error": err})
	}

	if err := os.RemoveAll(constants.GetProfilePath(viper.GetString(pkg_config.MachineProfile))); err != nil {
		if os.IsNotExist(err) {
			console.OutT(console.Meh, `"{{.profile_name}}" profile does not exist`, console.Arg{"profile_name": profile})
			os.Exit(0)
		}
		exit.WithError("Failed to remove profile", err)
	}
	console.OutT(console.Crushed, `The "{{.cluster_name}}" cluster has been deleted.`, console.Arg{"cluster_name": profile})

	machineName := pkg_config.GetMachineName()
	if err := pkgutil.DeleteKubeConfigContext(constants.KubeconfigPath, machineName); err != nil {
		exit.WithError("update config", err)
	}
}

func uninstallKubernetes(api libmachine.API, kc pkg_config.KubernetesConfig, bsName string) {
	console.OutT(console.Resetting, "Uninstalling Kubernetes {{.kubernetes_version}} using {{.bootstrapper_name}} ...", console.Arg{"kubernetes_version": kc.KubernetesVersion, "bootstrapper_name": bsName})
	clusterBootstrapper, err := getClusterBootstrapper(api, bsName)
	if err != nil {
		console.ErrT(console.Empty, "Unable to get bootstrapper: {{.error}}", console.Arg{"error": err})
	} else if err = clusterBootstrapper.DeleteCluster(kc); err != nil {
		console.ErrT(console.Empty, "Failed to delete cluster: {{.error}}", console.Arg{"error": err})
	}
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
