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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

var configSetCmd = &cobra.Command{
	Use:   "set PROPERTY_NAME PROPERTY_VALUE",
	Short: "Sets an individual value in a minikube config file",
	Long: `Sets the PROPERTY_NAME config value to PROPERTY_VALUE
	These values can be overwritten by flags or environment variables at runtime.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			exit.UsageT("not enough arguments ({{.ArgCount}}).\nusage: minikube config set PROPERTY_NAME PROPERTY_VALUE", out.V{"ArgCount": len(args)})
		}
		if len(args) > 2 {
			exit.UsageT("toom any arguments ({{.ArgCount}}).\nusage: minikube config set PROPERTY_NAME PROPERTY_VALUE", out.V{"ArgCount": len(args)})
		}
		err := Set(args[0], args[1])
		if err != nil {
			exit.WithError(exit.ProgramError, "Set failed", err)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(configSetCmd)
}

// Set sets a property to a value
func Set(name string, value string) error {
	s, err := findSetting(name)
	if err != nil {
		return errors.Wrapf(err, "find settings for %q value of %q", name, value)
	}
	// Validate the new value
	err = run(name, value, s.validations)
	if err != nil {
		return errors.Wrapf(err, "run validations for %q with value of %q", name, value)
	}

	// Set the value
	cc, err := config.ReadConfig(localpath.ConfigFile())
	if err != nil {
		return errors.Wrapf(err, "read config file %q", localpath.ConfigFile())
	}
	err = s.set(cc, name, value)
	if err != nil {
		return errors.Wrapf(err, "set")
	}

	// Run any callbacks for this property
	err = run(name, value, s.callbacks)
	if err != nil {
		return errors.Wrapf(err, "run callbacks for %q with value of %q", name, value)
	}

	// Write the value
	return config.WriteConfig(localpath.ConfigFile(), cc)
}
