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

package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/extract"
)

const (
	paths     = "paths"
	functions = "functions"
	output    = "output"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts all translatable strings and adds them to all translations files.",
	Long:  "Extracts all translatable strings and adds them to all translations files.",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := cmd.Flags().GetStringSlice(paths)
		if err != nil {
			exit.WithError("Invalid paths parameter", err)
		}

		f, err := cmd.Flags().GetStringSlice(functions)
		if err != nil {
			exit.WithError("Invalid functions parameter", err)
		}

		o, err := cmd.Flags().GetString(output)
		if err != nil {
			exit.WithError("Invalid output parameter", err)
		}

		extract.ExtractTranslatableStrings(p, f, o)
	},
}

func init() {
	extractCmd.Flags().StringSlice(paths, []string{"cmd", "pkg"}, "The paths to check for translatable strings.")
	extractCmd.Flags().StringSlice(functions, []string{"Translate"}, "The functions that translate strings.")
	extractCmd.Flags().String(output, "pkg/minikube/translate/translations", "The path where translation files are.")
	RootCmd.AddCommand(extractCmd)
}
