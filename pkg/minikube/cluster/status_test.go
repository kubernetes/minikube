/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"io"
	"os"
	"testing"

	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
)

func TestGetStatePropagatesErrorEventToNodes(t *testing.T) {
	const profile = "minikube"
	t.Setenv(localpath.MinikubeHome, t.TempDir())

	register.SetOutputFile(io.Discard)
	defer register.SetOutputFile(os.Stdout)
	register.SetEventLogPath(localpath.EventLog(profile))
	register.PrintErrorExitCode("not enough storage", reason.ExInsufficientStorage)

	got := GetState([]*Status{
		{
			Name:       profile,
			Host:       state.Running.String(),
			Kubelet:    state.Running.String(),
			APIServer:  state.Running.String(),
			Kubeconfig: Configured,
		},
		{
			Name:       profile + "-m02",
			Host:       state.Running.String(),
			Kubelet:    state.Running.String(),
			APIServer:  Irrelevant,
			Kubeconfig: Configured,
			Worker:     true,
		},
	}, profile, &config.ClusterConfig{Name: profile})

	if len(got.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(got.Nodes))
	}
	for _, node := range got.Nodes {
		if node.StatusCode != InsufficientStorage {
			t.Errorf("node %q StatusCode = %d, want %d", node.Name, node.StatusCode, InsufficientStorage)
		}
		if node.StatusName != codeNames[InsufficientStorage] {
			t.Errorf("node %q StatusName = %q, want %q", node.Name, node.StatusName, codeNames[InsufficientStorage])
		}
	}
}
