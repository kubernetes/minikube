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
	"fmt"
	"os"
	"text/template"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/util/kubeconfig"
)

var statusFormat string

type Status struct {
	MinikubeStatus   string
	ClusterStatus    string
	KubeconfigStatus string
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a local kubernetes cluster",
	Long:  `Gets the status of a local kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		ms, err := cluster.GetHostStatus(api)
		if err != nil {
			glog.Errorln("Error getting machine status:", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}

		cs := state.None.String()
		ks := state.None.String()
		if ms == state.Running.String() {
			h, err := api.Load(config.GetMachineName())
			if err != nil {
				glog.Errorln("Error getting host")
			}
			cmdRunner, err := machine.GetCommandRunner(h)
			if err != nil {
				glog.Errorln("Error getting command runner interface")
			}
			cs, err = cluster.GetLocalkubeStatus(cmdRunner)
			if err != nil {
				glog.Errorln("Error localkube status:", err)
				cmdUtil.MaybeReportErrorAndExit(err)
			}
			ip, err := cluster.GetHostDriverIP(api)
			if err != nil {
				glog.Errorln("Error host driver ip status:", err)
				cmdUtil.MaybeReportErrorAndExit(err)
			}
			kstatus, err := kubeconfig.GetKubeConfigStatus(ip, cmdUtil.GetKubeConfigPath(), config.GetMachineName())
			if err != nil {
				glog.Errorln("Error kubeconfig status:", err)
				cmdUtil.MaybeReportErrorAndExit(err)
			}
			if kstatus {
				ks = "Correctly Configured: pointing to minikube-vm at " + ip.String()
			} else {
				ks = "Misconfigured: pointing to stale minikube-vm." +
					"\nTo fix the kubectl context, run minikube update-context"
			}
		}

		status := Status{ms, cs, ks}

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
For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status`)
	RootCmd.AddCommand(statusCmd)
}
