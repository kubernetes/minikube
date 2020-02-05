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
	"time"

	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cluster"
	pkg_config "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
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

	nonexistent := false
	stop := func() (err error) {
		err = cluster.StopHost(api)
		if err == nil {
			return nil
		}
		glog.Warningf("stop host returned error: %v", err)

		switch err := errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			out.T(out.Meh, `"{{.profile_name}}" VM does not exist, nothing to stop`, out.V{"profile_name": profile})
			nonexistent = true
			return nil
		default:
			return err
		}
	}

	if err := retry.Expo(stop, 5*time.Second, 3*time.Minute, 5); err != nil {
		exit.WithError("Unable to stop VM", err)
	}

	if !nonexistent {
		out.T(out.Stopped, `"{{.profile_name}}" stopped.`, out.V{"profile_name": profile})
	}

	if err := killMountProcess(); err != nil {
		out.T(out.WarningType, "Unable to kill mount process: {{.error}}", out.V{"error": err})
	}

	err = kubeconfig.UnsetCurrentContext(profile, constants.KubeconfigPath)
	if err != nil {
		exit.WithError("update config", err)
	}
}
