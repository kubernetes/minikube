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

// Special note:
//
// This code is invoked by "minikube status" if --legacy=true
//
// This file, and associated mode, should be removed after minikube v1.15 (~October 2020)

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
)

const (
	// # Additional states used by kubeconfig:

	// Configured means configured
	Configured = "Configured" // ~state.Saved
	// Misconfigured means misconfigured
	Misconfigured = "Misconfigured" // ~state.Error

	// # Additional states used for clarity:

	// Nonexistent means nonexistent
	Nonexistent = "Nonexistent" // ~state.None
	// Irrelevant is used for statuses that aren't meaningful for worker nodes
	Irrelevant = "Irrelevant"
)

// LegacyStatus holds string representations of component states
type LegacyStatus struct {
	Name       string
	Host       string
	Kubelet    string
	APIServer  string
	Kubeconfig string
	Worker     bool
}

const (
	minikubeNotRunningStatusFlag = 1 << 0
	clusterNotRunningStatusFlag  = 1 << 1
	k8sNotRunningStatusFlag      = 1 << 2
	legacyStatusTmpl             = `{{.Name}}
type: Control Plane
host: {{.Host}}
kubelet: {{.Kubelet}}
apiserver: {{.APIServer}}
kubeconfig: {{.Kubeconfig}}

`
	legacyWorkerTmpl = `{{.Name}}
type: Worker
host: {{.Host}}
kubelet: {{.Kubelet}}

`
)

// legacyExitCode returns status of minikube's VM, cluster and Kubernetes encoded on it's bits in this order from right to left.
// Eg: 7 meaning: 1 (for minikube NOK) + 2 (for cluster NOK) + 4 (for Kubernetes NOK)`,
func legacyExitCode(statuses []*LegacyStatus) int {
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

func legacyStatus(st cluster.State, sc statusConfig) []*LegacyStatus {
	var statuses []*LegacyStatus

	for _, n := range cc.Nodes {
		machineName := driver.MachineName(*cc, n)

		st, err := legacyStatus(api, *cc, n)
		if err != nil {
			glog.Errorf("status error: %v", err)
		}

		if st.Host == Nonexistent {
			glog.Errorf("The %q host does not exist!", machineName)
		}

		statuses = append(statuses, st)
	}

	glog.Infof("collected: %+v", statuses)

	switch sc.Format {
	case "text":
		for _, st := range statuses {
			tmpl := sc.Template
			if tmpl == "" {
				if st.Worker {
					tmpl = legacyWorkerTmpl
				} else {
					tmpl = legacyStatusTmpl
				}
			}
			if err := legacyStatusText(st, os.Stdout, tmpl); err != nil {
				exit.WithError("status text failure", err)
			}
		}
	case "json":
		if err := legacyStatusJSON(statuses, os.Stdout); err != nil {
			exit.WithError("status json failure", err)
		}
	default:
		exit.WithCodeT(exit.BadUsage, fmt.Sprintf("invalid output format: %s. Valid values: 'text', 'json'", sc.Format))
	}

	os.Exit(legacyExitCode(statuses))

	ls := []*LegacyStatus{}

	/*
		st := &ClusterState{
			Name:       name,
			Host:       Nonexistent,
			APIServer:  Nonexistent,
			Kubelet:    Nonexistent,
			Kubeconfig: Nonexistent,
			Worker:     !controlPlane,
		}
	*/

	for _, n := range cs.Nodes {
		ls = append(ls, &LegacyStatus{
			Name: n.Name,
		})
	}

	return ls
}

func legacyStatusText(st *LegacyStatus, w io.Writer, format string) error {
	glog.Infof("legacy status text for: %+v", st)
	glog.Infof("format: %s", format)

	tmpl, err := template.New("status").Parse(format)
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	if err := tmpl.Execute(w, st); err != nil {
		return errors.Wrap(err, "execute")
	}

	if st.Kubeconfig == Misconfigured {
		_, err := w.Write([]byte("\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n"))
		return errors.Wrap(err, "write")
	}

	return nil
}

func legacyStatusJSON(st []*LegacyStatus, w io.Writer) error {
	glog.Infof("legacy JSON text for: %+v", st)

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
