/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	cmdConfig "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

const cacheListFormat = "- {{.CacheImageName}}\n"

type CacheListTemplate struct {
	CacheImageName string
}

// listCacheCmd represents the cache list command
var listCacheCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available images from the local cache.",
	Long:  "List all available images from the local cache.",
	Run: func(cmd *cobra.Command, args []string) {
		// list images from config file
		images, err := cmdConfig.ListConfigMap(constants.Cache)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing image entries from config: %s\n", err)
			os.Exit(1)
		}
		if err := cacheList(images); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing images: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	cacheCmd.AddCommand(listCacheCmd)
}

func cacheList(images map[string]interface{}) error {
	for imageName := range images {
		tmpl, err := template.New("list").Parse(cacheListFormat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating list template: %s\n", err)
			os.Exit(1)
		}
		listTmplt := CacheListTemplate{imageName}
		err = tmpl.Execute(os.Stdout, listTmplt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing list template: %s\n", err)
			os.Exit(1)
		}
	}
	return nil
}
