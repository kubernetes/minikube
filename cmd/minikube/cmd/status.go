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

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
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

const (
	// Additional states used by kubeconfig
	Configured    = "Configured"    // ~state.Saved
	Misconfigured = "Misconfigured" // ~state.Error
	// Additional states used for clarity
	Nonexistent = "Nonexistent" // ~state.None
)

// Status holds string representations of component states
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

		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithCodeT(exit.Unavailable, "Error getting client: {{.error}}", out.V{"error": err})
		}
		defer api.Close()

		machineName := viper.GetString(config.MachineProfile)
		st, err := status(api, machineName)
		if err != nil {
			glog.Errorf("status error: %v", err)
		}
		if st.Host == Nonexistent {
			glog.Errorf("The %q cluster does not exist!", machineName)
		}

		switch strings.ToLower(output) {
		case "text":
			if err := statusText(st, os.Stdout); err != nil {
				exit.WithError("status text failure", err)
			}
		case "json":
			if err := statusJSON(st, os.Stdout); err != nil {
				exit.WithError("status json failure", err)
			}
		default:
			exit.WithCodeT(exit.BadUsage, fmt.Sprintf("invalid output format: %s. Valid values: 'text', 'json'", output))
		}

		os.Exit(exitCode(st))
	},
}

func exitCode(st *Status) int {
	c := 0
	if st.Host != state.Running.String() {
		c |= minikubeNotRunningStatusFlag
	}
	if st.APIServer != state.Running.String() || st.Kubelet != state.Running.String() {
		c |= clusterNotRunningStatusFlag
	}
	if st.Kubeconfig != Configured {
		c |= k8sNotRunningStatusFlag
	}
	return c
}

func status(api libmachine.API, name string) (*Status, error) {
	st := &Status{
		Host:       Nonexistent,
		APIServer:  Nonexistent,
		Kubelet:    Nonexistent,
		Kubeconfig: Nonexistent,
	}

	hs, err := machine.GetHostStatus(api, name)
	glog.Infof("%s host status = %q (err=%v)", name, hs, err)
	if err != nil {
		return st, errors.Wrap(err, "host")
	}

	// We have no record of this host. Return nonexistent struct
	if hs == state.None.String() {
		return st, nil
	}
	st.Host = hs

	// If it's not running, quickly bail out rather than delivering conflicting messages
	if st.Host != state.Running.String() {
		glog.Infof("host is not running, skipping remaining checks")
		st.APIServer = st.Host
		st.Kubelet = st.Host
		st.Kubeconfig = st.Host
		return st, nil
	}

	// We have a fully operational host, now we can check for details
	ip, err := cluster.GetHostDriverIP(api, name)
	if err != nil {
		glog.Errorln("Error host driver ip status:", err)
		st.APIServer = state.Error.String()
		return st, err
	}

	port, err := kubeconfig.Port(name)
	if err != nil {
		glog.Warningf("unable to get port: %v", err)
		port = constants.APIServerPort
	}

	st.Kubeconfig = Misconfigured
	ok, err := kubeconfig.IsClusterInConfig(ip, name)
	glog.Infof("%s is in kubeconfig at ip %s: %v (err=%v)", name, ip, ok, err)
	if ok {
		st.Kubeconfig = Configured
	}

	host, err := machine.CheckIfHostExistsAndLoad(api, name)
	if err != nil {
		return st, err
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		return st, err
	}

	stk, err := kverify.KubeletStatus(cr)
	glog.Infof("%s kubelet status = %s (err=%v)", name, stk, err)

	if err != nil {
		glog.Warningf("kubelet err: %v", err)
		st.Kubelet = state.Error.String()
	} else {
		st.Kubelet = stk.String()
	}

	sta, err := kverify.APIServerStatus(cr, ip, port)
	glog.Infof("%s apiserver status = %s (err=%v)", name, stk, err)

	if err != nil {
		glog.Errorln("Error apiserver status:", err)
		st.APIServer = state.Error.String()
	} else {
		st.APIServer = sta.String()
	}

	return st, nil
}

func init() {
	statusCmd.Flags().StringVarP(&statusFormat, "format", "f", defaultStatusFormat,
		`Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status`)
	statusCmd.Flags().StringVarP(&output, "output", "o", "text",
		`minikube status --output OUTPUT. json, text`)
}

func statusText(st *Status, w io.Writer) error {
	tmpl, err := template.New("status").Parse(statusFormat)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(w, st); err != nil {
		return err
	}
	if st.Kubeconfig == Misconfigured {
		_, err := w.Write([]byte("\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n"))
		return err
	}
	return nil
}

func statusJSON(st *Status, w io.Writer) error {
	js, err := json.Marshal(st)
	if err != nil {
		return err
	}
	_, err = w.Write(js)
	return err
}
