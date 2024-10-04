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
	"bytes"
	"encoding/json"
	"testing"

	"k8s.io/minikube/pkg/minikube/cluster"
)

func TestExitCode(t *testing.T) {
	var tests = []struct {
		name  string
		want  int
		state *cluster.Status
	}{
		{"ok", 0, &cluster.Status{Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: cluster.Configured}},
		{"paused", 2, &cluster.Status{Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: cluster.Configured}},
		{"down", 7, &cluster.Status{Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: cluster.Misconfigured}},
		{"missing", 7, &cluster.Status{Host: "Nonexistent", Kubelet: "Nonexistent", APIServer: "Nonexistent", Kubeconfig: "Nonexistent"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := exitCode([]*cluster.Status{tc.state})
			if got != tc.want {
				t.Errorf("exitcode(%+v) = %d, want: %d", tc.state, got, tc.want)
			}
		})
	}
}

func TestStatusText(t *testing.T) {
	var tests = []struct {
		name  string
		state *cluster.Status
		want  string
	}{
		{
			name:  "ok",
			state: &cluster.Status{Name: "minikube", Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: cluster.Configured, TimeToStop: "10m"},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Running\napiserver: Running\nkubeconfig: Configured\ntimeToStop: 10m\n\n",
		},
		{
			name:  "paused",
			state: &cluster.Status{Name: "minikube", Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: cluster.Configured},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Stopped\napiserver: Paused\nkubeconfig: Configured\n\n",
		},
		{
			name:  "down",
			state: &cluster.Status{Name: "minikube", Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: cluster.Misconfigured},
			want:  "minikube\ntype: Control Plane\nhost: Stopped\nkubelet: Stopped\napiserver: Stopped\nkubeconfig: Misconfigured\n\n\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			err := statusText(tc.state, &b)
			if err != nil {
				t.Errorf("text(%+v) error: %v", tc.state, err)
			}

			got := b.String()
			if got != tc.want {
				t.Errorf("text(%+v) = %q, want: %q", tc.state, got, tc.want)
			}
		})
	}
}

func TestStatusJSON(t *testing.T) {
	var tests = []struct {
		name  string
		state *cluster.Status
	}{
		{"ok", &cluster.Status{Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: cluster.Configured, TimeToStop: "10m"}},
		{"paused", &cluster.Status{Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: cluster.Configured}},
		{"down", &cluster.Status{Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: cluster.Misconfigured}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			err := statusJSON([]*cluster.Status{tc.state}, &b)
			if err != nil {
				t.Errorf("json(%+v) error: %v", tc.state, err)
			}

			st := &cluster.Status{}
			if err := json.Unmarshal(b.Bytes(), st); err != nil {
				t.Errorf("json(%+v) unmarshal error: %v", tc.state, err)
			}
		})
	}
}
