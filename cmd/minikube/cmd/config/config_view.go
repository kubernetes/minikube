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
	"fmt"
	"os"
	"text/template"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

var configViewFormat string

type ConfigViewTemplate struct {
	ConfigKey   string
	ConfigValue interface{}
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display values currently set in the minikube config file",
	Long:  "Display values currently set in the minikube config file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := configView()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	configViewCmd.Flags().StringVar(&configViewFormat, "format", constants.DefaultConfigViewFormat,
		`Go template format string for the config view output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd/config#ConfigViewTemplate`)
	ConfigCmd.AddCommand(configViewCmd)
}

func configView() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}
	for k, v := range cfg {
		tmpl, err := template.New("view").Parse(configViewFormat)
		if err != nil {
			glog.Errorln("Error creating view template:", err)
			os.Exit(1)
		}
		viewTmplt := ConfigViewTemplate{k, v}
		err = tmpl.Execute(os.Stdout, viewTmplt)
		if err != nil {
			glog.Errorln("Error executing view template:", err)
			os.Exit(1)
		}
	}
	return nil
}
