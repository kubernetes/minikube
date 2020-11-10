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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/version"
)

var (
	statusFormat string
	output       string
	layout       string
)

const (
	// Additional legacy states:

	// Configured means configured
	Configured = "Configured" // ~state.Saved
	// Misconfigured means misconfigured
	Misconfigured = "Misconfigured" // ~state.Error
	// Nonexistent means the resource does not exist
	Nonexistent = "Nonexistent" // ~state.None
	// Irrelevant is used for statuses that aren't meaningful for worker nodes
	Irrelevant = "Irrelevant"

	// New status modes, based roughly on HTTP/SMTP standards

	// 1xx signifies a transitional state. If retried, it will soon return a 2xx, 4xx, or 5xx

	Starting  = 100
	Pausing   = 101
	Unpausing = 102
	Stopping  = 110
	Deleting  = 120

	// 2xx signifies that the API Server is able to service requests

	OK      = 200
	Warning = 203

	// 4xx signifies an error that requires help from the client to resolve

	NotFound = 404
	Stopped  = 405
	Paused   = 418 // I'm a teapot!

	// 5xx signifies a server-side error (that may be retryable)

	Error               = 500
	InsufficientStorage = 507
	Unknown             = 520
)

var (
	exitCodeToHTTPCode = map[int]int{
		// exit code 26 corresponds to insufficient storage
		26: 507,
	}

	codeNames = map[int]string{
		100: "Starting",
		101: "Pausing",
		102: "Unpausing",
		110: "Stopping",
		103: "Deleting",

		200: "OK",
		203: "Warning",

		404: "NotFound",
		405: "Stopped",
		418: "Paused",

		500: "Error",
		507: "InsufficientStorage",
		520: "Unknown",
	}

	codeDetails = map[int]string{
		507: "/var is almost out of disk space",
	}
)

// Status holds string representations of component states
type Status struct {
	Name       string
	Host       string
	Kubelet    string
	APIServer  string
	Kubeconfig string
	Worker     bool
}

// ClusterState holds a cluster state representation
type ClusterState struct {
	BaseState

	BinaryVersion string
	Components    map[string]BaseState
	Nodes         []NodeState
}

// NodeState holds a node state representation
type NodeState struct {
	BaseState
	Components map[string]BaseState `json:",omitempty"`
}

// BaseState holds a component state representation, such as "apiserver" or "kubeconfig"
type BaseState struct {
	// Name is the name of the object
	Name string

	// StatusCode is an HTTP-like status code for this object
	StatusCode int
	// Name is a human-readable name for the status code
	StatusName string
	// StatusDetail is long human-readable string describing why this particular status code was chosen
	StatusDetail string `json:",omitempty"` // Not yet implemented

	// Step is which workflow step the object is at.
	Step string `json:",omitempty"`
	// StepDetail is a long human-readable string describing the step
	StepDetail string `json:",omitempty"`
}

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
	Run: func(cmd *cobra.Command, args []string) {
		output = strings.ToLower(output)
		if output != "text" && statusFormat != defaultStatusFormat {
			exit.Message(reason.Usage, "Cannot use both --output and --format options")
		}

		out.SetJSON(output == "json")

		cname := ClusterFlagValue()
		api, cc := mustload.Partial(cname)

		var statuses []*Status

		if nodeName != "" || statusFormat != defaultStatusFormat && len(cc.Nodes) > 1 {
			n, _, err := node.Retrieve(*cc, nodeName)
			if err != nil {
				exit.Error(reason.GuestNodeRetrieve, "retrieving node", err)
			}

			st, err := nodeStatus(api, *cc, *n)
			if err != nil {
				klog.Errorf("status error: %v", err)
			}
			statuses = append(statuses, st)
		} else {
			for _, n := range cc.Nodes {
				machineName := driver.MachineName(*cc, n)
				klog.Infof("checking status of %s ...", machineName)
				st, err := nodeStatus(api, *cc, n)
				klog.Infof("%s status: %+v", machineName, st)

				if err != nil {
					klog.Errorf("status error: %v", err)
				}
				if st.Host == Nonexistent {
					klog.Errorf("The %q host does not exist!", machineName)
				}
				statuses = append(statuses, st)
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
				if err := clusterStatusJSON(statuses, os.Stdout); err != nil {
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

		os.Exit(exitCode(statuses))
	},
}

// exitCode calcluates the appropriate exit code given a set of status messages
func exitCode(statuses []*Status) int {
	c := 0
	for _, st := range statuses {
		if st.Host != state.Running.String() {
			c |= minikubeNotRunningStatusFlag
		}
		if (st.APIServer != state.Running.String() && st.APIServer != Irrelevant) || st.Kubelet != state.Running.String() {
			c |= clusterNotRunningStatusFlag
		}
		if st.Kubeconfig != Configured && st.Kubeconfig != Irrelevant {
			c |= k8sNotRunningStatusFlag
		}
	}
	return c
}

// nodeStatus looks up the status of a node
func nodeStatus(api libmachine.API, cc config.ClusterConfig, n config.Node) (*Status, error) {
	controlPlane := n.ControlPlane
	name := driver.MachineName(cc, n)

	st := &Status{
		Name:       name,
		Host:       Nonexistent,
		APIServer:  Nonexistent,
		Kubelet:    Nonexistent,
		Kubeconfig: Nonexistent,
		Worker:     !controlPlane,
	}

	hs, err := machine.Status(api, name)
	klog.Infof("%s host status = %q (err=%v)", name, hs, err)
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
		klog.Infof("host is not running, skipping remaining checks")
		st.APIServer = st.Host
		st.Kubelet = st.Host
		st.Kubeconfig = st.Host
		return st, nil
	}

	// We have a fully operational host, now we can check for details
	if _, err := cluster.DriverIP(api, name); err != nil {
		klog.Errorf("failed to get driver ip: %v", err)
		st.Host = state.Error.String()
		return st, err
	}

	st.Kubeconfig = Configured
	if !controlPlane {
		st.Kubeconfig = Irrelevant
		st.APIServer = Irrelevant
	}

	host, err := machine.LoadHost(api, name)
	if err != nil {
		return st, err
	}

	cr, err := machine.CommandRunner(host)
	if err != nil {
		return st, err
	}

	// Check storage
	p, err := machine.DiskUsed(cr, "/var")
	if err != nil {
		klog.Errorf("failed to get storage capacity of /var: %v", err)
		st.Host = state.Error.String()
		return st, err
	}
	if p >= 99 {
		st.Host = codeNames[InsufficientStorage]
	}

	stk := kverify.ServiceStatus(cr, "kubelet")
	st.Kubelet = stk.String()

	// Early exit for worker nodes
	if !controlPlane {
		return st, nil
	}

	hostname, _, port, err := driver.ControlPlaneEndpoint(&cc, &n, host.DriverName)
	if err != nil {
		klog.Errorf("forwarded endpoint: %v", err)
		st.Kubeconfig = Misconfigured
	} else {
		err := kubeconfig.VerifyEndpoint(cc.Name, hostname, port)
		if err != nil {
			klog.Errorf("kubeconfig endpoint: %v", err)
			st.Kubeconfig = Misconfigured
		}
	}

	sta, err := kverify.APIServerStatus(cr, hostname, port)
	klog.Infof("%s apiserver status = %s (err=%v)", name, stk, err)

	if err != nil {
		klog.Errorln("Error apiserver status:", err)
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
	statusCmd.Flags().StringVarP(&layout, "layout", "l", "nodes",
		`output layout (EXPERIMENTAL, JSON only): 'nodes' or 'cluster'`)
	statusCmd.Flags().StringVarP(&nodeName, "node", "n", "", "The node to check status for. Defaults to control plane. Leave blank with default format for status on all nodes.")
}

func statusText(st *Status, w io.Writer) error {
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
	if st.Kubeconfig == Misconfigured {
		_, err := w.Write([]byte("\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n"))
		return err
	}
	return nil
}

func statusJSON(st []*Status, w io.Writer) error {
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

// readEventLog reads cloudevent logs from $MINIKUBE_HOME/profiles/<name>/events.json
func readEventLog(name string) ([]cloudevents.Event, time.Time, error) {
	path := localpath.EventLog(name)

	st, err := os.Stat(path)
	if err != nil {
		return nil, time.Time{}, errors.Wrap(err, "stat")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, st.ModTime(), errors.Wrap(err, "open")
	}
	var events []cloudevents.Event

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ev := cloudevents.NewEvent()
		if err = json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			return events, st.ModTime(), err
		}
		events = append(events, ev)
	}

	return events, st.ModTime(), nil
}

// clusterState converts Status structs into a ClusterState struct
func clusterState(sts []*Status) ClusterState {
	statusName := sts[0].APIServer
	if sts[0].Host == codeNames[InsufficientStorage] {
		statusName = sts[0].Host
	}
	sc := statusCode(statusName)
	cs := ClusterState{
		BinaryVersion: version.GetVersion(),

		BaseState: BaseState{
			Name:         ClusterFlagValue(),
			StatusCode:   sc,
			StatusName:   statusName,
			StatusDetail: codeDetails[sc],
		},

		Components: map[string]BaseState{
			"kubeconfig": {Name: "kubeconfig", StatusCode: statusCode(sts[0].Kubeconfig)},
		},
	}

	for _, st := range sts {
		ns := NodeState{
			BaseState: BaseState{
				Name:       st.Name,
				StatusCode: statusCode(st.Host),
			},
			Components: map[string]BaseState{
				"kubelet": {Name: "kubelet", StatusCode: statusCode(st.Kubelet)},
			},
		}

		if st.APIServer != Irrelevant {
			ns.Components["apiserver"] = BaseState{Name: "apiserver", StatusCode: statusCode(st.APIServer)}
		}

		// Convert status codes to status names
		ns.StatusName = codeNames[ns.StatusCode]
		for k, v := range ns.Components {
			v.StatusName = codeNames[v.StatusCode]
			ns.Components[k] = v
		}

		cs.Nodes = append(cs.Nodes, ns)
	}

	evs, mtime, err := readEventLog(sts[0].Name)
	if err != nil {
		klog.Errorf("unable to read event log: %v", err)
		return cs
	}

	transientCode := 0
	var finalStep map[string]string

	for _, ev := range evs {
		//		klog.Infof("read event: %+v", ev)
		if ev.Type() == "io.k8s.sigs.minikube.step" {
			var data map[string]string
			err := ev.DataAs(&data)
			if err != nil {
				klog.Errorf("unable to parse data: %v\nraw data: %s", err, ev.Data())
				continue
			}

			switch data["name"] {
			case string(register.InitialSetup):
				transientCode = Starting
			case string(register.Done):
				transientCode = 0
			case string(register.Stopping):
				klog.Infof("%q == %q", data["name"], register.Stopping)
				transientCode = Stopping
			case string(register.Deleting):
				transientCode = Deleting
			case string(register.Pausing):
				transientCode = Pausing
			case string(register.Unpausing):
				transientCode = Unpausing
			}

			finalStep = data
			klog.Infof("transient code %d (%q) for step: %+v", transientCode, codeNames[transientCode], data)
		}
		if ev.Type() == "io.k8s.sigs.minikube.error" {
			var data map[string]string
			err := ev.DataAs(&data)
			if err != nil {
				klog.Errorf("unable to parse data: %v\nraw data: %s", err, ev.Data())
				continue
			}
			exitCode, err := strconv.Atoi(data["exitcode"])
			if err != nil {
				klog.Errorf("exit code not found: %v", err)
				continue
			}
			if val, ok := exitCodeToHTTPCode[exitCode]; ok {
				exitCode = val
			}
			transientCode = exitCode
			for _, n := range cs.Nodes {
				n.StatusCode = transientCode
				n.StatusName = codeNames[n.StatusCode]
			}

			klog.Infof("transient code %d (%q) for step: %+v", transientCode, codeNames[transientCode], data)
		}
	}

	if finalStep != nil {
		if mtime.Before(time.Now().Add(-10 * time.Minute)) {
			klog.Warningf("event stream is too old (%s) to be considered a transient state", mtime)
		} else {
			cs.Step = strings.TrimSpace(finalStep["name"])
			cs.StepDetail = strings.TrimSpace(finalStep["message"])
			if transientCode != 0 {
				cs.StatusCode = transientCode
			}
		}
	}

	cs.StatusName = codeNames[cs.StatusCode]
	cs.StatusDetail = codeDetails[cs.StatusCode]
	return cs
}

// statusCode returns a status code number given a name
func statusCode(st string) int {
	// legacy names
	switch st {
	case "Running", "Configured":
		return OK
	case "Misconfigured":
		return Error
	}

	// new names
	for code, name := range codeNames {
		if name == st {
			return code
		}
	}

	return Unknown
}

func clusterStatusJSON(statuses []*Status, w io.Writer) error {
	cs := clusterState(statuses)

	bs, err := json.Marshal(cs)
	if err != nil {
		return errors.Wrap(err, "marshal")
	}

	_, err = w.Write(bs)
	return err
}
