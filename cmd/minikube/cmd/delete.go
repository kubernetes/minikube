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
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a local kubernetes cluster",
	Long: `Deletes a local kubernetes cluster. This command deletes the VM, and removes all
associated files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			exit.Usage("usage: minikube delete")
		}
		profile := viper.GetString(pkg_config.MachineProfile)
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()

		cc, err := pkg_config.Load()
		if err != nil && !os.IsNotExist(err) {
			console.ErrLn("Error loading profile config: %v", err)
		}

		// In the case of "none", we want to uninstall Kubernetes as there is no VM to delete
		if err == nil && cc.MachineConfig.VMDriver == "none" {
			kc := cc.KubernetesConfig
			bsName := viper.GetString(cmdcfg.Bootstrapper)
			console.OutStyle("resetting", "Uninstalling Kubernetes %s using %s ...", kc.KubernetesVersion, bsName)
			clusterBootstrapper, err := GetClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
			if err != nil {
				console.ErrLn("Unable to get bootstrapper: %v", err)
			} else {
				if err = clusterBootstrapper.DeleteCluster(kc); err != nil {
					console.ErrLn("Failed to delete cluster: %v", err)
				}
			}
		}

		if err = cluster.DeleteHost(api); err != nil {
			switch err := errors.Cause(err).(type) {
			case mcnerror.ErrHostDoesNotExist:
				console.OutStyle("meh", "%q cluster does not exist", profile)
			default:
				exit.WithError("Failed to delete cluster", err)
			}
		}

		if err := cmdUtil.KillMountProcess(); err != nil {
			console.Fatal("Failed to kill mount process: %v", err)
		}

		if err := os.Remove(constants.GetProfileFile(viper.GetString(pkg_config.MachineProfile))); err != nil {
			if os.IsNotExist(err) {
				console.OutStyle("meh", "%q profile does not exist", profile)
				os.Exit(0)
			}
			exit.WithError("Failed to remove profile", err)
		}
		console.OutStyle("crushed", "The %q cluster has been deleted.", profile)
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
