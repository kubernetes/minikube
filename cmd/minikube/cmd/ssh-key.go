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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

// sshKeyCmd represents the sshKey command
var sshKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "Retrieve the ssh identity key path of the specified cluster",
	Long:  "Retrieve the ssh identity key path of the specified cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := config.Load(viper.GetString(config.MachineProfile))
		if err != nil {
			exit.WithError("Getting machine config failed", err)
		}
		out.Ln(filepath.Join(localpath.MiniPath(), "machines", cc.Name, "id_rsa"))
	},
}
