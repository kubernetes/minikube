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

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/util/retry"
)

var stopAll bool

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops a running local Kubernetes cluster",
	Long: `Stops a local Kubernetes cluster running in Virtualbox. This command stops the VM
itself, leaving all files intact. The cluster can be started again with the "start" command.`,
	Run: runStop,
}

func init() {

	stopCmd.Flags().BoolVar(&stopAll, "all", false, "Set flag to stop all profiles (clusters)")

	if err := viper.GetViper().BindPFlags(stopCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}

	RootCmd.AddCommand(stopCmd)
}

// runStop handles the executes the flow of "minikube stop"
func runStop(cmd *cobra.Command, args []string) {
	register.SetEventLogPath(localpath.EventLog(ClusterFlagValue()))
	register.Reg.SetStep(register.Stopping)

	// new code
	var profilesToStop []string
	if stopAll {
		validProfiles, _, err := config.ListProfiles()
		if err != nil {
			glog.Warningf("'error loading profiles in minikube home %q: %v", localpath.MiniPath(), err)
		}
		for _, profile := range validProfiles {
			profilesToStop = append(profilesToStop, profile.Name)
		}
	} else {
		cname := ClusterFlagValue()
		profilesToStop = append(profilesToStop, cname)
	}

	stoppedNodes := 0

	for _, profile := range profilesToStop {
		register.Reg.SetStep(register.Stopping)

		// end new code
		api, cc := mustload.Partial(profile)
		defer api.Close()

		for _, n := range cc.Nodes {
			machineName := driver.MachineName(*cc, n)

			nonexistent := stop(api, machineName)
			if !nonexistent {
				stoppedNodes++
			}
		}

		if err := killMountProcess(); err != nil {
			out.WarningT("Unable to kill mount process: {{.error}}", out.V{"error": err})
		}

		if err := kubeconfig.UnsetCurrentContext(profile, kubeconfig.PathFromEnv()); err != nil {
			exit.WithError("update config", err)
		}
	}

	register.Reg.SetStep(register.Done)
	if stoppedNodes > 0 {
		out.T(out.Stopped, `{{.count}} nodes stopped.`, out.V{"count": stoppedNodes})
	}
}

func stop(api libmachine.API, machineName string) bool {
	nonexistent := false

	tryStop := func() (err error) {
		err = machine.StopHost(api, machineName)
		if err == nil {
			return nil
		}
		glog.Warningf("stop host returned error: %v", err)

		switch err := errors.Cause(err).(type) {
		case mcnerror.ErrHostDoesNotExist:
			out.T(out.Meh, `"{{.machineName}}" does not exist, nothing to stop`, out.V{"machineName": machineName})
			nonexistent = true
			return nil
		default:
			return err
		}
	}

	if err := retry.Expo(tryStop, 1*time.Second, 120*time.Second, 5); err != nil {
		exit.WithError("Unable to stop VM", err)
	}

	return nonexistent
}
