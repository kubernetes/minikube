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
	"os"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
)

var statusFormat string

type Status struct {
	MinikubeStatus  string
	LocalkubeStatus string
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a local kubernetes cluster.",
	Long:  `Gets the status of a local kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		ms, err := cluster.GetHostStatus(api)
		if err != nil {
			glog.Errorln("Error getting machine status:", err)
			os.Exit(1)
		}
		ls := "N/A"
		if ms == state.Running.String() {
			ls, err = cluster.GetLocalkubeStatus(api)
		}
		if err != nil {
			glog.Errorln("Error getting machine status:", err)
			os.Exit(1)
		}
		status := Status{ms, ls}

		tmpl, err := template.New("status").Parse(statusFormat)
		if err != nil {
			glog.Errorln("Error creating status template:", err)
			os.Exit(1)
		}
		err = tmpl.Execute(os.Stdout, status)
		if err != nil {
			glog.Errorln("Error executing status template:", err)
			os.Exit(1)
		}
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusFormat, "format", constants.DefaultStatusFormat,
		`Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list accessible variables for the template, see the struct values here:https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status`)
	RootCmd.AddCommand(statusCmd)
}
