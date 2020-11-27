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
	"time"

	"k8s.io/minikube/pkg/minikube/config"
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
			state: &Status{Name: "minikube", Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: Configured},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Running\napiserver: Running\nkubeconfig: Configured\n\n",
		},
		{
			name:  "paused",
			state: &Status{Name: "minikube", Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: Configured},
			want:  "minikube\ntype: Control Plane\nhost: Running\nkubelet: Stopped\napiserver: Paused\nkubeconfig: Configured\n\n",
		},
		{
			name:  "down",
			state: &Status{Name: "minikube", Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: Misconfigured},
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
		state *Status
	}{
		{"ok", &Status{Host: "Running", Kubelet: "Running", APIServer: "Running", Kubeconfig: Configured}},
		{"paused", &Status{Host: "Running", Kubelet: "Stopped", APIServer: "Paused", Kubeconfig: Configured}},
		{"down", &Status{Host: "Stopped", Kubelet: "Stopped", APIServer: "Stopped", Kubeconfig: Misconfigured}},
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

func TestScheduledStopStatusText(t *testing.T) {
	now := time.Now().Unix()
	initiationTime := time.Unix(now, 0).String()

	stopAt := time.Now().Add(time.Minute * 10).Unix()
	scheduledTime := time.Unix(stopAt, 0).String()
	var tests = []struct {
		name  string
		state *config.ScheduledStopConfig
		want  string
	}{
		{
			name:  "valid",
			state: &config.ScheduledStopConfig{InitiationTime: now, Duration: time.Minute * 10},
			want:  "type: ScheduledDuration\ninitiatedTime: " + initiationTime + "\nscheduledTime: " + scheduledTime + "\n\n",
		},
		{
			name:  "missing",
			state: &config.ScheduledStopConfig{},
			want:  "type: ScheduledDuration\ninitiatedTime: -\nscheduledTime: -\n\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			err := scheduledStopStatusText(tc.state, &b)
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

func TestScheduledStopStatusJSON(t *testing.T) {
	var tests = []struct {
		name  string
		state *config.ScheduledStopConfig
	}{
		{
			name:  "valid",
			state: &config.ScheduledStopConfig{InitiationTime: time.Now().Unix(), Duration: time.Minute * 5},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			err := scheduledStopStatusJSON(tc.state, &b)
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
