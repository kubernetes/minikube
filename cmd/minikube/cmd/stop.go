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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a running local kubernetes cluster",
	Long: `Stops a local kubernetes cluster running in Virtualbox. This command stops the VM
itself, leaving all files intact. The cluster can be started again with the "start" command.`,
	Run: runStop,
}

// runStop handles the executes the flow of "minikube stop"
func runStop(cmd *cobra.Command, args []string) {
	profile := viper.GetString(pkg_config.MachineProfile)
	api, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Error getting client", err)
	}
	defer api.Close()

	cc, err := config.Load(profile)
	if err != nil {
		exit.WithError("Error retrieving config", err)
	}

	// TODO replace this back with expo backoff
	for _, n := range cc.Nodes {
		err := machine.StopHost(api, driver.MachineName(profile, n.Name))
		if err != nil {
			exit.WithError("Unable to stop VM", err)
		}
		/*if err := retry.Expo(fn, 5*time.Second, 3*time.Minute, 5); err != nil {
			exit.WithError("Unable to stop VM", err)
		}*/
	}

	out.T(out.Stopped, `"{{.profile_name}}" stopped.`, out.V{"profile_name": profile})

	if err := killMountProcess(); err != nil {
		out.T(out.WarningType, "Unable to kill mount process: {{.error}}", out.V{"error": err})
	}

	err = kubeconfig.UnsetCurrentContext(profile, kubeconfig.PathFromEnv())
	if err != nil {
		exit.WithError("update config", err)
	}
}
