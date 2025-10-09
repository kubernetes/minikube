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
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/reason"
)

const defaultConfigViewFormat = "- {{.ConfigKey}}: {{.ConfigValue}}\n"

var viewFormat string

// ViewTemplate represents the view template
type ViewTemplate struct {
	ConfigKey   string
	ConfigValue interface{}
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display values currently set in the minikube config file",
	Long:  `Display values currently set in the minikube config file. 
	The output format can be customized using the --format flag, which accepts a Go template. 
	The config file is typically located at "~/.minikube/config/config.json".`,
	Run: func(_ *cobra.Command, _ []string) {
		err := View()
		if err != nil {
			exit.Error(reason.InternalConfigView, "config view failed", err)
		}
	},
}

func init() {
	configViewCmd.Flags().StringVar(&viewFormat, "format", defaultConfigViewFormat,
		`Go template format string for the config view output.  The format for Go templates can be found here: https://pkg.go.dev/text/template
For the list of accessible variables for the template, see the struct values here: https://pkg.go.dev/k8s.io/minikube/cmd/minikube/cmd/config#ConfigViewTemplate`)
	ConfigCmd.AddCommand(configViewCmd)
}

// View displays the current config
func View() error {
	cfg, err := config.ReadConfig(localpath.ConfigFile())
	if err != nil {
		return err
	}
	for k, v := range cfg {
		tmpl, err := template.New("view").Parse(viewFormat)
		if err != nil {
			exit.Error(reason.InternalViewTmpl, "Error creating view template", err)
		}
		viewTmplt := ViewTemplate{k, v}
		err = tmpl.Execute(os.Stdout, viewTmplt)
		if err != nil {
			exit.Error(reason.InternalViewExec, "Error executing view template", err)
		}
	}
	return nil
}
