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
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var docsPath string
var testPath string
var codePath string

// generateDocs represents the generate-docs command
var generateDocs = &cobra.Command{
	Use:     "generate-docs",
	Short:   "Populates the specified folder with documentation in markdown about minikube",
	Long:    "Populates the specified folder with documentation in markdown about minikube",
	Example: "minikube generate-docs --path <FOLDER_PATH>",
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		// if directory does not exist
		st, err := os.Stat(docsPath)
		if err != nil || !st.IsDir() {
			exit.Message(reason.Usage, "Unable to generate the documentation. Please ensure that the path specified is a directory, exists & you have permission to write to it.")
		}

		// generate docs
		if err := generate.Docs(RootCmd, docsPath, testPath, codePath); err != nil {
			exit.Error(reason.InternalGenerateDocs, "Unable to generate docs", err)
		}
		out.Step(style.Documentation, "Docs have been saved at - {{.path}}", out.V{"path": docsPath})
		out.Step(style.Documentation, "Test docs have been saved at - {{.path}}", out.V{"path": testPath})
		out.Step(style.Documentation, "Error code docs have been saved at - {{.path}}", out.V{"path": codePath})
	},
}

func init() {
	generateDocs.Flags().StringVar(&docsPath, "path", "", "The path on the file system where the docs in markdown need to be saved")
	generateDocs.Flags().StringVar(&testPath, "test-path", "", "The path on the file system where the testing docs in markdown need to be saved")
	generateDocs.Flags().StringVar(&codePath, "code-path", "", "The path on the file system where the error code docs in markdown need to be saved")
	RootCmd.AddCommand(generateDocs)
}
