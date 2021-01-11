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
)

func TestExitCode(t *testing.T) {
	var tests = []struct {
		name  string
		want  int
		state *Status
	}{
		{"ok", 0, &Status{Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: Configured}},
		{"paused", 2, &Status{Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: Configured}},
		{"down", 7, &Status{Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: Misconfigured}},
		{"missing", 7, &Status{Host: "Nonexistent", Kubelet: "Nonexistent", APIServer: "Nonexistent", Kubeconfig: "Nonexistent"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := exitCode([]*Status{tc.state})
			if got != tc.want {
				t.Errorf("exitcode(%+v) = %d, want: %d", tc.state, got, tc.want)
			}
		})
	}
}

func TestStatusText(t *testing.T) {
	var tests = []struct {
		name  string
		state *Status
		want  string
	}{
		{
			name:  "ok",
			state: &Status{Name: "minikube", Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: Configured, TimeToStop: "10m", IP: "192.168.39.10"},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Running\napiserver: Running\nkubeconfig: Configured\ntimeToStop: 10m\nIP: 192.168.39.10\n\n",
		},
		{
			name:  "paused",
			state: &Status{Name: "minikube", Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: Configured, TimeToStop: Nonexistent, IP: "192.168.39.10"},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Stopped\napiserver: Paused\nkubeconfig: Configured\ntimeToStop: Nonexistent\nIP: 192.168.39.10\n\n",
		},
		{
			name:  "down",
			state: &Status{Name: "minikube", Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: Misconfigured, TimeToStop: Nonexistent, IP: "192.168.39.10"},
			want:  "minikube\ntype: Control Plane\nhost: Stopped\nkubelet: Stopped\napiserver: Stopped\nkubeconfig: Misconfigured\ntimeToStop: Nonexistent\nIP: 192.168.39.10\n\n\nWARNING: Your kubectl is pointing to stale minikube-vm.\nTo fix the kubectl context, run `minikube update-context`\n",
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
				t.Errorf("text(%+v)\n got: %q\nwant: %q", tc.state, got, tc.want)
			}
		})
	}
}

func TestStatusJSON(t *testing.T) {
	var tests = []struct {
		name  string
		state *Status
	}{
		{"ok", &Status{Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: Configured, TimeToStop: "10m"}},
		{"paused", &Status{Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: Configured, TimeToStop: Nonexistent}},
		{"down", &Status{Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: Misconfigured, TimeToStop: Nonexistent}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			err := statusJSON([]*Status{tc.state}, &b)
			if err != nil {
				t.Errorf("json(%+v) error: %v", tc.state, err)
			}

			st := &Status{}
			if err := json.Unmarshal(b.Bytes(), st); err != nil {
				t.Errorf("json(%+v) unmarshal error: %v", tc.state, err)
			}
		})
	}
}
