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
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"

	"github.com/spf13/cobra"
)

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all configurable fields",
	Long:  "List displays all configurable fields",
	Run: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(output) {
		case "json":
			printSettingJSON()
		case "table":
			printSetting()
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'table', 'json'", output))
		}
	},
}

func init() {
	configListCmd.Flags().StringVarP(&output, "output", "o", "table", "The output format. One of 'json', 'table'")
	ConfigCmd.AddCommand(configListCmd)
}

func printSetting() {
	for _, s := range settings {
		fmt.Printf("%s - %s\n", s.name, s.description)
	}
}

func printSettingJSON() {
	var body []interface{}

	for _, s := range settings {
		body = append(body, struct {
			Name        string
			Description string
		}{
			s.name, s.description,
		})
	}

	jsonString, _ := json.Marshal(body)
	out.String(string(jsonString))
}
