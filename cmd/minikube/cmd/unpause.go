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
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/pause"
)

// unpauseCmd represents the docker-pause command
var unpauseCmd = &cobra.Command{
	Use:   "unpause",
	Short: "unpause Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		cname := viper.GetString(config.MachineProfile)
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		cc, err := config.Load()
		if err != nil {
			exit.WithError("Error getting config", err)
		}
		glog.Infof("config: %+v", cc)
		host, err := cluster.CheckIfHostExistsAndLoad(api, cname)
		if err != nil {
			exit.WithError("Error getting host", err)
		}

		r, err := machine.CommandRunner(host)
		if err != nil {
			exit.WithError("Failed to get command runner", err)
		}

		config := cruntime.Config{Type: cc.ContainerRuntime, Runner: r}
		cr, err := cruntime.New(config)
		if err != nil {
			exit.WithError("Failed runtime", err)
		}
		err = pause.Unpause(cr, r)
		if err != nil {
			exit.WithError("Pause", err)
		}
		out.T(out.Unpause, "The '{{.name}}' cluster is now unpaused", out.V{"name": cc.Name})
	},
}
