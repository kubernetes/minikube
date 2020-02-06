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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/exit"
)

var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a node.",
	Long:  "Starts an existing stopped node in a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		name := viper.GetString("name")
		if name == "" {
			exit.UsageT("name is required")
		}
	},
}

func init() {
	nodeStartCmd.Flags().String("name", "", "The name of the node to delete")
	nodeCmd.AddCommand(nodeStartCmd)

}
