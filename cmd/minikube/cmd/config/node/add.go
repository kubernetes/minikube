/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
)

var name string

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a node to the given cluster.",
	Long:  "Adds a node to the given cluster config, without starting it.",
	Run: func(cmd *cobra.Command, args []string) {

		profile := viper.GetString(config.MachineProfile)
		cc, err := config.Load(profile)
		if err != nil {
			exit.WithError(err)
		}
	},
}

func init() {
	nodeAddCmd.Flags().StringVar(&name, "name", "", "The name of the node to add. Defaults to a variation of the profile name.")
	NodeCmd.AddCommand(nodeAddCmd)
}
