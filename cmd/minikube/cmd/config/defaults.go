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
	"strings"

	"github.com/spf13/cobra"
)

var configDefaultsCommand = &cobra.Command{
	Use:   "defaults PROPERTY_NAME",
	Short: "Lists all valid default values for PROPERTY_NAME",
	Long: `list displays all valid default settings for PROPERTY_NAME
Acceptable fields: ` + "\n\n" + fieldsWithDefaults(),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.SilenceErrors = true
			return errors.New("not enough arguments.\nusage: minikube config list PROPERTY_NAME")
		}
		if len(args) > 1 {
			cmd.SilenceErrors = true
			return fmt.Errorf("too many arguments (%d)\nusage: minikube config list PROPERTY_NAME", len(args))
		}

		property := args[0]
		return listDefaults(property)
	},
}

func listDefaults(property string) error {
	setting, err := findSetting(property)
	if err != nil {
		return err
	}
	if setting.validDefaults == nil {
		return fmt.Errorf("%s is not a valid option for the `defaults` command; to see valid options run `minikube config defaults -h`", property)
	}
	defaults := setting.validDefaults()
	for _, d := range defaults {
		fmt.Printf("* %s\n", d)
	}
	return nil
}

func fieldsWithDefaults() string {
	fields := []string{}
	for _, s := range settings {
		if s.validDefaults != nil {
			fields = append(fields, " * "+s.name)
		}
	}
	return strings.Join(fields, "\n")
}

func init() {
	ConfigCmd.AddCommand(configDefaultsCommand)
}
