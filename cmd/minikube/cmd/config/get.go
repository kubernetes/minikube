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

package config

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
)

var configGetCmd = &cobra.Command{
	Use:   "get PROPERTY_NAME",
	Short: "Gets the value of PROPERTY_NAME from the minikube config file",
	Long:  "Returns the value of PROPERTY_NAME from the minikube config file.  Can be overwritten at runtime by flags or environmental variables.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmd.SilenceErrors = true
			return errors.New("usage: minikube config get PROPERTY_NAME")
		}

		cmd.SilenceUsage = true
		val, err := Get(args[0])
		if err != nil {
			return err
		}
		if val == "" {
			return fmt.Errorf("no value for key '%s'", args[0])
		}

		console.OutLn(val)
		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(configGetCmd)
}

// Get gets a property
func Get(name string) (string, error) {
	return pkgConfig.Get(name)
}
