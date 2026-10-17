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

package node

import (
	"errors"
	"reflect"
	"testing"

	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestJoinClusterProvisioner(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	failure := errors.New("bootstrap failure")
	for _, tc := range []struct {
		name        string
		node        config.Node
		generateErr error
		joinErr     error
		labelErr    error
		want        []string
	}{
		{
			name: "Linux worker", node: config.Node{Worker: true},
			want: []string{"controlplane.GenerateToken", "worker.JoinCluster", "controlplane.LabelAndUntaintNode"},
		},
		{
			name: "Linux control plane", node: config.Node{ControlPlane: true, Worker: true},
			want: []string{"controlplane.GenerateToken", "worker.JoinCluster", "controlplane.LabelAndUntaintNode"},
		},
		{
			name: "Windows worker", node: config.Node{Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}},
			want: []string{"controlplane.GenerateJoinConfigWindows", "worker.JoinClusterWindows", "controlplane.LabelAndUntaintNode"},
		},
		{
			name: "Linux token failure", node: config.Node{Worker: true}, generateErr: failure,
			want: []string{"controlplane.GenerateToken"},
		},
		{
			name: "Windows configuration failure", node: config.Node{Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}}, generateErr: failure,
			want: []string{"controlplane.GenerateJoinConfigWindows"},
		},
		{
			name: "Windows join failure is not retried", node: config.Node{Worker: true, Guest: config.Guest{Name: config.GuestOSWindows}}, joinErr: failure,
			want: []string{"controlplane.GenerateJoinConfigWindows", "worker.JoinClusterWindows"},
		},
		{
			name: "Linux label failure", node: config.Node{Worker: true}, labelErr: failure,
			want: []string{"controlplane.GenerateToken", "worker.JoinCluster", "controlplane.LabelAndUntaintNode"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			controlplane := &provisionerBootstrapper{calls: &calls, name: "controlplane", generateErr: tc.generateErr, labelErr: tc.labelErr}
			worker := &provisionerBootstrapper{calls: &calls, name: "worker", joinErr: tc.joinErr}
			starter := Starter{Cfg: &config.ClusterConfig{Name: "test-profile"}, Node: &tc.node}
			err := joinCluster(starter, controlplane, worker, nil)
			if tc.generateErr != nil || tc.joinErr != nil || tc.labelErr != nil {
				if !errors.Is(err, failure) {
					t.Fatalf("error = %v; want bootstrap failure", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(calls, tc.want) {
				t.Fatalf("calls = %v; want %v", calls, tc.want)
			}
		})
	}
}

type provisionerBootstrapper struct {
	bootstrapper.Bootstrapper
	calls       *[]string
	name        string
	generateErr error
	joinErr     error
	labelErr    error
}

func (b *provisionerBootstrapper) GenerateToken(config.ClusterConfig) (string, error) {
	*b.calls = append(*b.calls, b.name+".GenerateToken")
	return "join command", b.generateErr
}

func (b *provisionerBootstrapper) GenerateJoinConfigWindows(config.ClusterConfig, config.Node) (string, error) {
	*b.calls = append(*b.calls, b.name+".GenerateJoinConfigWindows")
	return "join configuration", b.generateErr
}

func (b *provisionerBootstrapper) JoinCluster(config.ClusterConfig, config.Node, string) error {
	*b.calls = append(*b.calls, b.name+".JoinCluster")
	return b.joinErr
}

func (b *provisionerBootstrapper) JoinClusterWindows(*host.Host, string) (string, error) {
	*b.calls = append(*b.calls, b.name+".JoinClusterWindows")
	return "", b.joinErr
}

func (b *provisionerBootstrapper) LabelAndUntaintNode(config.ClusterConfig, config.Node) error {
	*b.calls = append(*b.calls, b.name+".LabelAndUntaintNode")
	return b.labelErr
}
