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
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

var (
	statusFormat string
	output       string
	layout       string
	watch        time.Duration
)

const (
	minikubeNotRunningStatusFlag = 1 << 0
	clusterNotRunningStatusFlag  = 1 << 1
	k8sNotRunningStatusFlag      = 1 << 2
	defaultStatusFormat          = `{{.Name}}
type: Control Plane
host: {{.Host}}
kubelet: {{.Kubelet}}
apiserver: {{.APIServer}}
kubeconfig: {{.Kubeconfig}}
{{- if .TimeToStop }}
timeToStop: {{.TimeToStop}}
{{- end }}
{{- if .DockerEnv }}
docker-env: {{.DockerEnv}}
{{- end }}
{{- if .PodManEnv }}
podman-env: {{.PodManEnv}}
{{- end }}

`
	workerStatusFormat = `{{.Name}}
type: Worker
host: {{.Host}}
kubelet: {{.Kubelet}}

`
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of a local Kubernetes cluster",
	Long: `Gets the status of a local Kubernetes cluster.
	Exit status contains the status of minikube's VM, cluster and Kubernetes encoded on it's bits in this order from right to left.
	Eg: 7 meaning: 1 (for minikube NOK) + 2 (for cluster NOK) + 4 (for Kubernetes NOK)`,
	Run: func(cmd *cobra.Command, _ []string) {
		output = strings.ToLower(output)
		if output != "text" && statusFormat != defaultStatusFormat {
			exit.Message(reason.Usage, "Cannot use both --output and --format options")
		}

		out.SetJSON(output == "json")
		go notify.MaybePrintUpdateTextFromGithub()

		cname := ClusterFlagValue()
		api, cc := mustload.Partial(cname)

		duration := watch
		if !cmd.Flags().Changed("watch") || watch < 0 {
			duration = 0
		}
		writeStatusesAtInterval(duration, api, cc)
	},
}

// writeStatusesAtInterval writes statuses in a given output format - at intervals defined by duration
func writeStatusesAtInterval(duration time.Duration, api libmachine.API, cc *config.ClusterConfig) {
	for {
		var statuses []*cluster.Status

		if nodeName != "" || statusFormat != defaultStatusFormat && len(cc.Nodes) > 1 {
			n, _, err := node.Retrieve(*cc, nodeName)
			if err != nil {
				exit.Error(reason.GuestNodeRetrieve, "retrieving node", err)
			}

			st, err := cluster.NodeStatus(api, *cc, *n)
			if err != nil {
				klog.Errorf("status error: %v", err)
			}
			statuses = append(statuses, st)
		} else {
			var err error
			statuses, err = cluster.GetStatus(api, cc)
			if err != nil {
				klog.Errorf("status error: %v", err)
			}
		}

		switch output {
		case "text":
			for _, st := range statuses {
				if err := statusText(st, os.Stdout); err != nil {
					exit.Error(reason.InternalStatusText, "status text failure", err)
				}
			}
		case "json":
			// Layout is currently only supported for JSON mode
			if layout == "cluster" {
				if err := clusterStatusJSON(statuses, os.Stdout, cc); err != nil {
					exit.Error(reason.InternalStatusJSON, "status json failure", err)
				}
			} else {
				if err := statusJSON(statuses, os.Stdout); err != nil {
					exit.Error(reason.InternalStatusJSON, "status json failure", err)
				}
			}
		default:
			exit.Message(reason.Usage, fmt.Sprintf("invalid output format: %s. Valid values: 'text', 'json'", output))
		}

		if duration == 0 {
			os.Exit(exitCode(statuses))
		}
		time.Sleep(duration)
	}
}

// exitCode calculates the appropriate exit code given a set of status messages
func exitCode(statuses []*cluster.Status) int {
	c := 0
	for _, st := range statuses {
		if st.Host != state.Running.String() {
			c |= minikubeNotRunningStatusFlag
		}
		if (st.APIServer != state.Running.String() && st.APIServer != cluster.Irrelevant) || st.Kubelet != state.Running.String() {
			c |= clusterNotRunningStatusFlag
		}
		if st.Kubeconfig != cluster.Configured && st.Kubeconfig != cluster.Irrelevant {
			c |= k8sNotRunningStatusFlag
		}
	}
	return c
}

func init() {
	statusCmd.Flags().StringVarP(&statusFormat, "format", "f", defaultStatusFormat,
		`Go template format string for the status output.  The format for Go templates can be found here: https://pkg.go.dev/text/template
For the list accessible variables for the template, see the struct values here: https://pkg.go.dev/k8s.io/minikube/cmd/minikube/cmd#Status`)
	statusCmd.Flags().StringVarP(&output, "output", "o", "text",
		`minikube status --output OUTPUT. json, text`)
	statusCmd.Flags().StringVarP(&layout, "layout", "l", "nodes",
		`output layout (EXPERIMENTAL, JSON only): 'nodes' or 'cluster'`)
	statusCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to check status for. Defaults to control plane. Leave blank with default format for status on all nodes.")
	statusCmd.Flags().DurationVarP(&watch, "watch", "w", 1*time.Second, "Continuously listing/getting the status with optional interval duration.")
	statusCmd.Flags().Lookup("watch").NoOptDefVal = "1s"
}

func statusText(st *cluster.Status, w io.Writer) error {
	tmpl, err := template.New("status").Parse(statusFormat)
	if st.Worker && statusFormat == defaultStatusFormat {
		tmpl, err = template.New("worker-status").Parse(workerStatusFormat)
	}
	if err != nil {
		return err
	}
	if err := tmpl.Execute(w, st); err != nil {
		return err
	}
	if st.Kubeconfig == cluster.Misconfigured {
		_, err := w.Write([]byte("\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n"))
		return err
	}
	return nil
}

func statusJSON(st []*cluster.Status, w io.Writer) error {
	var js []byte
	var err error
	// Keep backwards compat with single node clusters to not break anyone
	if len(st) == 1 {
		js, err = json.Marshal(st[0])
	} else {
		js, err = json.Marshal(st)
	}
	if err != nil {
		return err
	}
	_, err = w.Write(js)
	return err
}

func clusterStatusJSON(statuses []*cluster.Status, w io.Writer, cc *config.ClusterConfig) error {
	cs := cluster.GetState(statuses, ClusterFlagValue(), cc)

	bs, err := json.Marshal(cs)
	if err != nil {
		return errors.Wrap(err, "marshal")
	}

	_, err = w.Write(bs)
	return err
}
