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
	"encoding/json"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/version"
)

var (
	versionOutput string
	shortVersion  bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of minikube",
	Long:  `Print the version of minikube.`,
	Run: func(command *cobra.Command, args []string) {
		minikubeVersion := version.GetVersion()
		gitCommitID := version.GetGitCommitID()
		data := map[string]string{
			"minikubeVersion": minikubeVersion,
			"commit":          gitCommitID,
		}
		switch versionOutput {
		case "":
			out.Ln("minikube version: %v", minikubeVersion)
			if !shortVersion && gitCommitID != "" {
				out.Ln("commit: %v", gitCommitID)
			}
		case "json":
			json, err := json.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalJSONMarshal, "version json failure", err)
			}
			out.Ln(string(json))
		case "yaml":
			yaml, err := yaml.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalYamlMarshal, "version yaml failure", err)
			}
			out.Ln(string(yaml))
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'yaml' or 'json'")
		}
	},
}

func init() {
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "", "One of 'yaml' or 'json'.")
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "Print just the version number.")
}
