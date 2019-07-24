// +build integration

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

package integration

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/test/integration/util"
)

func TestStartStop(t *testing.T) {
	p := "TestStartStop" // profile name
	tests := []struct {
		name string
		args []string
	}{
		{"nocache_oldest", []string{
			"--cache-images=false",
			fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion),
			// default is the network created by libvirt, if we change the name minikube won't boot
			// because the given network doesn't exist
			"--kvm-network=default",
			"--kvm-qemu-uri=qemu:///system",
		}},
		{"feature_gates_newest_cni", []string{
			"--feature-gates",
			"ServerSideApply=true",
			"--network-plugin=cni",
			"--extra-config=kubelet.network-plugin=cni",
			"--extra-config=kubeadm.pod-network-cidr=192.168.111.111/16",
			fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion),
		}},
		{"containerd_and_non_default_apiserver_port", []string{
			"--container-runtime=containerd",
			"--docker-opt containerd=/var/run/containerd/containerd.sock",
			"--apiserver-port=8444",
		}},
		{"crio_ignore_preflights", []string{
			"--container-runtime=crio",
			"--extra-config",
			"kubeadm.ignore-preflight-errors=SystemVerification",
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mk := NewMinikubeRunner(t, p+tc.name)
			if !usingNoneDriver(mk) {
				t.Parallel()
			}
			if !strings.Contains(tc.name, "docker") && usingNoneDriver(mk) {
				t.Skipf("skipping %s - incompatible with none driver", tc.name)
			}

			mk.RunCommand("config set WantReportErrorPrompt false", true)
			mk.RunCommand("delete", false)
			mk.CheckStatus(state.None.String())
			mk.Start(tc.args...)
			mk.CheckStatus(state.Running.String())

			ip := mk.RunCommand("ip", true)
			ip = strings.TrimRight(ip, "\n")
			if net.ParseIP(ip) == nil {
				t.Fatalf("IP command returned an invalid address: %s", ip)
			}

			// check for the current-context before and after the stop
			// TODO: medya move this test to its own test so we can do more parallel
			// kr := util.NewKubectlRunner(t, p)
			// currentContext, err := kr.RunCommand([]string{"config", "current-context"}, false)
			// if err != nil {
			// 	t.Fatalf("Failed to fetch current-context")
			// }
			// if strings.TrimRight(string(currentContext), "\n") != p {
			// 	t.Fatalf("got current-context - %q, want  current-context %q", string(currentContext), p)
			// }

			checkStop := func() error {
				mk.RunCommand("stop", true)
				return mk.CheckStatusNoFail(state.Stopped.String())
			}

			if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
				t.Fatalf("timed out while checking stopped status: %v", err)
			}

			// TODO medyagh:  the commented code beollow was not correct ! I leave it for another PR
			// https://github.com/kubernetes/minikube/issues/4854

			// running this command results in error when the current-context is not set
			// if err := mk.Run("config current-context"); err != nil {
			// 	t.Logf("current-context is not set to minikube")
			// }

			mk.Start(tc.args...)
			mk.CheckStatus(state.Running.String())

			mk.RunCommand("delete", true)
			mk.CheckStatus(state.None.String())
		})
	}
}
