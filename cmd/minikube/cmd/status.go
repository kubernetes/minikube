/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/state"
)

/* Example output:

cluster:    OK (healthz returned in 0.15s)
kubeconfig: OK (endpoint: 127.0.0.1:8080)

nodes:
* minikube: OK (load: 10.40)
*/

const (
	clusterTmpl = `cluster:  {{.Condition}} ({{.Message}})

components:
	{{ range .Components }}{{ .Name }}: {{.Condition}} ({{.Message}}){{ end}}

nodes:
  {{ range .Nodes }}{{ .Name }}: {{.State}} ({{.Message}}){{ end}}`

	nodeTmpl = `{{ .Name }}: {{.State}} ({{.Message}})
	
components:
	{{ range .Components }}{{ .Name }}: {{.Condition}} ({{.Message}}){{ end}}`
)

var (
	outputFlag string
	compatFlag string
	nodeFlag   string
	formatFlag string

	exitCodeMap = map[state.Condition]int{
		state.Unknown:     exit.Software,
		state.Nonexistent: exit.NoInput,
		state.Unavailable: exit.Unavailable,
		state.OK:          0,
		state.Warning:     0,
		state.Error:       exit.Failure,

		state.Starting: 0,
		state.Pausing:  0,
		state.Paused:   0,
		state.Stopping: 0,
		state.Stopped:  0,
	}
)

type statusConfig struct {
	Cluster string
	Node    string

	Format     string
	Template   string
	compatMode bool
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Output the status of a local Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		outputFormat := strings.ToLower(outputFlag)

		// Backwards compatibility while we transition from legacy to new output
		compatMode := false
		glog.Infof("compat: %q", viper.GetString("compat"))
		if cmd.Flags().Changed("compat") {
			glog.Infof("compat changed")
		}

		if compatFlag == "true" {
			glog.Infof("compat=true")
			compatMode = true
		} else if compatFlag == "auto" {
			glog.Infof("compat=auto")
			// Preserve existing behavior for programattic use cases
			if cmd.Flags().Changed("output") || cmd.Flags().Changed("format") {
				glog.Errorf("Automatically enabling --compat=true. Please explicitly pass --compat=(false|true) to avoid warning")
				compatMode = true
			}
		}

		tmpl := ""

		if cmd.Flags().Changed("format") {
			if outputFormat != "text" {
				exit.UsageT("Cannot use both --output and --format options")
			}
			tmpl = formatFlag
		}

		cname := ClusterFlagValue()
		api, cc := mustload.Partial(cname)

		err, cs := cluster.Status(api, cc)
		if err != nil {
			exit.WithError("status", cs)
		}

		node := viper.GetString("node")
		if node != "" {
			nodeStatus(api, cc, statusConfig{Cluster: cname, Node: node, Format: outputFormat, Template: tmpl, compatMode: compatMode})
		}

		clusterStatus(api, cc, statusConfig{Cluster: cname, Format: outputFormat, Template: tmpl, compatMode: compatMode})
	},
}

func outputStatus(api libmachine.API, cc *config.ClusterConfig, sc statusConfig) {
	glog.Errorf("status config: %+v", sc)

	cs := cluster.Status(api, cc)
	if sc.CompatMode {
		return legacyStatus(cs, sc)
	}

}

func init() {
	statusCmd.Flags().StringVarP(&formatFlag, "format", "f", "", `Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/

	For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status`)
	statusCmd.Flags().StringVarP(&outputFlag, "output", "o", "text", "Format for status output. Options are: json, text")
	statusCmd.Flags().StringVarP(&nodeFlag, "nodee", "n", "", "The node to check status for. Defaults to control plane. Leave blank with default format for status on all nodes.")
	statusCmd.Flags().StringVar(&compatFlag, "compat", "auto", "Output status in backwards-compatible format (currently defaults to true when --output or --format are set)")
}
