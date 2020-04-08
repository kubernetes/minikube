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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/generate"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
)

var path string

// generateDocs represents the generate-docs command
var generateDocs = &cobra.Command{
	Use:     "generate-docs",
	Short:   "Populates the specified folder with documentation in markdown about minikube",
	Long:    "Populates the specified folder with documentation in markdown about minikube",
	Example: "minikube generate-docs --path <FOLDER_PATH>",
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {

		// if directory does not exist
		docsPath, err := os.Stat(path)
		if err != nil || !docsPath.IsDir() {
			exit.UsageT("Unable to generate the documentation. Please ensure that the path specified is a directory, exists & you have permission to write to it.")
		}

		// generate docs
		if err := generate.Docs(RootCmd, path); err != nil {
			exit.WithError("Unable to generate docs", err)
		}
		out.T(out.Documentation, "Docs have been saved at - {{.path}}", out.V{"path": path})
	},
}

func init() {
	generateDocs.Flags().StringVar(&path, "path", "", "The path on the file system where the docs in markdown need to be saved")
	RootCmd.AddCommand(generateDocs)
}
