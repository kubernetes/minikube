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
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
)

var statusFormat string
var output string

// KubeconfigStatus represents the kubeconfig status
var KubeconfigStatus = struct {
	Configured    string
	Misconfigured string
}{
	Configured:    `Configured`,
	Misconfigured: `Misconfigured`,
}

// Status represents the status
type Status struct {
	Host       string
	Kubelet    string
	APIServer  string
	Kubeconfig string
}

const (
	minikubeNotRunningStatusFlag = 1 << 0
	clusterNotRunningStatusFlag  = 1 << 1
	k8sNotRunningStatusFlag      = 1 << 2
	defaultStatusFormat          = `host: {{.Host}}
kubelet: {{.Kubelet}}
apiserver: {{.APIServer}}
kubeconfig: {{.Kubeconfig}}
`
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a local kubernetes cluster",
	Long: `Gets the status of a local kubernetes cluster.
	Exit status contains the status of minikube's VM, cluster and kubernetes encoded on it's bits in this order from right to left.
	Eg: 7 meaning: 1 (for minikube NOK) + 2 (for cluster NOK) + 4 (for kubernetes NOK)`,
	Run: func(cmd *cobra.Command, args []string) {

		if output != "text" && statusFormat != defaultStatusFormat {
			exit.UsageT("Cannot use both --output and --format options")
		}

		var returnCode = 0
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithCodeT(exit.Unavailable, "Error getting client: {{.error}}", out.V{"error": err})
		}
		defer api.Close()

		hostSt, err := cluster.GetHostStatus(api)
		if err != nil {
			exit.WithError("Error getting host status", err)
		}

		kubeletSt := state.None.String()
		kubeconfigSt := state.None.String()
		apiserverSt := state.None.String()

		if hostSt == state.Running.String() {
			clusterBootstrapper, err := getClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
			if err != nil {
				exit.WithError("Error getting bootstrapper", err)
			}
			kubeletSt, err = clusterBootstrapper.GetKubeletStatus()
			if err != nil {
				glog.Warningf("kubelet err: %v", err)
				returnCode |= clusterNotRunningStatusFlag
			} else if kubeletSt != state.Running.String() {
				returnCode |= clusterNotRunningStatusFlag
			}

			ip, err := cluster.GetHostDriverIP(api, config.GetMachineName())
			if err != nil {
				glog.Errorln("Error host driver ip status:", err)
			}

			apiserverPort, err := kubeconfig.Port(config.GetMachineName())
			if err != nil {
				// Fallback to presuming default apiserver port
				apiserverPort = constants.APIServerPort
			}

			apiserverSt, err = clusterBootstrapper.GetAPIServerStatus(ip, apiserverPort)
			if err != nil {
				glog.Errorln("Error apiserver status:", err)
			} else if apiserverSt != state.Running.String() {
				returnCode |= clusterNotRunningStatusFlag
			}

			ks, err := kubeconfig.IsClusterInConfig(ip, config.GetMachineName())
			if err != nil {
				glog.Errorln("Error kubeconfig status:", err)
			}
			if ks {
				kubeconfigSt = KubeconfigStatus.Configured
			} else {
				kubeconfigSt = KubeconfigStatus.Misconfigured
				returnCode |= k8sNotRunningStatusFlag
			}
		} else {
			returnCode |= minikubeNotRunningStatusFlag
		}

		status := Status{
			Host:       hostSt,
			Kubelet:    kubeletSt,
			APIServer:  apiserverSt,
			Kubeconfig: kubeconfigSt,
		}

		switch strings.ToLower(output) {
		case "text":
			printStatusText(status)
		case "json":
			printStatusJSON(status)
		default:
			exit.WithCodeT(exit.BadUsage, fmt.Sprintf("invalid output format: %s. Valid values: 'text', 'json'", output))
		}

		os.Exit(returnCode)
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusFormat, "format", "f", defaultStatusFormat,
		`Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status`)
	statusCmd.Flags().StringVarP(&output, "output", "o", "text",
		`minikube status --output OUTPUT. json, text`)
}

var printStatusText = func(status Status) {
	tmpl, err := template.New("status").Parse(statusFormat)
	if err != nil {
		exit.WithError("Error creating status template", err)
	}
	err = tmpl.Execute(os.Stdout, status)
	if err != nil {
		exit.WithError("Error executing status template", err)
	}
	if status.Kubeconfig == KubeconfigStatus.Misconfigured {
		out.WarningT("Warning: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`")
	}
}

var printStatusJSON = func(status Status) {

	jsonString, err := json.Marshal(status)
	if err != nil {
		exit.WithError("Error converting status to json", err)
	}
	out.String(string(jsonString))
}
